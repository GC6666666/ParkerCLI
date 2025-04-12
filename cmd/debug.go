package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/parker/ParkerCli/internal/debug"
	"github.com/urfave/cli/v2"
)

var DebugCommand = &cli.Command{
	Name:  "debug",
	Usage: "调试工具集: 查看日志、服务信息、公用测试等",
	Subcommands: []*cli.Command{
		{
			Name:  "logs",
			Usage: "查看或过滤日志，如 --level=ERROR",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "level", Usage: "过滤日志等级 (DEBUG, INFO, WARN, ERROR)"},
			},
			Action: debugLogsAction,
		},
		{
			Name:   "info",
			Usage:  "查看服务器或进程的基本信息",
			Action: debugInfoAction,
		},
		{
			Name:  "test",
			Usage: "模拟请求测试接口, 例如 --api='/health'",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "api", Usage: "请求路径，如 '/test'"},
				&cli.StringFlag{Name: "host", Value: "http://localhost:8080", Usage: "请求主机地址"},
				&cli.StringFlag{Name: "method", Value: "GET", Usage: "请求方法: GET, POST, PUT等"},
				&cli.IntFlag{Name: "timeout", Value: 5, Usage: "请求超时时间(秒)"},
			},
			Action: debugTestAction,
		},
	},
}

func debugLogsAction(c *cli.Context) error {
	// 创建调试器
	debugger := debug.NewStandardDebugger()

	// 获取日志级别
	level := c.String("level")

	// 获取日志
	logs, err := debugger.GetLogs(level)
	if err != nil {
		return fmt.Errorf("获取日志失败: %w", err)
	}

	// 输出结果
	if level == "" {
		fmt.Println("所有日志:")
	} else {
		fmt.Printf("过滤日志 (级别=%s):\n", level)
	}

	// 格式化输出
	fmt.Println(debug.FormatLogOutput(logs))

	return nil
}

func debugInfoAction(c *cli.Context) error {
	// 创建调试器
	debugger := debug.NewStandardDebugger()

	// 获取系统信息
	info, err := debugger.GetSystemInfo()
	if err != nil {
		return fmt.Errorf("获取系统信息失败: %w", err)
	}

	// 格式化输出
	fmt.Println(debug.FormatSystemInfo(info))

	return nil
}

func debugTestAction(c *cli.Context) error {
	// 创建调试器
	debugger := debug.NewStandardDebugger()

	// 获取参数
	apiPath := c.String("api")
	host := c.String("host")
	method := c.String("method")
	timeout := c.Int("timeout")

	if apiPath == "" {
		apiPath = "/"
	}

	// 创建请求上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// 执行请求
	fmt.Printf("测试API: %s %s%s\n", method, host, apiPath)
	resp, err := debugger.TestEndpoint(ctx, host, apiPath, method, timeout)
	if err != nil {
		return fmt.Errorf("测试请求失败: %w", err)
	}

	// 输出响应
	fmt.Println(debug.FormatHTTPResponse(resp))

	return nil
}
