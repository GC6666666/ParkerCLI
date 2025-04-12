package builder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/parker/ParkerCli/internal/config"
	"github.com/parker/ParkerCli/internal/utils"
	"github.com/parker/ParkerCli/pkg/logger"
)

// BuildType 构建类型
type BuildType string

const (
	// TypeBinary 二进制构建
	TypeBinary BuildType = "binary"
	// TypeDocker Docker镜像构建
	TypeDocker BuildType = "docker"
)

// BuildOptions 构建选项
type BuildOptions struct {
	Type        BuildType // 构建类型
	OutputPath  string    // 输出路径
	Name        string    // 名称
	Version     string    // 版本
	MainFile    string    // 主文件
	GoOS        string    // 目标操作系统
	GoArch      string    // 目标架构
	LDFlags     string    // 链接标志
	Tags        string    // 构建标签
	Dockerfile  string    // Dockerfile路径
	DockerImage string    // Docker镜像名称
	DockerTags  []string  // Docker标签
	Debug       bool      // 调试模式
	Compress    bool      // 是否压缩
	CleanBuild  bool      // 是否清理构建
}

// BuildResult 构建结果
type BuildResult struct {
	Success      bool      // 是否成功
	OutputPath   string    // 输出路径
	BuildTime    time.Time // 构建时间
	Duration     float64   // 构建耗时(秒)
	Size         int64     // 文件大小
	ImageID      string    // Docker镜像ID
	ImageSize    int64     // Docker镜像大小
	ErrorMessage string    // 错误信息
}

// Builder 构建器接口
type Builder interface {
	BuildBinary(ctx context.Context, opts BuildOptions) (*BuildResult, error)
	BuildDocker(ctx context.Context, opts BuildOptions) (*BuildResult, error)
}

// StandardBuilder 标准构建器实现
type StandardBuilder struct{}

// NewStandardBuilder 创建标准构建器
func NewStandardBuilder() *StandardBuilder {
	return &StandardBuilder{}
}

// BuildBinary 构建Go二进制
func (b *StandardBuilder) BuildBinary(ctx context.Context, opts BuildOptions) (*BuildResult, error) {
	startTime := time.Now()
	result := &BuildResult{
		BuildTime: startTime,
	}

	// 如果未指定输出目录则使用默认目录
	if opts.OutputPath == "" {
		opts.OutputPath = "dist"
	}

	// 确保输出目录存在
	if err := utils.EnsureDir(opts.OutputPath); err != nil {
		return result, fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 如果没有指定名称，使用当前目录名
	if opts.Name == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return result, fmt.Errorf("获取当前目录失败: %w", err)
		}
		opts.Name = filepath.Base(cwd)
	}

	// 确定可执行文件名称（Windows添加.exe后缀）
	exeName := utils.ExecutableName(opts.Name)
	outputFile := filepath.Join(opts.OutputPath, exeName)

	// 如果启用了清理构建，删除现有文件
	if opts.CleanBuild && utils.FileExists(outputFile) {
		if err := os.Remove(outputFile); err != nil {
			logger.Warn("清理旧构建文件失败: %v", err)
		}
	}

	// 构建命令
	args := []string{"build"}

	// 添加构建标签
	if opts.Tags != "" {
		args = append(args, "-tags", opts.Tags)
	}

	// 添加LDFlags（如版本信息）
	ldflags := opts.LDFlags
	if opts.Version != "" && !strings.Contains(ldflags, "main.Version") {
		if ldflags != "" {
			ldflags += " "
		}
		ldflags += fmt.Sprintf("-X 'main.Version=%s'", opts.Version)
	}

	if ldflags != "" {
		args = append(args, "-ldflags", ldflags)
	}

	// 设置输出文件
	args = append(args, "-o", outputFile)

	// 设置主文件路径
	if opts.MainFile != "" {
		args = append(args, opts.MainFile)
	}

	// 配置环境变量
	env := os.Environ()
	if opts.GoOS != "" {
		env = append(env, "GOOS="+opts.GoOS)
	}
	if opts.GoArch != "" {
		env = append(env, "GOARCH="+opts.GoArch)
	}

	// 如果是调试模式，不剔除调试信息
	if opts.Debug {
		env = append(env, "GODEBUG=gctrace=1")
	} else if !strings.Contains(ldflags, "-s") && !strings.Contains(ldflags, "-w") {
		args[2] += " -s -w" // 剔除调试信息减小体积
	}

	// 创建命令
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 执行构建
	logger.Info("开始构建二进制: %s", outputFile)
	logger.Debug("构建命令: go %s", strings.Join(args, " "))

	if err := cmd.Run(); err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		return result, fmt.Errorf("构建失败: %w", err)
	}

	// 检查输出文件是否存在
	if !utils.FileExists(outputFile) {
		result.Success = false
		result.ErrorMessage = "构建成功但输出文件未找到"
		return result, fmt.Errorf("构建成功但输出文件未找到: %s", outputFile)
	}

	// 获取文件信息
	fileInfo, err := os.Stat(outputFile)
	if err != nil {
		logger.Warn("获取输出文件信息失败: %v", err)
	} else {
		result.Size = fileInfo.Size()
	}

	// 如果需要压缩，可以在此添加压缩逻辑

	// 计算构建时间
	duration := time.Since(startTime)
	result.Duration = duration.Seconds()
	result.Success = true
	result.OutputPath = outputFile

	logger.Info("构建成功: %s (大小: %s, 耗时: %.2f秒)",
		outputFile, utils.BytesToHumanReadable(result.Size), result.Duration)

	return result, nil
}

