package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/parker/ParkerCli/internal/logs"
	"github.com/urfave/cli/v2"
)

var LogsCommand = &cli.Command{
	Name:  "logs",
	Usage: "日志查看与管理",
	Subcommands: []*cli.Command{
		{
			Name:  "tail",
			Usage: "实时追踪日志",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "file", Value: "./logs/app.log", Usage: "日志文件路径"},
				&cli.IntFlag{Name: "lines", Value: 10, Usage: "显示最后几行"},
				&cli.DurationFlag{Name: "interval", Value: time.Millisecond * 500, Usage: "刷新间隔"},
			},
			Action: logsTailAction,
		},
		{
			Name:  "grep",
			Usage: "过滤日志关键字",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "keyword", Usage: "过滤关键字"},
				&cli.StringFlag{Name: "file", Value: "./logs/app.log", Usage: "日志文件路径"},
				&cli.BoolFlag{Name: "ignore-case", Aliases: []string{"i"}, Usage: "忽略大小写"},
				&cli.BoolFlag{Name: "regex", Usage: "使用正则表达式"},
			},
			Action: logsGrepAction,
		},
	},
}

func logsTailAction(c *cli.Context) error {
	logFile := c.String("file")
	lines := c.Int("lines")
	interval := c.Duration("interval")

	// 创建日志管理器
	manager := logs.NewStandardLogsManager()

	fmt.Printf("追踪日志文件: %s (按 Ctrl+C 停止)\n", logFile)

	// 显示最后几行并开始追踪
	logStream, err := manager.TailLogs(logFile, lines, interval)
	if err != nil {
		return err
	}

	// 监听日志流
	for line := range logStream.Lines {
		fmt.Println(line)
	}

	return nil
}

func logsGrepAction(c *cli.Context) error {
	keyword := c.String("keyword")
	logFile := c.String("file")
	ignoreCase := c.Bool("ignore-case")
	useRegex := c.Bool("regex")

	if keyword == "" {
		return fmt.Errorf("必须提供关键字参数")
	}

	// 创建日志管理器
	manager := logs.NewStandardLogsManager()

	fmt.Printf("过滤包含关键字 '%s' 的日志...\n", keyword)

	// 执行过滤
	matches, lineCount, err := manager.FilterLogs(logFile, keyword, ignoreCase, useRegex)
	if err != nil {
		return err
	}

	// 显示匹配结果
	for _, line := range matches {
		fmt.Println(line)
	}

	// 获取文件的绝对路径
	absPath, _ := filepath.Abs(logFile)

	// 显示统计信息
	fmt.Print(logs.FormatFilterResultsSummary(matches, lineCount, keyword, absPath))

	return nil
}
