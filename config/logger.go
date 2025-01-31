package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

// NewLogger initializes the logger
func NewLogger() (*logrus.Logger, error) {
	log := logrus.New()

	// Define the log file path
	logFilePath := "/app/app/logs/app.log"

	// Ensure the log directory exists
	logDir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDir, 0777); err != nil {
		fmt.Printf("ERROR: Failed to create log directory: %v\n", err)
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open the log file (creating it if it doesn't exist) with O_APPEND flag
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("ERROR: Failed to open log file: %v\n", err)
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	// Set log output to the file
	log.SetOutput(file)

	// Set log format
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.000Z",
		DisableColors:   true,
	})

	// Log initialization message
	log.Info("Logger initialized successfully.")

	return log, nil
}

// InitLogger initializes the logger once
func InitLogger() {
	var err error
	if logger == nil {
		logger, err = NewLogger()
		if err != nil {
			fmt.Printf("ERROR: Failed to initialize logger: %v\n", err)
			os.Exit(1)
		}
	}
}

// GetLogger returns the logger instance
func GetLogger() *logrus.Logger {
	if logger == nil {
		InitLogger()
	}
	return logger
}

// CloseLogger ensures the log file is properly closed
func CloseLogger() {
	// No need to close the file explicitly, as the OS will close it after the process exits
}

func init() {
	InitLogger()
}
