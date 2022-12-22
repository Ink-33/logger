// Package logger provides simple logger 
package logger

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmsgprefix)
	// TODO: add log rotation
}

// SetProductName updates the prefix
func SetProductName(name string) {
	logger.SetPrefix(fmt.Sprintf("[%v] ", name))
}

// Info prints log message with INFO level
func Info(format string, args ...any) {
	logger.Printf("[INFO] "+format+"\n", args...)
}

// Warn prints log message with WARN level
func Warn(format string, args ...any) {
	logger.Printf("[WARN] "+format+"\n", args...)
}

// Error prints log message with ERROR level
func Error(format string, args ...any) {
	logger.Printf("[ERROR] "+format+"\n"+string(debug.Stack()), args...)
}

// Fatal prints log message with FATAL level and calls os.Exit(1)
func Fatal(format string, args ...any) {
	logger.Printf("[FATAL] "+format+"\n"+string(debug.Stack()), args...)
	os.Exit(1)
}
