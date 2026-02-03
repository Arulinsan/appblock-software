package main

import (
	"appblock/autostart"
	"appblock/blocker"
	"appblock/config"
	"appblock/gemini"
	"appblock/popup"
	"appblock/scheduler"
	"appblock/tray"
	"appblock/utils"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	// Check for single instance (prevent multiple instances)
	if err := checkSingleInstance(); err != nil {
		// Show notification that app is already running
		popup.ShowInfo("APPBlock Sudah Berjalan! ✅", 
			"APPBlock sudah aktif di system tray (pojok kanan bawah).\n\n"+
			"Right-click icon APPBlock untuk:\n"+
			"• Buka Settings\n"+
			"• Toggle Blocking\n"+
			"• Lihat Status")
		return
	}
	defer releaseSingleInstance()

	// Initialize logger
	if err := utils.InitLogger(); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer utils.CloseLogger()

	utils.LogInfo("APPBlock starting...")

	// Load configuration
	if err := config.Init(); err != nil {
		utils.LogError("Failed to initialize config: %v", err)
		panic("Failed to initialize config: " + err.Error())
	}

	cfg := config.Get()
	utils.LogInfo("Configuration loaded successfully")

	// Sync autostart with config
	if err := autostart.Sync(cfg.Autostart); err != nil {
		utils.LogWarning("Failed to sync autostart: %v", err)
	}

	// Initialize Gemini client
	var geminiClient *gemini.Client
	if cfg.AI.Enabled {
		client, err := gemini.NewClient(cfg.AI.Model, cfg.AI.Personality)
		if err != nil {
			utils.LogWarning("Failed to initialize Gemini client: %v", err)
			utils.LogWarning("AI features will be disabled")
		} else {
			geminiClient = client
			utils.LogInfo("Gemini AI client initialized")
		}
	}

	// Create scheduler
	sched := scheduler.NewScheduler(cfg)
	
	// Create blocker
	block := blocker.NewBlocker(cfg, sched, geminiClient)

	// Start scheduler
	sched.Start()
	defer sched.Stop()

	// Start blocker
	block.Start()
	defer block.Stop()

	// Create system tray app
	trayApp := tray.NewApp(cfg)

	// Set up scheduler callback to update tray status
	sched.SetStatusCallback(func(isProductive bool) {
		trayApp.UpdateProductiveStatus(isProductive)
	})

	// Set up tray callbacks
	trayApp.SetCallbacks(
		func() {
			// On toggle enabled - reload config
			if err := config.Load(); err != nil {
				utils.LogError("Failed to reload config: %v", err)
			} else {
				cfg = config.Get()
				sched.UpdateConfig(cfg)
				block.UpdateConfig(cfg)
				sched.ForceCheck()
			}
		},
		func() {
			// On reload config - update all components
			utils.LogInfo("Reloading all components with new config...")
			cfg = config.Get()
			
			// Update scheduler with new config
			sched.UpdateConfig(cfg)
			
			// Update blocker with new config
			block.UpdateConfig(cfg)
			
			// Force scheduler to recheck productive time
			sched.ForceCheck()
			
			utils.LogInfo("All components updated with new configuration")
		},
		func() {
			// On quit
			utils.LogInfo("APPBlock shutting down...")
		},
	)

	// Handle system signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		utils.LogInfo("Received shutdown signal")
		trayApp.Quit()
	}()

	utils.LogInfo("APPBlock started successfully - Running in system tray")

	// Always auto-open settings on startup
	trayApp.SetOpenSettingsOnReady(true)
	
	// Check if this is first run for welcome message
	isFirstRun := !cfg.FirstRunCompleted
	utils.LogInfo("First run check: FirstRunCompleted=%v, isFirstRun=%v", cfg.FirstRunCompleted, isFirstRun)
	
	if isFirstRun {
		// Show welcome notification for first time users
		go popup.ShowInfo("Selamat Datang di APPBlock! 🚀",
			"APPBlock sekarang aktif di system tray.\n\n"+
			"Setup cepat:\n"+
			"1. Atur aplikasi yang ingin diblokir\n"+
			"2. Pilih jam produktif\n"+
			"3. Pilih AI personality\n\n"+
			"Window Settings akan terbuka otomatis...")
	} else {
		// Regular startup - show simple notification
		go popup.ShowInfo("APPBlock Aktif ✅",
			"APPBlock berjalan di system tray.\n\n"+
			"Right-click icon untuk mengatur.")
	}

	// Run system tray (this blocks until quit)
	trayApp.Start()

	utils.LogInfo("APPBlock stopped")
}

// Single instance management using lock file
var lockFile *os.File

func checkSingleInstance() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	exeDir := filepath.Dir(exePath)
	lockPath := filepath.Join(exeDir, "appblock.lock")
	
	// Try to create lock file
	lockFile, err = os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		if os.IsExist(err) {
			// Check if process still running, if not delete stale lock
			os.Remove(lockPath)
			// Try again
			lockFile, err = os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if err != nil {
				return fmt.Errorf("another instance is already running")
			}
			return nil
		}
		return fmt.Errorf("failed to create lock file: %w", err)
	}
	
	// Write PID to lock file
	fmt.Fprintf(lockFile, "%d", os.Getpid())
	return nil
}

func releaseSingleInstance() {
	if lockFile != nil {
		lockFile.Close()
		
		exePath, err := os.Executable()
		if err == nil {
			exeDir := filepath.Dir(exePath)
			lockPath := filepath.Join(exeDir, "appblock.lock")
			os.Remove(lockPath)
		}
	}
}
