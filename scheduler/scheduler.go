package scheduler

import (
	"appblock/config"
	"appblock/utils"
	"sync"
	"time"
)

// Scheduler manages productive time checking
type Scheduler struct {
	config         *config.Config
	isProductive   bool
	mu             sync.RWMutex
	ticker         *time.Ticker
	stopChan       chan bool
	statusCallback func(bool)
}

// NewScheduler creates a new scheduler instance
func NewScheduler(cfg *config.Config) *Scheduler {
	return &Scheduler{
		config:       cfg,
		isProductive: false,
		stopChan:     make(chan bool),
	}
}

// Start starts the scheduler loop
func (s *Scheduler) Start() {
	// Check immediately on start
	s.checkProductiveTime()
	
	// Log initial status with current time
	currentTime := time.Now().Format("15:04")
	if s.isProductive {
		utils.LogInfo("Starting in productive time (current: %s) - blocking active", currentTime)
	} else {
		utils.LogInfo("Starting outside productive time (current: %s) - blocking idle", currentTime)
	}
	
	// Create ticker for periodic checks (every 30 seconds)
	s.ticker = time.NewTicker(30 * time.Second)
	
	utils.LogInfo("Scheduler started with 30s check interval")
	
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.checkProductiveTime()
			case <-s.stopChan:
				s.ticker.Stop()
				utils.LogInfo("Scheduler stopped")
				return
			}
		}
	}()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	if s.ticker != nil {
		s.stopChan <- true
	}
}

// checkProductiveTime checks if we're in productive time
func (s *Scheduler) checkProductiveTime() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	wasProductive := s.isProductive
	s.isProductive = s.config.IsProductiveTime()
	
	// Debug: Always log current check result
	currentTime := time.Now().Format("15:04")
	utils.LogInfo("Productive time check at %s: enabled=%v, isProductive=%v", currentTime, s.config.Enabled, s.isProductive)
	
	// Log state changes with timestamp
	if wasProductive != s.isProductive {
		if s.isProductive {
			utils.LogInfo("Entered productive time at %s - blocking now active", currentTime)
		} else {
			utils.LogInfo("Left productive time at %s - blocking now idle", currentTime)
		}
		
		// Notify callback if set
		if s.statusCallback != nil {
			s.statusCallback(s.isProductive)
		}
	}
}

// IsProductive returns whether we're currently in productive time
func (s *Scheduler) IsProductive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isProductive
}

// SetStatusCallback sets a callback for status changes
func (s *Scheduler) SetStatusCallback(callback func(bool)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.statusCallback = callback
}

// UpdateConfig updates the scheduler configuration
func (s *Scheduler) UpdateConfig(cfg *config.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = cfg
	utils.LogInfo("Scheduler config updated")
}

// ForceCheck forces an immediate check of productive time
func (s *Scheduler) ForceCheck() {
	s.checkProductiveTime()
}
