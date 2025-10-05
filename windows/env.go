package windows

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

/*
SetEnv は HKCU\Volatile Environment に
指定されたキーと値（REG_SZ）を登録します。
値はログオフ／再起動で自動的に破棄されます。

	key   登録する環境変数名
	value 登録する文字列

戻り値:

	error 成功した場合は nil
*/
func SetEnv(key, value string) error {
	// レジストリキーを作成 (存在しない場合は新規作成)
	k, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		`Volatile Environment`,
		registry.SET_VALUE,
	)
	if err != nil {
		return err
	}
	defer k.Close()

	// REG_SZ で値を書き込む
	if err := k.SetStringValue(key, value); err != nil {
		return err
	}

	// WM_SETTINGCHANGE をシステムへ通知しておく
	broadcastEnvChange()

	return nil
}

// 環境変数変更をブロードキャスト (WM_SETTINGCHANGE)
func broadcastEnvChange() {
	user32 := windows.NewLazySystemDLL("user32.dll")
	sendMsg := user32.NewProc("SendMessageTimeoutW")
	if sendMsg.Find() != nil {
		return // DLL が見つからない場合は無視
	}
	const (
		HWND_BROADCAST   = 0xFFFF
		WM_SETTINGCHANGE = 0x001A
		SMTO_ABORTIFHUNG = 0x0002
		timeoutMillis    = 5000
	)
	envPtr, _ := windows.UTF16PtrFromString("Environment")
	sendMsg.Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_SETTINGCHANGE),
		0,
		uintptr(unsafe.Pointer(envPtr)),
		uintptr(SMTO_ABORTIFHUNG),
		uintptr(timeoutMillis),
		0,
	)
}

/*
GetEnv は指定したキー名の環境変数を取得します。
存在しない場合は空文字列を返します。
*/
func GetEnv(key string) string {
	return os.Getenv(key)
}
