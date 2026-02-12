# logger
Simple logger with multi-output support and real-time subscription

## Features

- Basic logging levels: INFO, WARN, ERROR, FATAL
- Custom product name prefix
- Multi-output support: simultaneous console and custom writer output
- Reader copy functionality: capture log output for processing
- **Real-time log subscription**: Subscribe to log events via channels
- **Buffer overflow handling**: Automatic log dropping when buffers are full

## Installation

```bash
go get github.com/Ink-33/logger
```

## Usage

### Basic Logging

```go
import "github.com/Ink-33/logger"

// Set product name (optional)
logger.SetProductName("MyApp")

// Log messages with different levels
logger.Info("Application started successfully")
logger.Warn("This is a warning message")
logger.Error("An error occurred: %v", err)
logger.Fatal("Critical error, exiting") // This will call os.Exit(1)
```

### Custom Output Writer

```go
// Set a custom writer (e.g., file, network connection)
file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
    log.Fatal(err)
}
defer file.Close()

logger.SetOutput(file)
logger.Info("This will be written to both console and file")
```

### Reader Copy Functionality

```go
// Get a reader copy to capture log output
reader, err := logger.GetReaderCopy()
if err != nil {
    log.Fatal("Failed to get reader copy:", err)
}

// Read log output in another goroutine
go func() {
    data, err := io.ReadAll(reader)
    if err != nil {
        log.Println("Error reading logs:", err)
        return
    }
    // Process captured log data
    processLogs(data)
}()

// Your application continues logging normally
logger.Info("This message will appear in console and be captured by reader")
logger.Warn("Warning message")

// Remove reader copy when done
logger.RemoveReaderCopy()
```

### Real-time Log Subscription (NEW!)

```go
// Get a log channel for real-time subscription
logCh := logger.GetLogChannel("monitor")

// Subscribe to log events in real-time
go func() {
    for entry := range logCh {
        fmt.Printf("[%s] %s: %s\n", 
            entry.Timestamp.Format("15:04:05"),
            entry.Level,
            entry.Message)
    }
}()

// Logs are automatically sent to subscribers
logger.Info("This will be sent to the channel")
logger.Warn("So will this warning")

// Remove channel when done
logger.RemoveLogChannel("monitor")
```

### Configurable Channel Settings

```go
// Set buffer size (default: 100)
logger.SetChannelBufferSize(50)

// Get channel with custom configuration
config := logger.LogChannelConfig{
    BufferSize: 200,
    Timeout:    500 * time.Millisecond,
}
logCh := logger.GetLogChannelWithConfig("custom-channel", config)
```

### Buffer Overflow Handling

When the channel buffer is full, new log entries are automatically dropped:

```go
// Set small buffer to demonstrate overflow
logger.SetChannelBufferSize(5)

// Create channel
logCh := logger.GetLogChannel("small-buffer")

// Rapid logging (some messages will be dropped)
for i := 0; i < 20; i++ {
    logger.Info("Message %d", i)
    time.Sleep(10 * time.Millisecond)
}

// Only ~5 messages will be received due to buffer limit
received := 0
for entry := range logCh {
    fmt.Printf("Received: %s\n", entry.Message)
    received++
}
fmt.Printf("Total received: %d (others were dropped)\n", received)
```

### Multiple Channel Subscriptions

```go
// Create different channels for different purposes
alerts := logger.GetLogChannel("alerts")
debug := logger.GetLogChannel("debug")
audit := logger.GetLogChannel("audit")

// Handle alerts (errors/warnings)
go func() {
    for entry := range alerts {
        if entry.Level == "ERROR" || entry.Level == "WARN" {
            sendAlert(entry.Message)
        }
    }
}()

// Handle debug logs
go func() {
    for entry := range debug {
        writeToDebugFile(entry)
    }
}()

// Handle audit trail
go func() {
    for entry := range audit {
        addToAuditTrail(entry)
    }
}()

// All logs go to all channels simultaneously
logger.Info("Normal operation")
logger.Warn("Warning condition")
logger.Error("Error occurred")
```

### Complete Example

