# logger

A lightweight Go logger with multi-output support, real-time subscriptions, and thread-safe internals.

## Features

- Log levels: `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`
- Product prefix support via `SetProductName`
- Multi-output writing (console + custom writer)
- Reader mirror stream via `GetReaderCopy`
- Real-time log fan-out to named channels
- Overflow strategy: drop oldest entry when channel buffer is full
- Thread-safe for concurrent goroutines

## Installation

```bash
go get repo.smlk.org/logger
```

## Quick Start

```go
package main

import (
	"os"

	"repo.smlk.org/logger"
)

func main() {
	logger.SetProductName("MyApp")

	file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	defer file.Close()

	// Logs go to stdout and file
	logger.SetOutput(file)

	logger.Debug("debug details: id=%d", 42)
	logger.Info("service started")
	logger.Warn("cache miss rate is high")
	logger.Error("db connection failed: %v", "timeout")
}
```

## Reader Copy

`GetReaderCopy` lets you consume a mirrored stream of log output while normal logging continues.

> Note: You must call `SetOutput(...)` before `GetReaderCopy()`, otherwise it returns an error.

```go
reader, err := logger.GetReaderCopy()
if err != nil {
	panic(err)
}

go func() {
	data, _ := io.ReadAll(reader)
	println(string(data))
}()

logger.Info("captured by reader copy")
logger.RemoveReaderCopy()
```

## Real-time Log Channels

Create named channels to subscribe to structured log entries.

```go
ch := logger.GetLogChannel("monitor")

go func() {
	for entry := range ch {
		fmt.Printf("[%s] %s: %s\n",
			entry.Timestamp.Format("15:04:05"),
			entry.Level,
			entry.Message,
		)
	}
}()

logger.Info("hello channel")
logger.RemoveLogChannel("monitor")
```

### Buffer Behavior

When a channel buffer is full, the logger drops the **oldest** entry and keeps newer logs.

```go
logger.SetChannelBufferSize(3)
ch := logger.GetLogChannel("small")

for i := 0; i < 10; i++ {
	logger.Info("msg %d", i)
}

_ = ch
logger.RemoveLogChannel("small")
```

## API Reference

### Configuration

- `SetProductName(name string)`
- `SetOutput(w io.Writer)`
- `SetChannelBufferSize(size int)`

### Reader Mirror

- `GetReaderCopy() (io.Reader, error)`
- `RemoveReaderCopy()`

### Channel Subscription

- `GetLogChannel(name string) <-chan LogEntry`
- `GetLogChannelWithConfig(name string, config LogChannelConfig) <-chan LogEntry`
- `RemoveLogChannel(name string)`

### Logging

- `Debug(format string, args ...any)`
- `Info(format string, args ...any)`
- `Warn(format string, args ...any)`
- `Error(format string, args ...any)`
- `Fatal(format string, args ...any)`

### Types

```go
type LogEntry struct {
	Timestamp  time.Time
	Level      string
	Message    string
	Prefix     string
	StackTrace []byte
}

type LogChannelConfig struct {
	BufferSize int
	Timeout    time.Duration
}
```

## Behavior Notes

- `Fatal(...)` logs and then exits via `os.Exit(1)`.
- `Error(...)` and `Fatal(...)` print stack traces to output.
- Channel `Timeout` is part of config type but is not currently used in send logic.

## Example Program

See [`examples/main.go`](examples/main.go) for a full end-to-end demonstration.

## License

MIT