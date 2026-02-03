package blocker

import (
	"appblock/config"
	"appblock/gemini"
	"appblock/popup"
	"appblock/scheduler"
	"appblock/utils"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// Blocker manages process blocking
type Blocker struct {
	config        *config.Config
	scheduler     *scheduler.Scheduler
	geminiClient  *gemini.Client
	ticker        *time.Ticker
	stopChan      chan bool
	lastPopupTime time.Time
	mu            sync.Mutex
}

// NewBlocker creates a new blocker instance
func NewBlocker(cfg *config.Config, sched *scheduler.Scheduler, geminiClient *gemini.Client) *Blocker {
	return &Blocker{
		config:        cfg,
		scheduler:     sched,
		geminiClient:  geminiClient,
		stopChan:      make(chan bool),
		lastPopupTime: time.Time{}, // Zero time
	}
}

// Start starts the blocker loop
func (b *Blocker) Start() {
	scanInterval := time.Duration(b.config.ScanIntervalSeconds) * time.Second
	b.ticker = time.NewTicker(scanInterval)
	
	utils.LogInfo("Blocker started with scan interval: %d seconds", b.config.ScanIntervalSeconds)
	
	go func() {
		for {
			select {
			case <-b.ticker.C:
				b.scanAndBlock()
			case <-b.stopChan:
				b.ticker.Stop()
				utils.LogInfo("Blocker stopped")
				return
			}
		}
	}()
}

// Stop stops the blocker
func (b *Blocker) Stop() {
	if b.ticker != nil {
		b.stopChan <- true
	}
}

// scanAndBlock scans for blocked processes and terminates them
func (b *Blocker) scanAndBlock() {
	// Only block if we're in productive time
	if !b.scheduler.IsProductive() {
		return
	}

	utils.LogInfo("Scanning for blocked processes...")

	// Get all running processes
	processes, err := process.Processes()
	if err != nil {
		utils.LogError("Failed to get processes: %v", err)
		return
	}

	foundBlocked := false
	// Check each process against blocklist
	for _, proc := range processes {
		name, err := proc.Name()
		if err != nil {
			continue
		}

		// Check if process is in blocklist (case-insensitive)
		if b.isBlocked(name) {
			foundBlocked = true
			b.terminateProcess(proc, name)
		}
	}
	
	if !foundBlocked {
		utils.LogInfo("Scan complete - no blocked apps found")
	}
}

// isBlocked checks if a process name is in the blocklist
func (b *Blocker) isBlocked(processName string) bool {
	processLower := strings.ToLower(processName)
	
	for _, blocked := range b.config.Blocklist {
		if strings.ToLower(blocked) == processLower {
			return true
		}
	}
	
	return false
}

// terminateProcess terminates a process and shows notification
func (b *Blocker) terminateProcess(proc *process.Process, name string) {
	// Try to terminate the process
	err := proc.Terminate()
	if err != nil {
		// If terminate fails, try kill
		err = proc.Kill()
		if err != nil {
			utils.LogError("Failed to kill process %s (PID: %d): %v", name, proc.Pid, err)
			return
		}
	}

	utils.LogBlocked(name)
	
	// Show popup notification with cooldown
	b.showBlockedNotification(name)
}

// showBlockedNotification shows a notification for blocked app
func (b *Blocker) showBlockedNotification(appName string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check cooldown
	cooldown := time.Duration(b.config.PopupCooldownSeconds) * time.Second
	if time.Since(b.lastPopupTime) < cooldown {
		return
	}

	b.lastPopupTime = time.Now()

	// Get AI message in goroutine to not block
	go func() {
		var message string
		
		if b.config.AI.Enabled && b.geminiClient != nil {
			message = b.geminiClient.GetMotivationalMessage(appName)
		} else {
			message = "Tetap fokus! Ini waktu produktif untuk belajar. Matikan distraksi dan kerjakan tugasmu."
		}

		// Show popup (this will block until user closes it, but we're in a goroutine)
		err := popup.ShowBlocked(appName, message)
		if err != nil {
			utils.LogError("Failed to show popup: %v", err)
		}
	}()
}

// UpdateConfig updates the blocker configuration
func (b *Blocker) UpdateConfig(cfg *config.Config) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.config = cfg
	
	// Restart ticker with new interval if changed
	if b.ticker != nil {
		b.ticker.Reset(time.Duration(cfg.ScanIntervalSeconds) * time.Second)
		utils.LogInfo("Blocker scan interval updated to %d seconds", cfg.ScanIntervalSeconds)
	}
}
