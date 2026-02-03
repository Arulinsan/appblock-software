package tray

import (
	"appblock/autostart"
	"appblock/config"
	"appblock/gui"
	"appblock/popup"
	"appblock/utils"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getlantern/systray"
)

//go:embed icon.ico
var iconData []byte

// App holds references to application components
type App struct {
	config            *config.Config
	isProductiveTime  bool
	onToggleEnabled   func()
	onReloadConfig    func()
	onQuit            func()
	
	// Menu items
	mStatus           *systray.MenuItem
	mToggle           *systray.MenuItem
	mSettings         *systray.MenuItem
	mReloadConfig     *systray.MenuItem
	mToggleAutostart  *systray.MenuItem
	mQuit             *systray.MenuItem
}

// NewApp creates a new tray app
func NewApp(cfg *config.Config) *App {
	return &App{
		config:           cfg,
		isProductiveTime: false,
	}
}

// SetCallbacks sets the callback functions
func (a *App) SetCallbacks(onToggleEnabled, onReloadConfig, onQuit func()) {
	a.onToggleEnabled = onToggleEnabled
	a.onReloadConfig = onReloadConfig
	a.onQuit = onQuit
}

// UpdateProductiveStatus updates the productive time status
func (a *App) UpdateProductiveStatus(isProductive bool) {
	a.isProductiveTime = isProductive
	a.updateTooltip()
	a.updateStatusText()
}

// Start starts the system tray
func (a *App) Start() {
	systray.Run(a.onReady, a.onExit)
}

// Quit quits the system tray
func (a *App) Quit() {
	systray.Quit()
}

// onReady is called when the systray is ready
func (a *App) onReady() {
	// Set icon (using a simple text icon since we don't have image files)
	systray.SetIcon(getIcon())
	a.updateTooltip()

	// Create menu items
	a.mStatus = systray.AddMenuItem("Status: Checking...", "Current status")
	a.mStatus.Disable()

	systray.AddSeparator()

	a.mToggle = systray.AddMenuItem("Disable Blocking", "Toggle blocking on/off")
	a.mSettings = systray.AddMenuItem("⚙️ Settings", "Open settings window")
	a.mReloadConfig = systray.AddMenuItem("🔄 Reload Config", "Reload configuration from file")
	
	systray.AddSeparator()
	
	a.mToggleAutostart = systray.AddMenuItem("Enable Autostart", "Toggle autostart")
	
	systray.AddSeparator()
	
	a.mQuit = systray.AddMenuItem("Quit", "Exit APPBlock")

	// Update initial state
	a.updateToggleText()
	a.updateAutostartText()
	a.updateStatusText()

	// Handle menu clicks
	go a.handleMenuClicks()
}

// onExit is called when the systray is exiting
func (a *App) onExit() {
	utils.LogInfo("System tray exiting")
}

// handleMenuClicks handles menu item clicks
func (a *App) handleMenuClicks() {
	for {
		select {
		case <-a.mToggle.ClickedCh:
			a.handleToggle()
		case <-a.mSettings.ClickedCh:
			a.handleSettings()
		case <-a.mReloadConfig.ClickedCh:
			a.handleReloadConfig()
		case <-a.mToggleAutostart.ClickedCh:
			a.handleToggleAutostart()
		case <-a.mQuit.ClickedCh:
			a.handleQuit()
			return
		}
	}
}

// handleToggle handles the toggle enable/disable
func (a *App) handleToggle() {
	if err := a.config.ToggleEnabled(); err != nil {
		utils.LogError("Failed to toggle enabled state: %v", err)
		return
	}

	a.updateToggleText()
	a.updateTooltip()

	if a.onToggleEnabled != nil {
		a.onToggleEnabled()
	}

	enabledText := "disabled"
	if a.config.Enabled {
		enabledText = "enabled"
	}
	utils.LogInfo("Blocking %s", enabledText)
}

// handleSettings opens the settings window
func (a *App) handleSettings() {
	utils.LogInfo("Opening settings window...")
	
	go func() {
		err := gui.ShowSettings(func() {
			// On save callback - reload config
			a.handleReloadConfig()
		})
		
		if err != nil {
			utils.LogError("Failed to show settings: %v", err)
		}
	}()
}

