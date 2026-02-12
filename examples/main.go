package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Ink-33/logger"
)

func main() {
	fmt.Println("=== Logger Advanced Features Example ===")
	
	// 设置产品名称
	logger.SetProductName("AdvancedExample")
	
	// 创建文件输出
	file, err := os.OpenFile("advanced_example.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to create log file: %v\n", err)
		return
	}
	defer file.Close()
	
	// 设置自定义输出（文件）
	logger.SetOutput(file)
	
	fmt.Println("\n--- Demo 1: Reader Copy ---")
	demoReaderCopy()
	
	fmt.Println("\n--- Demo 2: Log Channel (Real-time Subscription) ---")
	demoLogChannel()
	
	fmt.Println("\n--- Demo 3: Multiple Channels ---")
	demoMultipleChannels()
	
	fmt.Println("\n--- Demo 4: Drop Oldest Log Behavior ---")
	demoDropOldest()
	
	// 展示文件内容
	fmt.Println("\n=== Log file content ===")
	showFileContent("advanced_example.log")
	
	fmt.Println("\n=== Example completed ===")
}

func demoReaderCopy() {
	// 获取 reader 拷贝用于实时处理
	reader, err := logger.GetReaderCopy()
	if err != nil {
		fmt.Printf("Failed to get reader copy: %v\n", err)
		return
	}
	
	// 启动日志处理器 goroutine
	go processLogs(reader)
	
	// 生成一些日志
	logger.Info("Reader copy demo started")
	time.Sleep(50 * time.Millisecond)
	logger.Warn("This is a warning in reader copy demo")
	time.Sleep(50 * time.Millisecond)
	logger.Info("Reader copy demo completed")
	
	// 等待处理完成
	time.Sleep(200 * time.Millisecond)
	logger.RemoveReaderCopy()
}

func demoLogChannel() {
	// 创建日志 channel
	logCh := logger.GetLogChannel("realtime-monitor")
	
	// 启动实时监控 goroutine
	go monitorLogs("Monitor", logCh)
	
	// 生成日志
	logger.Info("Starting real-time monitoring")
	time.Sleep(50 * time.Millisecond)
	logger.Warn("High CPU usage detected")
	time.Sleep(50 * time.Millisecond)
	logger.Error("Network connection failed")
	time.Sleep(50 * time.Millisecond)
	logger.Info("Recovery completed")
	
	// 等待处理完成
	time.Sleep(200 * time.Millisecond)
	logger.RemoveLogChannel("realtime-monitor")
}

func demoMultipleChannels() {
	// 创建不同类型的通知 channel
	alertCh := logger.GetLogChannel("alerts")
	debugCh := logger.GetLogChannel("debug")
	
	// 启动不同的处理器
	go handleAlerts(alertCh)
	go handleDebug(debugCh)
	
	// 生成各种级别的日志
	logger.Info("System initialization")
	time.Sleep(30 * time.Millisecond)
	logger.Warn("Memory usage at 80%")
	time.Sleep(30 * time.Millisecond)
	logger.Error("Database connection lost")
	time.Sleep(30 * time.Millisecond)
	logger.Info("Automatic recovery initiated")
	time.Sleep(30 * time.Millisecond)
	logger.Warn("Disk space low")
	
	// 等待处理完成
	time.Sleep(300 * time.Millisecond)
	
	// 清理
	logger.RemoveLogChannel("alerts")
	logger.RemoveLogChannel("debug")
}

