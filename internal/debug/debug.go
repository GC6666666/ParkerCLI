package debug

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/parker/ParkerCli/pkg/httpclient"
	"github.com/parker/ParkerCli/pkg/logger"
)

// LogAnalyzer 日志分析器接口
type LogAnalyzer interface {
	GetLogs(level string) ([]string, error)
	FilterLogs(logs []string, level string) []string
}

// HTTPTester HTTP测试接口
type HTTPTester interface {
	TestEndpoint(ctx context.Context, host, path, method string, timeout int) (*httpclient.Response, error)
}

// InfoProvider 系统信息提供者接口
type InfoProvider interface {
	GetSystemInfo() (map[string]interface{}, error)
	GetProcessInfo(pid int) (map[string]interface{}, error)
}

// StandardDebugger 标准调试器实现
type StandardDebugger struct {
	httpClient *httpclient.HTTPClient
}

// NewStandardDebugger 创建新的标准调试器
func NewStandardDebugger() *StandardDebugger {
	return &StandardDebugger{
		httpClient: httpclient.NewClient(),
	}
}

// GetLogs 获取日志（示例实现）
func (d *StandardDebugger) GetLogs(level string) ([]string, error) {
	// 实际应用中可以从文件或其他来源读取
	// 这里使用模拟数据
	fakeLogs := []string{
		"[INFO] 服务启动成功",
		"[WARN] 内存使用率较高",
		"[ERROR] 数据库连接失败",
		"[INFO] API请求: GET /users",
		"[ERROR] 文件不存在: config.json",
	}

	return d.FilterLogs(fakeLogs, level), nil
}

// FilterLogs 过滤日志
func (d *StandardDebugger) FilterLogs(logs []string, level string) []string {
	if level == "" {
		return logs
	}

	var filtered []string
	upperLevel := strings.ToUpper(level)

	for _, log := range logs {
		if strings.HasPrefix(log, "["+upperLevel+"]") {
			filtered = append(filtered, log)
		}
	}

	return filtered
}

// TestEndpoint 测试API端点
func (d *StandardDebugger) TestEndpoint(ctx context.Context, host, path, method string, timeout int) (*httpclient.Response, error) {
	logger.Info("测试HTTP端点: %s %s", method, host+path)

	// 创建临时客户端或使用现有客户端
	client := httpclient.NewClient(
		httpclient.WithTimeout(time.Duration(timeout)*time.Second),
		httpclient.WithBaseURL(host),
	)

	var response *httpclient.Response
	var err error

	// 根据HTTP方法执行请求
	switch strings.ToUpper(method) {
	case http.MethodGet:
		response, err = client.Get(ctx, path, nil)
	case http.MethodPost:
		response, err = client.Post(ctx, path, nil)
	case http.MethodPut:
		response, err = client.Put(ctx, path, nil)
	case http.MethodDelete:
		response, err = client.Delete(ctx, path)
	default:
		// 自定义请求
		response, err = client.Do(ctx, httpclient.Request{
			Method: method,
			Path:   path,
		})
	}

	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}

	return response, nil
}

// GetSystemInfo 获取系统信息（示例实现）
func (d *StandardDebugger) GetSystemInfo() (map[string]interface{}, error) {
	// 实际应用中可从系统API获取
	info := map[string]interface{}{
		"cpu_usage":      "40%",
		"memory_usage":   "512MB",
		"active_ports":   []int{8080, 3306},
		"os":             "Linux",
		"uptime":         "24h 13m",
		"disk_free":      "10GB",
		"total_memory":   "16GB",
		"go_version":     "1.18.3",
		"goroutines":     12,
		"parker_version": "0.1.0",
	}
	return info, nil
}

// GetProcessInfo 获取进程信息（示例实现）
func (d *StandardDebugger) GetProcessInfo(pid int) (map[string]interface{}, error) {
	// 实际应用中可从系统API获取
	info := map[string]interface{}{
		"pid":        pid,
		"status":     "running",
		"cpu_usage":  "5%",
		"memory":     "128MB",
		"start_time": "2023-05-10 10:30:45",
		"executable": "/usr/bin/myapp",
		"args":       []string{"--config", "config.json"},
		"open_files": 32,
	}
	return info, nil
}

// FormatLogOutput 格式化日志输出
func FormatLogOutput(logs []string) string {
	if len(logs) == 0 {
		return "没有找到匹配的日志"
	}

	return strings.Join(logs, "\n")
}

// FormatSystemInfo 格式化系统信息输出
func FormatSystemInfo(info map[string]interface{}) string {
	var builder strings.Builder

	builder.WriteString("系统信息:\n")
	builder.WriteString(fmt.Sprintf("CPU使用率: %s\n", info["cpu_usage"]))
	builder.WriteString(fmt.Sprintf("内存使用: %s / %s\n", info["memory_usage"], info["total_memory"]))
	builder.WriteString(fmt.Sprintf("磁盘剩余: %s\n", info["disk_free"]))
	builder.WriteString(fmt.Sprintf("操作系统: %s\n", info["os"]))
	builder.WriteString(fmt.Sprintf("运行时间: %s\n", info["uptime"]))

	builder.WriteString("活动端口: ")
	ports, ok := info["active_ports"].([]int)
	if ok {
		for i, port := range ports {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(fmt.Sprintf("%d", port))
		}
	}
	builder.WriteString("\n")

	builder.WriteString(fmt.Sprintf("Go版本: %s\n", info["go_version"]))
	builder.WriteString(fmt.Sprintf("Goroutines: %d\n", info["goroutines"]))
	builder.WriteString(fmt.Sprintf("ParkerCli版本: %s\n", info["parker_version"]))

	return builder.String()
}

// FormatHTTPResponse 格式化HTTP响应输出
func FormatHTTPResponse(resp *httpclient.Response) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("状态码: %d\n", resp.StatusCode))
	builder.WriteString("响应头:\n")

	for key, values := range resp.Headers {
		for _, value := range values {
			builder.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	builder.WriteString("\n响应体:\n")
	builder.WriteString(resp.String())

	return builder.String()
}
