package main

import (
	"appblock/autostart"
	"appblock/blocker"
	"appblock/config"
	"appblock/gemini"
	"appblock/scheduler"
	"appblock/tray"
	"appblock/utils"
	"os"
	"os/signal"
	"syscall"
)

func main() {
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
				block.UpdateConfig(cfg)
				sched.ForceCheck()
			}
		},
		func() {
			// On reload config - update all components
			utils.LogInfo("Reloading all components with new config...")
			cfg = config.Get()
			
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

	// Run system tray (this blocks until quit)
	trayApp.Start()

	utils.LogInfo("APPBlock stopped")
}