```go
package main

import (
    "fmt"
    "time"
    "github.com/Ink-33/logger"
)

func main() {
    // Set up logger
    logger.SetProductName("MyService")
    
    // Create file output
    file, _ := os.OpenFile("service.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    logger.SetOutput(file)
    
    // Create real-time monitor
    monitor := logger.GetLogChannel("monitor")
    go func() {
        for entry := range monitor {
            fmt.Printf("MONITOR: [%s] %s\n", entry.Level, entry.Message)
        }
    }()
    
    // Application logic
    logger.Info("Service starting")
    time.Sleep(100 * time.Millisecond)
    
    logger.Warn("Resource usage high")
    time.Sleep(100 * time.Millisecond)
    
    logger.Error("Connection failed")
    time.Sleep(100 * time.Millisecond)
    
    logger.Info("Recovery successful")
    
    // Cleanup
    logger.RemoveLogChannel("monitor")
}
```

```go
import "github.com/Ink-33/logger"

// Set product name (optional)
logger.SetProductName("MyApp")

// Log messages with different levels
logger.Info("Application started successfully")
logger.Warn("This is a warning message")
logger.Error("An error occurred: %v", err)
logger.Fatal("Critical error, exiting") // This will call os.Exit(1)
```

### Custom Output Writer

```go
// Set a custom writer (e.g., file, network connection)
file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
    log.Fatal(err)
}
defer file.Close()

logger.SetOutput(file)
logger.Info("This will be written to both console and file")
```

### Reader Copy Functionality

```go
// Get a reader copy to capture log output
reader, err := logger.GetReaderCopy()
if err != nil {
    log.Fatal("Failed to get reader copy:", err)
}

// Read log output in another goroutine
go func() {
    data, err := io.ReadAll(reader)
    if err != nil {
        log.Println("Error reading logs:", err)
        return
    }
    // Process captured log data
    processLogs(data)
}()

// Your application continues logging normally
logger.Info("This message will appear in console and be captured by reader")
logger.Warn("Warning message")

// Remove reader copy when done
logger.RemoveReaderCopy()
```

### Complete Example

```go
package main

import (
    "bytes"
    "fmt"
    "io"
    "log"
    "sync"
    
    "github.com/Ink-33/logger"
)

func main() {
    // Set product name
    logger.SetProductName("MyService")
    
    // Set custom output (file)
    var buf bytes.Buffer
    logger.SetOutput(&buf)
    
    // Get reader copy for real-time processing
    reader, err := logger.GetReaderCopy()
    if err != nil {
        log.Fatal(err)
    }
    
    // Start log processor
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        data, _ := io.ReadAll(reader)
        fmt.Printf("Processed logs: %s\n", string(data))
    }()
    
    // Generate some logs
    logger.Info("Service starting")
    logger.Warn("High memory usage detected")
    logger.Error("Database connection failed")
    
    // Clean up
    logger.RemoveReaderCopy()
    wg.Wait()
    
    // Check buffered output
    fmt.Printf("Buffered logs: %s\n", buf.String())
}
```

## API Reference

### Core Functions

- `SetProductName(name string)` - Set the product name prefix for all log messages
- `SetOutput(w io.Writer)` - Set custom output writer (replaces default console output)
- `SetChannelBufferSize(size int)` - Set the buffer size for log channels (default: 100)

### Reader Copy Functions

- `GetReaderCopy() (io.Reader, error)` - Get a reader to capture log output copies
- `RemoveReaderCopy()` - Remove the reader copy from log output

### Channel Functions

- `GetLogChannel(name string) <-chan LogEntry` - Create/get a log channel with default config
- `GetLogChannelWithConfig(name string, config LogChannelConfig) <-chan LogEntry` - Create/get channel with custom config
- `RemoveLogChannel(name string)` - Remove and close a log channel

### Logging Functions

- `Info(format string, args ...any)` - Log INFO level message
- `Warn(format string, args ...any)` - Log WARN level message
- `Error(format string, args ...any)` - Log ERROR level message
- `Fatal(format string, args ...any)` - Log FATAL level message and exit

### Types

```go
type LogEntry struct {
    Timestamp time.Time
    Level     string    // "INFO", "WARN", "ERROR", "FATAL"
    Message   string
    Prefix    string
}

type LogChannelConfig struct {
    BufferSize int           // Channel buffer size
    Timeout    time.Duration // Send timeout (not currently used)
}
```

## Thread Safety

The logger is thread-safe and can be used concurrently from multiple goroutines.

## License

MIT