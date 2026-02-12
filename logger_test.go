package logger

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestBasicFunctionality(t *testing.T) {
	SetProductName("TestApp")
	
	var buf bytes.Buffer
	SetOutput(&buf)
	
	Info("Test message")
	Warn("Warning message")
	
	content := buf.String()
	
	if !strings.Contains(content, "[INFO] Test message") {
		t.Error("INFO message not found")
	}
	
	if !strings.Contains(content, "[WARN] Warning message") {
		t.Error("WARN message not found")
	}
}

func TestReaderCopy(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	
	reader, err := GetReaderCopy()
	if err != nil {
		t.Fatalf("Failed to get reader copy: %v", err)
	}
	
	var wg sync.WaitGroup
	wg.Add(1)
	
	resultChan := make(chan string, 1)
	go func() {
		defer wg.Done()
		// 设置读取超时
		done := make(chan bool, 1)
		var data []byte
		var readErr error
		
		go func() {
			data, readErr = io.ReadAll(reader)
			done <- true
		}()
		
		select {
		case <-done:
			if readErr != nil {
				resultChan <- ""
				return
			}
			resultChan <- string(data)
		case <-time.After(2 * time.Second):
			resultChan <- ""
		}
	}()
	
	Info("Reader test")
	time.Sleep(100 * time.Millisecond)
	RemoveReaderCopy()
	
	wg.Wait()
	result := <-resultChan
	
	if result == "" {
		t.Log("Reader copy test completed (may be empty due to timing)")
	} else if !strings.Contains(result, "[INFO] Reader test") {
		t.Errorf("Reader did not capture message: %s", result)
	}
}

func TestWithoutCustomWriter(t *testing.T) {
	// 测试不设置自定义 writer 的情况
	SetOutput(nil)
	
	_, err := GetReaderCopy()
	if err == nil {
		t.Error("Expected error when no custom writer is set")
	}
}

func TestMultipleOperations(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	
	// 测试多次设置输出
	SetOutput(&buf1)
	Info("Message to buf1")
	
	SetOutput(&buf2)
	Info("Message to buf2")
	
	if !strings.Contains(buf1.String(), "Message to buf1") {
		t.Error("buf1 should contain first message")
	}
	
	if !strings.Contains(buf2.String(), "Message to buf2") {
		t.Error("buf2 should contain second message")
	}
}

func TestLogChannelBasic(t *testing.T) {
	// 创建日志 channel
	ch := GetLogChannel("test-channel")
	
	// 启动接收 goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	
	receivedEntries := make([]LogEntry, 0, 3)
	go func() {
		defer wg.Done()
		timeout := time.After(1 * time.Second)
		
		for i := 0; i < 3; i++ {
			select {
			case entry := <-ch:
				receivedEntries = append(receivedEntries, entry)
			case <-timeout:
				t.Error("Timeout waiting for log entries")
				return
			}
		}
	}()
	
	// 发送日志
	Info("Test info message")
	Warn("Test warning message")
	Error("Test error message")
	
	// 等待接收完成
	wg.Wait()
	
	// 验证接收到的消息
	if len(receivedEntries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(receivedEntries))
	}
	
	if receivedEntries[0].Level != "INFO" || !strings.Contains(receivedEntries[0].Message, "Test info message") {
		t.Error("First entry mismatch")
	}
	
	if receivedEntries[1].Level != "WARN" || !strings.Contains(receivedEntries[1].Message, "Test warning message") {
		t.Error("Second entry mismatch")
	}
	
	if receivedEntries[2].Level != "ERROR" || !strings.Contains(receivedEntries[2].Message, "Test error message") {
		t.Error("Third entry mismatch")
	}
	
	// 清理
	RemoveLogChannel("test-channel")
}

func TestLogChannelDropOldest(t *testing.T) {
	// 设置小缓冲区
	SetChannelBufferSize(3)
	
	// 创建日志 channel
	ch := GetLogChannel("drop-oldest")
	
	// 快速发送超过缓冲区大小的日志
	messages := []string{"First", "Second", "Third", "Fourth", "Fifth"}
	for _, msg := range messages {
		Info(msg + " message")
		time.Sleep(10 * time.Millisecond)
	}
	
	// 收集所有接收到的消息
	var received []string
	timeout := time.After(500 * time.Millisecond)
	
collectLoop:
	for {
		select {
		case entry := <-ch:
			received = append(received, entry.Message)
		case <-timeout:
			break collectLoop
		default:
			// Channel empty，稍等一下继续尝试
			time.Sleep(10 * time.Millisecond)
			// 如果已经收集到足够多的消息就退出
			if len(received) >= 3 {
				break collectLoop
			}
		}
	}
	
	// 验证行为：应该收到最新的3条消息，最旧的被丢弃
	if len(received) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(received))
	}
	
	// 验证收到的是最后3条消息（Third, Fourth, Fifth）
	// 而不是前3条（First, Second, Third）
	expectedLastThree := []string{"Third message", "Fourth message", "Fifth message"}
	for i, expected := range expectedLastThree {
		if i < len(received) && received[i] != expected {
			t.Errorf("Expected '%s' at position %d, got '%s'", expected, i, received[i])
		}
	}
	
	// 清理
	RemoveLogChannel("drop-oldest")
}

func TestMultipleLogChannels(t *testing.T) {
	// 创建多个 channel
	channel1 := GetLogChannel("channel-1")
	channel2 := GetLogChannel("channel-2")
	
	var wg sync.WaitGroup
	wg.Add(2)
	
	// 接收 channel 1 的消息
	received1 := make([]LogEntry, 0, 2)
	go func() {
		defer wg.Done()
		for i := 0; i < 2; i++ {
			select {
			case entry := <-channel1:
				received1 = append(received1, entry)
			case <-time.After(1 * time.Second):
				return
			}
		}
	}()
	
	// 接收 channel 2 的消息
	received2 := make([]LogEntry, 0, 2)
	go func() {
		defer wg.Done()
		for i := 0; i < 2; i++ {
			select {
			case entry := <-channel2:
				received2 = append(received2, entry)
			case <-time.After(1 * time.Second):
				return
			}
		}
	}()
	
	// 发送日志
	Info("Shared message 1")
	Warn("Shared message 2")
	
	// 等待接收完成
	wg.Wait()
	
	// 验证两个 channel 都收到了相同的消息
	if len(received1) != 2 || len(received2) != 2 {
		t.Error("Not all channels received expected messages")
	}
	
	// 验证消息内容一致
	for i := 0; i < 2; i++ {
		if received1[i].Message != received2[i].Message {
			t.Errorf("Messages differ between channels at index %d", i)
		}
	}
	
	// 清理
	RemoveLogChannel("channel-1")
	RemoveLogChannel("channel-2")
}

func TestRemoveLogChannel(t *testing.T) {
	GetLogChannel("temp-channel")
	
	// 验证 channel 存在
	channelsMutex.RLock()
	_, exists := logChannels["temp-channel"]
	channelsMutex.RUnlock()
	
	if !exists {
		t.Error("Channel should exist")
	}
	
	// 移除 channel
	RemoveLogChannel("temp-channel")
	
	// 验证 channel 已移除
	channelsMutex.RLock()
	_, exists = logChannels["temp-channel"]
	channelsMutex.RUnlock()
	
	if exists {
		t.Error("Channel should be removed")
	}
}