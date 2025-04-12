package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// IsWindows 判断当前操作系统是否为Windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsLinux 判断当前操作系统是否为Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsMacOS 判断当前操作系统是否为macOS
func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// EnsureDir 确保目录存在，如不存在则创建
func EnsureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirExists 检查目录是否存在
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// GetExecutablePath 获取可执行文件路径
func GetExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

// GetWorkingDir 获取当前工作目录
func GetWorkingDir() (string, error) {
	return os.Getwd()
}

// GetHomeDir 获取用户主目录
func GetHomeDir() (string, error) {
	return os.UserHomeDir()
}

// JoinPath 拼接路径，处理跨平台差异
func JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}

// FormatDuration 格式化时间间隔为易读格式
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%d小时%d分钟%d秒", h, m, s)
	} else if m > 0 {
		return fmt.Sprintf("%d分钟%d秒", m, s)
	}
	return fmt.Sprintf("%d秒", s)
}

// SplitArgs 分割命令行参数
func SplitArgs(cmd string) []string {
	var args []string
	inQuote := false
	current := ""

	for _, char := range cmd {
		switch {
		case char == '"' || char == '\'':
			inQuote = !inQuote
		case char == ' ' && !inQuote:
			if current != "" {
				args = append(args, current)
				current = ""
			}
		default:
			current += string(char)
		}
	}

	if current != "" {
		args = append(args, current)
	}

	return args
}

// ExecutableName 获取可执行文件名称(跨平台)
func ExecutableName(name string) string {
	if IsWindows() {
		if !strings.HasSuffix(name, ".exe") {
			return name + ".exe"
		}
	}
	return name
}

// NormalizePath 标准化路径(跨平台)
func NormalizePath(path string) string {
	return filepath.Clean(path)
}

// TruncateString 截断字符串到指定长度并添加省略号
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// BytesToHumanReadable 将字节数转换为人类可读格式
func BytesToHumanReadable(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// MaskSensitiveInfo 掩盖敏感信息
func MaskSensitiveInfo(s string) string {
	// 简化的掩盖逻辑，可根据需要扩展
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}
