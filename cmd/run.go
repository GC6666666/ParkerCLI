package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
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

	// 设置Gin模式
	if isRelease {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	r := gin.Default()

	// 添加基本路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ParkerCli API服务运行中",
			"status":  "running",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// 添加健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	// 创建带有context的HTTP服务器
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 在goroutine中启动服务器
	go func() {
		fmt.Printf("启动主服务中: 监听端口 %s...\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("服务器启动失败: %s\n", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("正在关闭服务器...")

	// 设置5秒超时关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("服务器关闭: %w", err)
	}

	fmt.Println("服务器已成功关闭")
	return nil
}

func runJobAction(c *cli.Context) error {
	isOnce := c.Bool("once")
	schedule := c.String("schedule")
	name := c.String("name")

	// 任务执行逻辑
	jobFunc := func() {
		fmt.Printf("[%s] 执行任务: %s\n", time.Now().Format("2006-01-02 15:04:05"), name)
		// 这里可以添加实际的业务逻辑
		time.Sleep(time.Second) // 模拟任务执行
		fmt.Printf("[%s] 任务完成: %s\n", time.Now().Format("2006-01-02 15:04:05"), name)
	}

	if isOnce {
		fmt.Printf("执行一次性任务 '%s'...\n", name)
		jobFunc()
		return nil
	}

	fmt.Printf("进入循环任务模式，使用计划: '%s'...\n", schedule)

	// 创建一个新的cron调度器
	scheduler := cron.New(cron.WithSeconds())

	// 添加带有指定调度的任务
	_, err := scheduler.AddFunc(schedule, jobFunc)
	if err != nil {
		return fmt.Errorf("添加任务失败: %w", err)
	}

	// 启动cron调度器
	scheduler.Start()
	fmt.Printf("任务 '%s' 已启动，按Ctrl+C停止...\n", name)

	// 等待中断信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	// 停止调度器（等待任务完成）
	stopCtx := scheduler.Stop()
	<-stopCtx.Done()

	fmt.Println("任务调度器已停止")
	return nil
}
