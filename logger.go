// Package logger provides simple logger 
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"sync"
)

var (
	logger       *log.Logger
	multiWriter  io.Writer
	consoleWriter io.Writer
	customWriter io.Writer
	writerMutex  sync.RWMutex
	activeReader *io.PipeWriter
)

func init() {
	consoleWriter = os.Stdout
	multiWriter = consoleWriter
	logger = log.New(multiWriter, "", log.Ldate|log.Ltime|log.Lmsgprefix)
	// TODO: add log rotation
}

// SetProductName updates the prefix
func SetProductName(name string) {
	logger.SetPrefix(fmt.Sprintf("[%v] ", name))
}

// SetOutput sets the output destination for the logger
// This replaces the default console output
func SetOutput(w io.Writer) {
	writerMutex.Lock()
	defer writerMutex.Unlock()
	
	customWriter = w
	updateMultiWriter()
}

// GetReaderCopy returns a copy of the logger output that can be read from
// This allows reading log output while still writing to console
func GetReaderCopy() (io.Reader, error) {
	writerMutex.Lock()
	defer writerMutex.Unlock()
	
	if customWriter == nil {
		return nil, fmt.Errorf("no custom writer set, call SetOutput first")
	}
	
	// 如果已经有活跃的 reader，先关闭它
	if activeReader != nil {
		activeReader.Close()
	}
	
	// 创建新的管道
	reader, writer := io.Pipe()
	activeReader = writer
	
	// 更新多写入器包含管道写入器
	updateMultiWriter()
	
	return reader, nil
}

// RemoveReaderCopy removes the reader copy from logger output
func RemoveReaderCopy() {
	writerMutex.Lock()
	defer writerMutex.Unlock()
	
	if activeReader != nil {
		activeReader.Close()
		activeReader = nil
		updateMultiWriter()
	}
}

// updateMultiWriter 更新多写入器配置
func updateMultiWriter() {
	writers := []io.Writer{consoleWriter}
	
	if customWriter != nil {
		writers = append(writers, customWriter)
	}
	
	if activeReader != nil {
		writers = append(writers, activeReader)
	}
	
	if len(writers) == 1 {
		multiWriter = writers[0]
	} else {
		multiWriter = io.MultiWriter(writers...)
	}
	
	logger.SetOutput(multiWriter)
}

// Info prints log message with INFO level
func Info(format string, args ...any) {
	logger.Printf("[INFO] "+stripNewline(format)+"\n", args...)
}

// Warn prints log message with WARN level
func Warn(format string, args ...any) {
	logger.Printf("[WARN] "+stripNewline(format)+"\n", args...)
}

// Error prints log message with ERROR level
func Error(format string, args ...any) {
	logger.Printf("[ERROR] "+stripNewline(format)+"\n"+string(debug.Stack()), args...)
}

// Fatal prints log message with FATAL level and calls os.Exit(1)
func Fatal(format string, args ...any) {
	logger.Printf("[FATAL] "+stripNewline(format)+"\n"+string(debug.Stack()), args...)
	os.Exit(1)
}

func stripNewline(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}