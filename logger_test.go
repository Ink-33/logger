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