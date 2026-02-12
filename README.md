# logger
Simple logger with multi-output support

## Features

- Basic logging levels: INFO, WARN, ERROR, FATAL
- Custom product name prefix
- Multi-output support: simultaneous console and custom writer output
- Reader copy functionality: capture log output for processing

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

### Functions

- `SetProductName(name string)` - Set the product name prefix for all log messages
- `SetOutput(w io.Writer)` - Set custom output writer (replaces default console output)
- `GetReaderCopy() (io.Reader, error)` - Get a reader to capture log output copies
- `RemoveReaderCopy()` - Remove the reader copy from log output
- `Info(format string, args ...any)` - Log INFO level message
- `Warn(format string, args ...any)` - Log WARN level message
- `Error(format string, args ...any)` - Log ERROR level message
- `Fatal(format string, args ...any)` - Log FATAL level message and exit

## Thread Safety

The logger is thread-safe and can be used concurrently from multiple goroutines.

## License

MIT