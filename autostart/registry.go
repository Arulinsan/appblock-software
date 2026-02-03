//go:build windows
// +build windows

package autostart

import (
	"appblock/utils"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const (
	registryPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	appName      = "APPBlock"
)

// Enable adds the application to Windows startup
func Enable() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Get absolute path
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Open registry key
	key, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	// Set value
	err = key.SetStringValue(appName, exePath)
	if err != nil {
		return fmt.Errorf("failed to set registry value: %w", err)
	}

	utils.LogInfo("Autostart enabled: %s", exePath)
	return nil
}

// Disable removes the application from Windows startup
func Disable() error {
	// Open registry key
	key, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	// Delete value
	err = key.DeleteValue(appName)
	if err != nil {
		// If value doesn't exist, that's okay
		if err == registry.ErrNotExist {
			return nil
		}
		return fmt.Errorf("failed to delete registry value: %w", err)
	}

	utils.LogInfo("Autostart disabled")
	return nil
}

// IsEnabled checks if autostart is currently enabled
func IsEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue(appName)
	return err == nil
}

// Sync synchronizes autostart state with config
func Sync(shouldEnable bool) error {
	isCurrentlyEnabled := IsEnabled()

	if shouldEnable && !isCurrentlyEnabled {
		return Enable()
	} else if !shouldEnable && isCurrentlyEnabled {
		return Disable()
	}

	return nil
}