// handleReloadConfig reloads configuration from file
func (a *App) handleReloadConfig() {
	utils.LogInfo("Reloading configuration...")
	
	// Reload config from file
	if err := config.Load(); err != nil {
		utils.LogError("Failed to reload config: %v", err)
		go popup.ShowInfo("Reload Failed", fmt.Sprintf("Failed to reload config:\n%v", err))
		return
	}
	
	// Update local reference
	a.config = config.Get()
	
	// Trigger callback to update other components
	if a.onReloadConfig != nil {
		a.onReloadConfig()
	}
	
	// Update UI
	a.updateToggleText()
	a.updateAutostartText()
	a.updateTooltip()
	
	utils.LogInfo("Configuration reloaded successfully")
	go popup.ShowInfo("Config Reloaded ✅", "Configuration has been reloaded!\n\nNew settings are now active.")
}

// handleToggleAutostart toggles autostart
func (a *App) handleToggleAutostart() {
	newValue := !a.config.Autostart
	
	// Update config
	if err := a.config.SetAutostart(newValue); err != nil {
		utils.LogError("Failed to update autostart config: %v", err)
		return
	}

	// Sync with registry
	if err := autostart.Sync(newValue); err != nil {
		utils.LogError("Failed to sync autostart: %v", err)
		return
	}

	a.updateAutostartText()
	utils.LogInfo("Autostart %s", map[bool]string{true: "enabled", false: "disabled"}[newValue])
}

// handleQuit quits the application
func (a *App) handleQuit() {
	if a.onQuit != nil {
		a.onQuit()
	}
	systray.Quit()
}

// updateTooltip updates the system tray tooltip
func (a *App) updateTooltip() {
	status := "OFF"
	if a.config.Enabled {
		status = "ON"
	}

	productive := ""
	if a.isProductiveTime {
		productive = " | Productive Time 📚"
	}

	tooltip := fmt.Sprintf("APPBlock [%s]%s", status, productive)
	systray.SetTooltip(tooltip)
}

// updateToggleText updates the toggle menu item text
func (a *App) updateToggleText() {
	if a.config.Enabled {
		a.mToggle.SetTitle("Disable Blocking")
	} else {
		a.mToggle.SetTitle("Enable Blocking")
	}
}

// updateAutostartText updates the autostart menu item text
func (a *App) updateAutostartText() {
	if a.config.Autostart {
		a.mToggleAutostart.SetTitle("Disable Autostart")
	} else {
		a.mToggleAutostart.SetTitle("Enable Autostart")
	}
}

// updateStatusText updates the status menu item text
func (a *App) updateStatusText() {
	var statusText string
	
	if !a.config.Enabled {
		statusText = "Status: Disabled"
	} else if a.isProductiveTime {
		statusText = "Status: Blocking Active 🔒"
	} else {
		statusText = "Status: Idle (waiting)"
	}

	a.mStatus.SetTitle(statusText)
}

// getIcon returns the system tray icon
// Uses embedded icon from binary
func getIcon() []byte {
	// Use embedded icon data
	if len(iconData) > 0 {
		utils.LogInfo("Using embedded tray icon (%d bytes)", len(iconData))
		return iconData
	}
	
	// Fallback: try to load from icon.ico file next to executable
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		iconPath := filepath.Join(exeDir, "icon.ico")
		
		data, err := os.ReadFile(iconPath)
		if err == nil && len(data) > 0 {
			utils.LogInfo("Loaded tray icon from: %s", iconPath)
			return data
		}
	}
	
	// Last fallback to default icon
	utils.LogInfo("Using default embedded tray icon")
	return getDefaultIcon()
}

// getDefaultIcon returns a simple embedded icon
// This is a minimal valid ICO file (16x16, monochrome)
func getDefaultIcon() []byte {
	return []byte{
		0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x10, 0x10, 0x00, 0x00, 0x01, 0x00,
		0x20, 0x00, 0x68, 0x04, 0x00, 0x00, 0x16, 0x00, 0x00, 0x00, 0x28, 0x00,
		0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x01, 0x00,
		0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00,
	}
}
