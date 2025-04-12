package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

// LogLevel 定义日志级别类型
type LogLevel int

const (
	// 日志级别常量
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger 是日志记录器接口
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
	SetLevel(level LogLevel)
	SetOutput(w io.Writer)
}

// StandardLogger 实现Logger接口
type StandardLogger struct {
	level  LogLevel
	output io.Writer
}

// 创建新的标准日志记录器
func NewStandardLogger() *StandardLogger {
	return &StandardLogger{
		level:  INFO,
		output: os.Stdout,
	}
}

// SetLevel 设置日志级别
func (l *StandardLogger) SetLevel(level LogLevel) {
	l.level = level
}

// SetOutput 设置日志输出目标
func (l *StandardLogger) SetOutput(w io.Writer) {
	l.output = w
}

// log 内部日志记录方法
func (l *StandardLogger) log(level LogLevel, levelStr string, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	// 格式化时间戳
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// 格式化日志消息
	message := fmt.Sprintf(format, args...)

	// 完整日志行
	logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, levelStr, message)

	// 写入日志
	fmt.Fprint(l.output, logLine)

	// 如果是致命错误，退出程序
	if level == FATAL {
		os.Exit(1)
	}
}

// Debug 记录调试级别日志
func (l *StandardLogger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, "DEBUG", format, args...)
}

// Info 记录信息级别日志
func (l *StandardLogger) Info(format string, args ...interface{}) {
	l.log(INFO, "INFO", format, args...)
}

// Warn 记录警告级别日志
func (l *StandardLogger) Warn(format string, args ...interface{}) {
	l.log(WARN, "WARN", format, args...)
}

// Error 记录错误级别日志
func (l *StandardLogger) Error(format string, args ...interface{}) {
	l.log(ERROR, "ERROR", format, args...)
}

// Fatal 记录致命错误日志并退出程序
func (l *StandardLogger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, "FATAL", format, args...)
}

// ParseLevel 将字符串转换为日志级别
func ParseLevel(level string) LogLevel {
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

// 全局日志实例
var defaultLogger = NewStandardLogger()

// GetLogger 返回默认日志记录器
func GetLogger() Logger {
	return defaultLogger
}

// 提供全局方法，方便直接调用
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

func SetOutput(w io.Writer) {
	defaultLogger.SetOutput(w)
}
