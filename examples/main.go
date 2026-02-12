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
	fmt.Println("=== Logger Reader Copy Example ===")
	
	// 设置产品名称
	logger.SetProductName("ExampleApp")
	
	// 创建文件输出
	file, err := os.OpenFile("example.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to create log file: %v\n", err)
		return
	}
	defer file.Close()
	
	// 设置自定义输出（文件）
	logger.SetOutput(file)
	
	// 获取 reader 拷贝用于实时处理
	reader, err := logger.GetReaderCopy()
	if err != nil {
		fmt.Printf("Failed to get reader copy: %v\n", err)
		return
	}
	
	// 启动日志处理器 goroutine
	go processLogs(reader)
	
	// 模拟应用程序运行并生成日志
	fmt.Println("Generating sample logs...")
	
	logger.Info("Application started successfully")
	time.Sleep(100 * time.Millisecond)
	
	logger.Warn("High memory usage detected: 85%")
	time.Sleep(100 * time.Millisecond)
	
	logger.Info("Processing user request: userID=12345")
	time.Sleep(100 * time.Millisecond)
	
	logger.Error("Database connection timeout")
	time.Sleep(100 * time.Millisecond)
	
	logger.Info("Request processed successfully")
	
	// 等待一段时间让所有日志被处理
	time.Sleep(500 * time.Millisecond)
	
	// 移除 reader 拷贝
	logger.RemoveReaderCopy()
	
	// 展示文件内容
	fmt.Println("\n=== Log file content ===")
	showFileContent("example.log")
	
	fmt.Println("\n=== Example completed ===")
}

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
	
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading logs: %v\n", err)
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