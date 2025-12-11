package common

import (
	"log"
	"os"
)

// IsQuietMode checks if quiet mode is enabled via environment variable
func IsQuietMode() bool {
	return os.Getenv("QUIET_BUILD") == "1" || os.Getenv("QUIET_BUILD") == "true"
}

// LogInfo prints an info message unless quiet mode is enabled
func LogInfo(format string, v ...interface{}) {
	if !IsQuietMode() {
		log.Printf(format, v...)
	}
}

// LogDebug prints a debug message unless quiet mode is enabled
func LogDebug(format string, v ...interface{}) {
	if !IsQuietMode() {
		log.Printf(format, v...)
	}
}

// LogWarning always prints a warning message
func LogWarning(format string, v ...interface{}) {
	log.Printf("⚠️  "+format, v...)
}

// LogSuccess always prints a success message
func LogSuccess(format string, v ...interface{}) {
	log.Printf("✅ "+format, v...)
}

// LogError always prints an error message
func LogError(format string, v ...interface{}) {
	log.Printf("❌ "+format, v...)
}