func demoDropOldest() {
	fmt.Println("=== Drop Oldest Log Behavior Demo ===")
	
	// 设置小缓冲区来演示丢弃最旧行为
	logger.SetChannelBufferSize(3)
	
	dropOldestCh := logger.GetLogChannel("drop-oldest-demo")
	
	// 启动消费者，消费速度比生产慢
	receivedMessages := make([]string, 0, 10)
	go func() {
		for entry := range dropOldestCh {
			receivedMessages = append(receivedMessages, entry.Message)
			fmt.Printf("🔄 Received: %s\n", entry.Message)
			time.Sleep(150 * time.Millisecond) // 消费较慢
		}
	}()
	
	// 快速生产大量日志
	fmt.Println("Generating rapid log stream (buffer size: 3)...")
	messages := []string{
		"Message ONE",
		"Message TWO", 
		"Message THREE",
		"Message FOUR",
		"Message FIVE",
		"Message SIX",
		"Message SEVEN",
	}
	
	for _, msg := range messages {
		logger.Info(msg)
		fmt.Printf("📤 Sent: %s\n", msg)
		time.Sleep(50 * time.Millisecond) // 生产较快
	}
	
	// 等待处理完成
	time.Sleep(1500 * time.Millisecond)
	logger.RemoveLogChannel("drop-oldest-demo")
	
	// 分析结果
	fmt.Printf("\n📊 Analysis Results:\n")
	fmt.Printf("Total sent: %d messages\n", len(messages))
	fmt.Printf("Total received: %d messages\n", len(receivedMessages))
	
	// 验证行为：应该收到最新的几条消息，最旧的被丢弃
	if len(receivedMessages) > 0 {
		fmt.Printf("First received: %s\n", receivedMessages[0])
		fmt.Printf("Last received: %s\n", receivedMessages[len(receivedMessages)-1])
	}
	
	// 解释行为
	fmt.Println("\n💡 Behavior Explanation:")
	fmt.Println("- Buffer size is 3")
	fmt.Println("- Producer sends 7 messages rapidly") 
	fmt.Println("- Consumer processes slowly (150ms per message)")
	fmt.Println("- When buffer fills up, OLDEST messages are dropped")
	fmt.Println("- Only the LATEST messages remain in buffer")
	fmt.Println("- Consumer receives messages in chronological order")
}

// 辅助函数
func processLogs(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	
	fmt.Println("=== Real-time log processing ===")
	for scanner.Scan() {
		line := scanner.Text()
		
		// 根据日志级别进行不同处理
		switch {
		case strings.Contains(line, "[ERROR]"):
			fmt.Printf("🚨 ERROR DETECTED: %s\n", line)
		case strings.Contains(line, "[WARN]"):
			fmt.Printf("⚠️  WARNING: %s\n", line)
		case strings.Contains(line, "[INFO]"):
			fmt.Printf("ℹ️  INFO: %s\n", line)
		default:
			fmt.Printf("📝 LOG: %s\n", line)
		}
	}
}

func monitorLogs(name string, logCh <-chan logger.LogEntry) {
	fmt.Printf("=== %s Started ===\n", name)
	
	for entry := range logCh {
		timestamp := entry.Timestamp.Format("15:04:05")
		fmt.Printf("[%s] 📊 %s [%s]: %s\n", 
			timestamp, name, entry.Level, entry.Message)
	}
	
	fmt.Printf("=== %s Stopped ===\n", name)
}

func handleAlerts(alertCh <-chan logger.LogEntry) {
	fmt.Println("=== Alert Handler Started ===")
	
	for entry := range alertCh {
		if entry.Level == "ERROR" || entry.Level == "WARN" {
			timestamp := entry.Timestamp.Format("15:04:05")
			fmt.Printf("🚨 ALERT [%s]: %s\n", timestamp, entry.Message)
		}
	}
}

func handleDebug(debugCh <-chan logger.LogEntry) {
	fmt.Println("=== Debug Handler Started ===")
	
	for entry := range debugCh {
		timestamp := entry.Timestamp.Format("15:04:05.000")
		fmt.Printf("🔍 DEBUG [%s] %s: %s\n", 
			timestamp, entry.Level, entry.Message)
	}
}

func showFileContent(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	
	lines := bytes.Split(content, []byte("\n"))
	for i, line := range lines {
		if len(line) > 0 {
			fmt.Printf("%d: %s\n", i+1, string(line))
		}
	}
}