// BuildDocker 构建Docker镜像
func (b *StandardBuilder) BuildDocker(ctx context.Context, opts BuildOptions) (*BuildResult, error) {
	startTime := time.Now()
	result := &BuildResult{
		BuildTime: startTime,
	}

	// 检查docker命令是否可用
	if err := exec.Command("docker", "--version").Run(); err != nil {
		return result, fmt.Errorf("docker命令不可用，请确保已安装Docker: %w", err)
	}

	// 设置默认Dockerfile路径
	if opts.Dockerfile == "" {
		opts.Dockerfile = "Dockerfile"
	}

	// 确保Dockerfile存在
	if !utils.FileExists(opts.Dockerfile) {
		return result, fmt.Errorf("Dockerfile不存在: %s", opts.Dockerfile)
	}

	// 获取镜像名称
	imageName := opts.DockerImage
	if imageName == "" {
		// 使用配置的namespace和应用名称
		cfg := config.GetAll()
		namespace := cfg.Docker.Namespace
		if namespace == "" {
			namespace = "myapp"
		}
		imageName = fmt.Sprintf("%s/%s", namespace, opts.Name)
	}

	// 添加标签
	tags := opts.DockerTags
	if len(tags) == 0 {
		if opts.Version != "" {
			tags = append(tags, opts.Version)
		}
		tags = append(tags, "latest")
	}

	// 构建Docker命令
	args := []string{"build"}

	// 添加标签
	for _, tag := range tags {
		args = append(args, "-t", fmt.Sprintf("%s:%s", imageName, tag))
	}

	// 添加构建参数
	if opts.Version != "" {
		args = append(args, "--build-arg", fmt.Sprintf("VERSION=%s", opts.Version))
	}

	// 是否启用缓存
	if opts.CleanBuild {
		args = append(args, "--no-cache")
	}

	// 指定Dockerfile路径
	args = append(args, "-f", opts.Dockerfile)

	// 当前目录作为构建上下文
	args = append(args, ".")

	// 创建Docker构建命令
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 执行构建
	logger.Info("开始构建Docker镜像: %s", imageName)
	logger.Debug("构建命令: docker %s", strings.Join(args, " "))

	if err := cmd.Run(); err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		return result, fmt.Errorf("Docker构建失败: %w", err)
	}

	// 获取镜像ID
	imageIDCmd := exec.Command("docker", "images", "-q", fmt.Sprintf("%s:%s", imageName, tags[0]))
	imageIDBytes, err := imageIDCmd.Output()
	if err != nil {
		logger.Warn("获取镜像ID失败: %v", err)
	} else {
		result.ImageID = strings.TrimSpace(string(imageIDBytes))
	}

	// 获取镜像大小
	if result.ImageID != "" {
		inspectCmd := exec.Command("docker", "image", "inspect", "-f", "{{.Size}}", result.ImageID)
		sizeBytes, err := inspectCmd.Output()
		if err != nil {
			logger.Warn("获取镜像大小失败: %v", err)
		} else {
			size, err := strconv.ParseInt(strings.TrimSpace(string(sizeBytes)), 10, 64)
			if err == nil {
				result.ImageSize = size
			}
		}
	}

	// 计算构建时间
	duration := time.Since(startTime)
	result.Duration = duration.Seconds()
	result.Success = true
	result.OutputPath = imageName

	// 打印构建结果
	var tagsStr string
	for i, tag := range tags {
		if i > 0 {
			tagsStr += ", "
		}
		tagsStr += tag
	}

	logger.Info("Docker构建成功: %s (%s) (大小: %s, 耗时: %.2f秒)",
		imageName, tagsStr, utils.BytesToHumanReadable(result.ImageSize), result.Duration)

	return result, nil
}

