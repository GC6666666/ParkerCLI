package runner

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/parker/ParkerCli/internal/config"
	"github.com/parker/ParkerCli/pkg/logger"
	"github.com/robfig/cron/v3"
)

// JobFunc 定义任务函数类型
type JobFunc func(ctx context.Context) error

// RunnerMode 运行模式
type RunnerMode string

const (
	// ModeDevelopment 开发模式
	ModeDevelopment RunnerMode = "development"
	// ModeProduction 生产模式
	ModeProduction RunnerMode = "production"
	// ModeTest 测试模式
	ModeTest RunnerMode = "test"
)

// ServerOptions Web服务器选项
type ServerOptions struct {
	Port           int               // 端口
	Host           string            // 主机
	ReadTimeout    time.Duration     // 读取超时
	WriteTimeout   time.Duration     // 写入超时
	Mode           RunnerMode        // 运行模式
	EnableCORS     bool              // 是否启用CORS
	TrustedProxies []string          // 受信任代理
	StaticPath     string            // 静态文件路径
	TemplatePath   string            // 模板路径
	Middlewares    []gin.HandlerFunc // 中间件
}

// TaskOptions 定时任务选项
type TaskOptions struct {
	Cron        string        // Cron表达式
	Immediate   bool          // 是否立即执行
	WithSeconds bool          // 是否使用秒级精度
	Concurrent  bool          // 是否允许并发执行
	Timeout     time.Duration // 超时时间
}

// ServerRunner Web服务器运行器接口
type ServerRunner interface {
	// SetupRouter 设置路由
	SetupRouter() *gin.Engine
	// AddRoutes 添加路由
	AddRoutes(router *gin.Engine)
	// Run 运行服务器
	Run(ctx context.Context) error
	// RunWithGracefulShutdown 优雅关闭
	RunWithGracefulShutdown() error
}

// TaskRunner 定时任务运行器接口
type TaskRunner interface {
	// AddTask 添加定时任务
	AddTask(name string, spec string, job JobFunc, opts TaskOptions) (cron.EntryID, error)
	// RemoveTask 移除定时任务
	RemoveTask(id cron.EntryID)
	// Start 启动所有任务
	Start()
	// Stop 停止所有任务
	Stop()
	// RunOnce 执行一次任务
	RunOnce(name string) error
}

// StandardRunner 标准运行器实现
type StandardRunner struct {
	server        *http.Server
	router        *gin.Engine
	cronRunner    *cron.Cron
	tasks         map[string]JobFunc
	taskOpts      map[string]TaskOptions
	serverOpts    ServerOptions
	shutdownWg    sync.WaitGroup
	contextCancel context.CancelFunc
}

// NewStandardRunner 创建标准运行器
func NewStandardRunner(opts ServerOptions) *StandardRunner {
	// 设置Gin模式
	switch opts.Mode {
	case ModeProduction:
		gin.SetMode(gin.ReleaseMode)
	case ModeTest:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	// 创建路由引擎
	router := gin.New()

	// 添加默认中间件
	router.Use(gin.Recovery())

	// 添加自定义中间件
	if len(opts.Middlewares) > 0 {
		router.Use(opts.Middlewares...)
	}

	// 设置受信任代理
	if len(opts.TrustedProxies) > 0 {
		router.SetTrustedProxies(opts.TrustedProxies)
	}

	// 如果静态文件路径存在，加载静态文件
	if opts.StaticPath != "" {
		router.Static("/static", opts.StaticPath)
	}

	// 如果模板路径存在，加载HTML模板
	if opts.TemplatePath != "" {
		router.LoadHTMLGlob(opts.TemplatePath)
	}

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", opts.Host, opts.Port),
		Handler:      router,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
	}

	// 创建任务调度器
	cronRunner := cron.New(cron.WithSeconds())

	return &StandardRunner{
		server:     server,
		router:     router,
		cronRunner: cronRunner,
		tasks:      make(map[string]JobFunc),
		taskOpts:   make(map[string]TaskOptions),
		serverOpts: opts,
	}
}

// SetupRouter 设置路由
func (r *StandardRunner) SetupRouter() *gin.Engine {
	// 根路由，用于健康检查
	r.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API分组
	api := r.router.Group("/api")
	{
		// 版本信息
		api.GET("/version", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"version": config.GetString("version"),
				"app":     config.GetString("app_name"),
			})
		})
	}

	return r.router
}

// AddRoutes 添加自定义路由
func (r *StandardRunner) AddRoutes(router *gin.Engine) {
	// 默认实现不做任何操作
	// 继承者可以重写此方法添加自定义路由
}

