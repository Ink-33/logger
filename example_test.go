package logger_test

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Ink-33/logger"
)

func TestBasicLogging(t *testing.T) {
	// 测试基本日志功能
	logger.SetProductName("TestApp")
	
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	
	logger.Info("Test info message")
	logger.Warn("Test warn message")
	logger.Error("Test error message")
	
	content := buf.String()
	
	if !strings.Contains(content, "[INFO] Test info message") {
		t.Error("INFO message not found")
	}
	
	if !strings.Contains(content, "[WARN] Test warn message") {
		t.Error("WARN message not found")
	}
	
	if !strings.Contains(content, "[ERROR] Test error message") {
		t.Error("ERROR message not found")
	}
	
	t.Logf("Log content: %s", content)
}

func TestReaderCopyFunctionality(t *testing.T) {
	// 设置自定义输出
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	
	// 获取 reader 拷贝
	reader, err := logger.GetReaderCopy()
	if err != nil {
		t.Fatalf("Failed to get reader copy: %v", err)
	}
	
	// 启动读取 goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	
	readData := make(chan string, 1)
	go func() {
		defer wg.Done()
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Errorf("Failed to read: %v", err)
			readData <- ""
			return
		}
		readData <- string(data)
	}()
	
	// 写入日志
	logger.Info("Reader test message")
	logger.Warn("Reader warning")
	
	// 等待日志处理
	time.Sleep(100 * time.Millisecond)
	
	// 移除 reader 拷贝
	logger.RemoveReaderCopy()
	
	// 等待读取完成
	wg.Wait()
	
	// 验证读取结果
	result := <-readData
	if !strings.Contains(result, "[INFO] Reader test message") {
		t.Errorf("Reader did not capture INFO message")
	}
	
	if !strings.Contains(result, "[WARN] Reader warning") {
		t.Errorf("Reader did not capture WARN message")
	}
	
	t.Logf("Reader captured: %s", result)
	
	// 验证原始缓冲区
	bufContent := buf.String()
	if !strings.Contains(bufContent, "[INFO] Reader test message") {
		t.Errorf("Buffer did not capture INFO message")
	}
	
	t.Logf("Buffer content: %s", bufContent)
}

func TestMultipleReaderCopies(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	
	// 第一次获取 reader
	reader1, err := logger.GetReaderCopy()
	if err != nil {
		t.Fatalf("Failed to get first reader: %v", err)
	}
	
	readData1 := make(chan string, 1)
	var wg1 sync.WaitGroup
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		data, _ := io.ReadAll(reader1)
		readData1 <- string(data)
	}()
	
	// 记录日志
	logger.Info("First reader message")
	
	// 移除第一个 reader
	logger.RemoveReaderCopy()
	wg1.Wait()
	
	// 获取第二个 reader
	reader2, err := logger.GetReaderCopy()
	if err != nil {
		t.Fatalf("Failed to get second reader: %v", err)
	}
	
	readData2 := make(chan string, 1)
	var wg2 sync.WaitGroup
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		data, _ := io.ReadAll(reader2)
		readData2 <- string(data)
	}()
	
	// 记录更多日志
	logger.Warn("Second reader message")
	
	// 清理
	logger.RemoveReaderCopy()
	wg2.Wait()
	
	// 验证两个 reader 分别捕获了不同的内容
	result1 := <-readData1
	result2 := <-readData2
	
	if !strings.Contains(result1, "First reader message") {
		t.Error("First reader missed its message")
	}
	
	if strings.Contains(result1, "Second reader message") {
		t.Error("First reader incorrectly captured second message")
	}
	
	if !strings.Contains(result2, "Second reader message") {
		t.Error("Second reader missed its message")
	}
	
	t.Logf("Reader 1: %s", result1)
	t.Logf("Reader 2: %s", result2)
}

func Example_usage() {
	logger.SetProductName("DemoApp")
	
	// 基本使用
	logger.Info("Application started")
	logger.Warn("Memory usage high")
	logger.Error("Connection timeout")
	
	// Output:
	// [DemoApp] 2024/01/01 12:00:00 [INFO] Application started
	// [DemoApp] 2024/01/01 12:00:00 [WARN] Memory usage high
	// [DemoApp] 2024/01/01 12:00:00 [ERROR] Connection timeout
}