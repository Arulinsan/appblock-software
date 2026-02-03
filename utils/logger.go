package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logFile   *os.File
	logger    *log.Logger
	logMutex  sync.Mutex
	logPath   string
)

// InitLogger initializes the logging system
func InitLogger() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	exeDir := filepath.Dir(exePath)
	logPath = filepath.Join(exeDir, "app.log")
	
	// Open log file in append mode
	logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	
	logger = log.New(logFile, "", 0) // No prefix, we'll add our own
	
	LogInfo("Logger initialized")
	return nil
}

// CloseLogger closes the log file
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// LogInfo logs an info message
func LogInfo(format string, args ...interface{}) {
	logMessage("INFO", format, args...)
}

// LogError logs an error message
func LogError(format string, args ...interface{}) {
	logMessage("ERROR", format, args...)
}

// LogWarning logs a warning message
func LogWarning(format string, args ...interface{}) {
	logMessage("WARN", format, args...)
}

// LogBlocked logs when an app is blocked
func LogBlocked(processName string) {
	logMessage("BLOCK", "Terminated process: %s", processName)
}

// logMessage is the internal logging function
func logMessage(level, format string, args ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()
	
	if logger == nil {
		return
	}
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [%s] %s", timestamp, level, message)
	
	logger.Println(logLine)
}

// GetLogPath returns the log file path
func GetLogPath() string {
	return logPath
}
