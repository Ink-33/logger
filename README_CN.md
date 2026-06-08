# logger

一个轻量级的 Go 日志库，支持多输出、实时订阅和线程安全的内部实现。

## 特性

- 日志级别：`DEBUG`、`INFO`、`WARN`、`ERROR`、`FATAL`
- 支持通过 `SetProductName` 设置产品前缀
- 多输出写入（控制台 + 自定义写入器）
- 通过 `GetReaderCopy` 实现读取器镜像流
- 实时日志扇出到命名通道
- 溢出策略：当通道缓冲区满时丢弃最旧的条目
- 线程安全，支持并发 goroutine

## 安装

```bash
go get repo.smlk.org/logger
```

## 快速开始

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

	// 日志输出到 stdout 和文件
	logger.SetOutput(file)

	logger.Debug("调试信息：id=%d", 42)
	logger.Info("服务已启动")
	logger.Warn("缓存未命中率过高")
	logger.Error("数据库连接失败：%v", "超时")
}
```

## 读取器副本

`GetReaderCopy` 允许你消费日志输出的镜像流，同时正常日志记录继续进行。

> 注意：你必须在调用 `GetReaderCopy()` 之前调用 `SetOutput(...)`，否则会返回错误。

```go
reader, err := logger.GetReaderCopy()
if err != nil {
	panic(err)
}

go func() {
	data, _ := io.ReadAll(reader)
	println(string(data))
}()

logger.Info("被读取器副本捕获")
logger.RemoveReaderCopy()
```

## 实时日志通道

创建命名通道以订阅结构化的日志条目。

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

### 缓冲区行为

当通道缓冲区满时，日志记录器会丢弃**最旧**的条目并保留更新的日志。

```go
logger.SetChannelBufferSize(3)
ch := logger.GetLogChannel("small")

for i := 0; i < 10; i++ {
	logger.Info("消息 %d", i)
}

_ = ch
logger.RemoveLogChannel("small")
```

## API 参考

### 配置

- `SetProductName(name string)` - 设置产品名称
- `SetOutput(w io.Writer)` - 设置输出目标
- `SetChannelBufferSize(size int)` - 设置通道缓冲区大小
- `SetLevel(level string)` - 设置最低输出日志级别
- `GetLevel() string` - 获取当前最低输出日志级别
- `SetDebugStackTraceEnabled(enabled bool)` - 设置 DEBUG 日志是否打印堆栈
- `GetDebugStackTraceEnabled() bool` - 获取 DEBUG 日志堆栈打印开关状态

### 读取器镜像

- `GetReaderCopy() (io.Reader, error)` - 获取读取器副本
- `RemoveReaderCopy()` - 移除读取器副本

### 通道订阅

- `GetLogChannel(name string) <-chan LogEntry` - 获取日志通道
- `GetLogChannelWithConfig(name string, config LogChannelConfig) <-chan LogEntry` - 使用配置获取日志通道
- `RemoveLogChannel(name string)` - 移除日志通道

### 日志记录

- `Debug(format string, args ...any)` - 调试级别日志
- `Info(format string, args ...any)` - 信息级别日志
- `Warn(format string, args ...any)` - 警告级别日志
- `Error(format string, args ...any)` - 错误级别日志
- `Fatal(format string, args ...any)` - 致命级别日志

### 类型

```go
type LogEntry struct {
	Timestamp  time.Time    // 时间戳
	Level      string       // 日志级别
	Message    string       // 日志消息
	Prefix     string       // 前缀
	StackTrace []byte       // 堆栈跟踪
}

type LogChannelConfig struct {
	BufferSize int           // 缓冲区大小
	Timeout    time.Duration // 发送超时时间
}
```

## 行为说明

- `Fatal(...)` 记录日志后通过 `os.Exit(1)` 退出程序。
- `Error(...)` 和 `Fatal(...)` 会在输出中打印堆栈跟踪。
- `Debug(...)` 仅在调用 `SetDebugStackTraceEnabled(true)` 后才会在输出中打印堆栈跟踪。
- 通道的 `Timeout` 是配置类型的一部分，但当前未在发送逻辑中使用。

## 示例程序

查看 [`examples/main.go`](examples/main.go) 获取完整的端到端演示。

## 许可证

MIT
