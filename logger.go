// Package logger provides simple logger
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// LogLevel constants
const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
	LevelFatal = "FATAL"
)

var (
	logger        *log.Logger
	multiWriter   io.Writer
	consoleWriter io.Writer
	customWriter  io.Writer
	writerMutex   sync.RWMutex
	activeReader  *io.PipeWriter

	// Channel 相关变量
	logChannels   map[string]chan LogEntry
	channelsMutex sync.RWMutex
	bufferSize    = 100 // 默认缓冲区大小
)

// LogEntry represents a log message entry
type LogEntry struct {
	Timestamp  time.Time
	Level      string
	Message    string
	Prefix     string
	StackTrace []byte
}

// LogChannelConfig 配置日志 channel
type LogChannelConfig struct {
	BufferSize int           // 缓冲区大小
	Timeout    time.Duration // 发送超时时间
}

var productName string

func init() {
	consoleWriter = os.Stdout
	multiWriter = consoleWriter
	logger = log.New(multiWriter, "", log.Ldate|log.Ltime|log.Lmsgprefix)

	// 初始化 channel 映射
	logChannels = make(map[string]chan LogEntry)

	// TODO: add log rotation
}

// SetProductName updates the prefix
func SetProductName(name string) {
	productName = name
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

// SetChannelBufferSize 设置 channel 缓冲区大小
func SetChannelBufferSize(size int) {
	if size <= 0 {
		size = 100 // 默认值
	}
	bufferSize = size
}

// GetLogChannel 创建或获取指定名称的日志 channel
func GetLogChannel(name string) <-chan LogEntry {
	return GetLogChannelWithConfig(name, LogChannelConfig{
		BufferSize: bufferSize,
		Timeout:    100 * time.Millisecond,
	})
}

// GetLogChannelWithConfig 创建或获取带配置的日志 channel
func GetLogChannelWithConfig(name string, config LogChannelConfig) <-chan LogEntry {
	channelsMutex.Lock()
	defer channelsMutex.Unlock()

	// 如果 channel 已存在，返回它
	if ch, exists := logChannels[name]; exists {
		return ch
	}

	// 创建新的 channel
	ch := make(chan LogEntry, config.BufferSize)
	logChannels[name] = ch

	return ch
}

// RemoveLogChannel 移除指定名称的日志 channel
func RemoveLogChannel(name string) {
	channelsMutex.Lock()
	defer channelsMutex.Unlock()

	if ch, exists := logChannels[name]; exists {
		close(ch)
		delete(logChannels, name)
	}
}

// broadcastToChannels 广播日志条目到所有 channel
// 实现丢弃最旧日志的机制
func broadcastToChannels(entry LogEntry) {
	channelsMutex.RLock()
	defer channelsMutex.RUnlock()

	for name, ch := range logChannels {
		select {
		case ch <- entry:
			// 成功发送
		default:
			// 缓冲区满，丢弃最旧的日志条目
			select {
			case <-ch:
				// 成功丢弃最旧条目，现在可以发送新条目
				select {
				case ch <- entry:
					// 成功发送
				default:
					// 极少数情况下仍然无法发送，记录警告
					fmt.Fprintf(os.Stderr, "WARNING: Log channel '%s' still full after dropping oldest entry\n", name)
				}
			default:
				// 无法丢弃最旧条目（可能 channel 已关闭），记录警告
				fmt.Fprintf(os.Stderr, "WARNING: Cannot drop oldest entry from log channel '%s'\n", name)
			}
		}
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

// Debug prints log message with DEBUG level
func Debug(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	entry := LogEntry{
		Timestamp:  time.Now(),
		Level:      LevelDebug,
		Message:    message,
		Prefix:     GetPrefix(),
		StackTrace: debug.Stack(),
	}

	// 广播到所有 channel
	broadcastToChannels(entry)

	logger.Printf("[DEBUG] " + stripNewline(message) + "\n")
}

// Info prints log message with INFO level
func Info(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	entry := LogEntry{
		Timestamp:  time.Now(),
		Level:      LevelInfo,
		Message:    message,
		Prefix:     GetPrefix(),
		StackTrace: []byte{}, // Info 级别不包含堆栈信息
	}

	// 广播到所有 channel
	broadcastToChannels(entry)

	logger.Printf("[INFO] " + stripNewline(message) + "\n")
}

// Warn prints log message with WARN level
func Warn(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	entry := LogEntry{
		Timestamp:  time.Now(),
		Level:      LevelWarn,
		Message:    message,
		Prefix:     GetPrefix(),
		StackTrace: debug.Stack(),
	}

	// 广播到所有 channel
	broadcastToChannels(entry)

	logger.Printf("[WARN] " + stripNewline(message) + "\n")
}

// Error prints log message with ERROR level
func Error(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	entry := LogEntry{
		Timestamp:  time.Now(),
		Level:      LevelError,
		Message:    message,
		Prefix:     GetPrefix(),
		StackTrace: debug.Stack(),
	}

	// 广播到所有 channel
	broadcastToChannels(entry)

	logger.Printf("[ERROR] " + stripNewline(message) + "\n" + string(debug.Stack()))
}

// Fatal prints log message with FATAL level and calls os.Exit(1)
func Fatal(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	entry := LogEntry{
		Timestamp:  time.Now(),
		Level:      LevelFatal,
		Message:    message,
		Prefix:     GetPrefix(),
		StackTrace: debug.Stack(),
	}

	// 广播到所有 channel
	broadcastToChannels(entry)

	logger.Printf("[FATAL] " + stripNewline(message) + "\n" + string(debug.Stack()))
	os.Exit(1)
}

// GetPrefix 获取当前的日志前缀
func GetPrefix() string {
	return productName
}

func stripNewline(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}
