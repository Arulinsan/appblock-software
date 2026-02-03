//go:build windows
// +build windows

package popup

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	MB_OK                = 0x00000000
	MB_ICONINFORMATION   = 0x00000040
	MB_ICONWARNING       = 0x00000030
	MB_ICONERROR         = 0x00000010
	MB_TOPMOST           = 0x00040000
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	messageBoxW      = user32.NewProc("MessageBoxW")
)

// Show displays a Windows message box with the given title and message
func Show(title, message string) error {
	titlePtr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return fmt.Errorf("failed to convert title: %w", err)
	}

	messagePtr, err := syscall.UTF16PtrFromString(message)
	if err != nil {
		return fmt.Errorf("failed to convert message: %w", err)
	}

	// Show message box with info icon and topmost flag
	ret, _, _ := messageBoxW.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(messagePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(MB_OK|MB_ICONINFORMATION|MB_TOPMOST),
	)

	if ret == 0 {
		return fmt.Errorf("failed to show message box")
	}

	return nil
}

// ShowBlocked shows a notification for a blocked app with motivational message
func ShowBlocked(appName, aiMessage string) error {
	title := "APPBlock - Waktu Produktif! 📚"
	
	message := fmt.Sprintf("Aplikasi ditutup: %s\n\n%s\n\nTetap fokus dan semangat belajar!", 
		appName, aiMessage)
	
	return Show(title, message)
}

// ShowInfo shows an informational message
func ShowInfo(title, message string) error {
	return Show(title, message)
}

// ShowTestMessage shows a test popup
func ShowTestMessage() error {
	title := "APPBlock - Test Popup"
	message := "Popup berfungsi dengan baik!\n\nJika ini muncul, sistem notifikasi aktif."
	return Show(title, message)
}
