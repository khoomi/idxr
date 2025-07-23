package util

import (
	"log"
)

// LogError logs an error with context
func LogError(message string, err error) {
	if err != nil {
		log.Printf("ERROR: %s - %v", message, err)
	}
}

// LogInfo logs an informational message
func LogInfo(message string) {
	log.Printf("INFO: %s", message)
}

// LogWarning logs a warning message
func LogWarning(message string) {
	log.Printf("WARNING: %s", message)
}