// Run 运行服务器
func (r *StandardRunner) Run(ctx context.Context) error {
	// 设置路由
	r.SetupRouter()

	// 添加自定义路由
	r.AddRoutes(r.router)

	// 启动服务器
	logger.Info("启动Web服务器于 %s", r.server.Addr)

	// 监听上下文取消
	go func() {
		<-ctx.Done()
		logger.Info("正在关闭Web服务器...")

		// 创建关闭上下文，设置超时
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := r.server.Shutdown(shutdownCtx); err != nil {
			logger.Error("服务器关闭错误: %v", err)
		}
	}()

	// 启动HTTP服务
	if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("服务器错误: %w", err)
	}

	return nil
}

// RunWithGracefulShutdown 运行服务并支持优雅关闭
func (r *StandardRunner) RunWithGracefulShutdown() error {
	// 创建可取消上下文
	ctx, cancel := context.WithCancel(context.Background())
	r.contextCancel = cancel

	// 启动服务器
	serverErrCh := make(chan error, 1)
	go func() {
		if err := r.Run(ctx); err != nil {
			serverErrCh <- err
		}
		close(serverErrCh)
	}()

	// 捕获关闭信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待关闭信号或错误
	select {
	case err := <-serverErrCh:
		if err != nil {
			return err
		}
	case sig := <-quit:
		logger.Info("接收到关闭信号: %s", sig.String())
	}

	// 取消上下文，触发关闭
	cancel()

	// 等待所有任务完成
	r.shutdownWg.Wait()

	logger.Info("服务器已优雅关闭")
	return nil
}

// AddTask 添加定时任务
func (r *StandardRunner) AddTask(name string, spec string, job JobFunc, opts TaskOptions) (cron.EntryID, error) {
	if _, exists := r.tasks[name]; exists {
		return 0, fmt.Errorf("任务已存在: %s", name)
	}

	// 保存任务函数和选项
	r.tasks[name] = job
	r.taskOpts[name] = opts

	// 创建包装函数
	wrapper := func() {
		// 如果已经有上下文取消函数，说明服务正在关闭
		if r.contextCancel != nil {
			return
		}

		// 增加等待组计数
		r.shutdownWg.Add(1)
		defer r.shutdownWg.Done()

		// 创建任务上下文
		var ctx context.Context
		var cancel context.CancelFunc

		if opts.Timeout > 0 {
			ctx, cancel = context.WithTimeout(context.Background(), opts.Timeout)
		} else {
			ctx, cancel = context.WithCancel(context.Background())
		}
		defer cancel()

		// 执行任务
		logger.Info("执行定时任务: %s", name)
		if err := job(ctx); err != nil {
			logger.Error("任务执行失败 [%s]: %v", name, err)
		} else {
			logger.Info("任务执行成功: %s", name)
		}
	}

	// 添加到cron
	id, err := r.cronRunner.AddFunc(spec, wrapper)
	if err != nil {
		return 0, fmt.Errorf("添加任务失败: %w", err)
	}

	// 如果需要立即执行
	if opts.Immediate {
		go wrapper()
	}

	return id, nil
}

// RemoveTask 移除定时任务
func (r *StandardRunner) RemoveTask(id cron.EntryID) {
	r.cronRunner.Remove(id)
}

// Start 启动所有任务
func (r *StandardRunner) Start() {
	r.cronRunner.Start()
	logger.Info("启动定时任务调度器")
}

// Stop 停止所有任务
func (r *StandardRunner) Stop() {
	r.cronRunner.Stop()
	logger.Info("停止定时任务调度器")
}

// RunOnce 执行一次任务
func (r *StandardRunner) RunOnce(name string) error {
	job, exists := r.tasks[name]
	if !exists {
		return fmt.Errorf("任务不存在: %s", name)
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 执行任务
	logger.Info("手动执行任务: %s", name)
	if err := job(ctx); err != nil {
		logger.Error("任务执行失败 [%s]: %v", name, err)
		return err
	}

	logger.Info("任务执行成功: %s", name)
	return nil
}

// GetDefaultServerOptions 获取默认服务器选项
func GetDefaultServerOptions() ServerOptions {
	// 从配置获取服务器设置
	serverCfg := config.GetAll().Server

	return ServerOptions{
		Port:         serverCfg.Port,
		Host:         serverCfg.Host,
		ReadTimeout:  time.Duration(serverCfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(serverCfg.WriteTimeout) * time.Second,
		Mode:         RunnerMode(config.GetString("environment")),
		EnableCORS:   true,
		StaticPath:   "./static",
		TemplatePath: "./templates/*",
	}
}

// GetDefaultTaskOptions 获取默认任务选项
func GetDefaultTaskOptions() TaskOptions {
	return TaskOptions{
		Cron:        "0 * * * * *", // 每分钟执行一次
		Immediate:   false,
		WithSeconds: true,
		Concurrent:  false,
		Timeout:     30 * time.Second,
	}
}
