package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/parker/ParkerCli/internal/runner"
	"github.com/urfave/cli/v2"
)

var RunCommand = &cli.Command{
	Name:  "run",
	Usage: "启动主服务或任务",
	Subcommands: []*cli.Command{
		{
			Name:  "server",
			Usage: "启动后端服务",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "port", Value: "8080", Usage: "服务端口"},
				&cli.BoolFlag{Name: "release", Usage: "是否使用生产模式"},
			},
			Action: runServerAction,
		},
		{
			Name:  "job",
			Usage: "执行一次性或循环任务",
			Flags: []cli.Flag{
				&cli.BoolFlag{Name: "once", Usage: "仅执行一次"},
				&cli.StringFlag{Name: "schedule", Value: "*/1 * * * *", Usage: "Cron表达式(每分钟)"},
				&cli.StringFlag{Name: "name", Value: "default", Usage: "任务名称"},
			},
			Action: runJobAction,
		},
	},
}

func runServerAction(c *cli.Context) error {
	port := c.String("port")
	isRelease := c.Bool("release")

	// 将端口字符串转换为整数
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("无效的端口号: %s", port)
	}

	// 设置服务器选项
	opts := runner.GetDefaultServerOptions()
	opts.Port = portNum

	// 设置运行模式
	if isRelease {
		opts.Mode = runner.ModeProduction
	} else {
		opts.Mode = runner.ModeDevelopment
	}

	// 创建标准运行器并启动服务
	r := runner.NewStandardRunner(opts)

	// 使用优雅关闭机制运行服务器
	return r.RunWithGracefulShutdown()
}

func runJobAction(c *cli.Context) error {
	isOnce := c.Bool("once")
	schedule := c.String("schedule")
	name := c.String("name")

	// 创建运行器
	r := runner.NewStandardRunner(runner.GetDefaultServerOptions())

	// 创建任务选项
	opts := runner.GetDefaultTaskOptions()
	opts.WithSeconds = true

	// 创建任务函数
	jobFunc := func(ctx context.Context) error {
		fmt.Printf("[%s] 执行任务: %s\n", time.Now().Format("2006-01-02 15:04:05"), name)
		// 这里可以添加实际的业务逻辑
		time.Sleep(time.Second) // 模拟任务执行
		fmt.Printf("[%s] 任务完成: %s\n", time.Now().Format("2006-01-02 15:04:05"), name)
		return nil
	}

	if isOnce {
		fmt.Printf("执行一次性任务 '%s'...\n", name)
		// 使用runner的RunOnce方法执行一次性任务
		return r.RunOnce(name)
	}

	fmt.Printf("进入循环任务模式，使用计划: '%s'...\n", schedule)

	// 添加任务到调度器
	_, err := r.AddTask(name, schedule, jobFunc, opts)
	if err != nil {
		return fmt.Errorf("添加任务失败: %w", err)
	}

	// 启动所有任务
	r.Start()
	fmt.Printf("任务 '%s' 已启动，按Ctrl+C停止...\n", name)

	// 等待信号处理由RunWithGracefulShutdown内部处理
	select {}
}