// GetDefaultBuildOptions 获取默认构建选项
func GetDefaultBuildOptions(buildType BuildType) BuildOptions {
	opts := BuildOptions{
		Type:       buildType,
		OutputPath: "dist",
		MainFile:   "main.go",
		GoOS:       runtime.GOOS,
		GoArch:     runtime.GOARCH,
		CleanBuild: false,
		Compress:   false,
		Debug:      false,
	}

	// 从配置获取应用名称和版本
	cfg := config.GetAll()
	opts.Name = cfg.AppName
	opts.Version = cfg.Version

	// 如果是Docker构建
	if buildType == TypeDocker {
		opts.Dockerfile = "Dockerfile"
		if cfg.Docker.Registry != "" && cfg.Docker.Namespace != "" {
			registry := cfg.Docker.Registry
			namespace := cfg.Docker.Namespace
			opts.DockerImage = fmt.Sprintf("%s/%s/%s", registry, namespace, opts.Name)
		} else if cfg.Docker.Namespace != "" {
			opts.DockerImage = fmt.Sprintf("%s/%s", cfg.Docker.Namespace, opts.Name)
		}
		opts.DockerTags = []string{opts.Version, "latest"}
	}

	return opts
}

// FormatBuildResult 格式化构建结果输出
func FormatBuildResult(result *BuildResult, buildType BuildType) string {
	var builder strings.Builder

	if !result.Success {
		builder.WriteString(fmt.Sprintf("构建失败: %s\n", result.ErrorMessage))
		return builder.String()
	}

	builder.WriteString("构建结果:\n")

	if buildType == TypeBinary {
		builder.WriteString(fmt.Sprintf("输出文件: %s\n", result.OutputPath))
		builder.WriteString(fmt.Sprintf("文件大小: %s\n", utils.BytesToHumanReadable(result.Size)))
	} else {
		builder.WriteString(fmt.Sprintf("镜像名称: %s\n", result.OutputPath))
		builder.WriteString(fmt.Sprintf("镜像ID: %s\n", result.ImageID))
		builder.WriteString(fmt.Sprintf("镜像大小: %s\n", utils.BytesToHumanReadable(result.ImageSize)))
	}

	builder.WriteString(fmt.Sprintf("构建时间: %s\n", result.BuildTime.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("构建耗时: %.2f秒\n", result.Duration))

	return builder.String()
}
