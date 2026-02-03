package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TimeWindow represents a time range for productive hours
type TimeWindow struct {
	Start string `json:"start"` // Format: "HH:MM"
	End   string `json:"end"`   // Format: "HH:MM"
}

// AIConfig represents AI-related settings
type AIConfig struct {
	Enabled     bool   `json:"enabled"`
	Personality string `json:"personality"`
	Model       string `json:"model"`
}

// Config represents the application configuration
type Config struct {
	Enabled               bool         `json:"enabled"`
	Autostart             bool         `json:"autostart"`
	ScanIntervalSeconds   int          `json:"scan_interval_seconds"`
	PopupCooldownSeconds  int          `json:"popup_cooldown_seconds"`
	ActiveDays            []string     `json:"active_days"`
	TimeWindows           []TimeWindow `json:"time_windows"`
	Blocklist             []string     `json:"blocklist"`
	AI                    AIConfig     `json:"ai"`
	FirstRunCompleted     bool         `json:"first_run_completed"`
}

var (
	configPath string
	instance   *Config
)

// Default configuration
func defaultConfig() *Config {
	return &Config{
		Enabled:              false,
		Autostart:            true,
		ScanIntervalSeconds:  5,
		PopupCooldownSeconds: 60,
		ActiveDays:           []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
		TimeWindows: []TimeWindow{
			{Start: "09:00", End: "12:00"},
			{Start: "13:00", End: "17:00"},
			{Start: "19:00", End: "21:00"},
		},
		Blocklist: []string{"chrome.exe", "discord.exe", "telegram.exe"},
		AI: AIConfig{
			Enabled:     true,
			Personality: "",
			Model:       "gemini-3-flash-preview",
		},
		FirstRunCompleted: false,
	}
}

// Init initializes the config system
func Init() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)
	configPath = filepath.Join(exeDir, "config.json")
	
	return Load()
}

// Load loads configuration from file or creates default
func Load() error {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		instance = defaultConfig()
		return Save()
	}

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	instance = &Config{}
	if err := json.Unmarshal(data, instance); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return nil
}

// Save saves current configuration to file
func Save() error {
	if instance == nil {
		return fmt.Errorf("config not initialized")
	}

	data, err := json.MarshalIndent(instance, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Get returns the current configuration
func Get() *Config {
	return instance
}

// GetPath returns the config file path
func GetPath() string {
	return configPath
}

// IsProductiveTime checks if current time is within productive hours
func (c *Config) IsProductiveTime() bool {
	if !c.Enabled {
		return false
	}

	now := time.Now()
	
	// Check if today is an active day
	currentDay := now.Weekday().String()[:3] // Mon, Tue, etc.
	dayActive := false
	for _, day := range c.ActiveDays {
		if day == currentDay {
			dayActive = true
			break
		}
	}
	
	if !dayActive {
		return false
	}

	// Check if current time is in any time window
	currentHour := now.Hour()
	currentMinute := now.Minute()
	currentTimeInMinutes := currentHour*60 + currentMinute
	
	for _, window := range c.TimeWindows {
		startTime, err := time.Parse("15:04", window.Start)
		if err != nil {
			continue
		}
		endTime, err := time.Parse("15:04", window.End)
		if err != nil {
			continue
		}
		
		startMinutes := startTime.Hour()*60 + startTime.Minute()
		endMinutes := endTime.Hour()*60 + endTime.Minute()
		
		// Check if current time is within the window
		if currentTimeInMinutes >= startMinutes && currentTimeInMinutes <= endMinutes {
			return true
		}
	}

	return false
}

// ToggleEnabled toggles the enabled state
func (c *Config) ToggleEnabled() error {
	c.Enabled = !c.Enabled
	return Save()
}

// SetAutostart sets autostart value
func (c *Config) SetAutostart(value bool) error {
	c.Autostart = value
	return Save()
}

// MarkFirstRunCompleted marks the first run as completed
func (c *Config) MarkFirstRunCompleted() error {
	c.FirstRunCompleted = true
	return Save()
}
