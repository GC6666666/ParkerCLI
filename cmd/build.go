package cmd

import (
	"context"
	"fmt"
	"runtime"

	"github.com/parker/ParkerCli/internal/builder"
	"github.com/parker/ParkerCli/internal/config"
	"github.com/urfave/cli/v2"
)

var BuildCommand = &cli.Command{
	Name:  "build",
	Usage: "构建相关操作，如编译 Go 程序、打 Docker 镜像等",
	Subcommands: []*cli.Command{
		{
			Name:  "code",
			Usage: "编译 Go 程序，生成二进制文件",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Value: "dist", Usage: "输出目录路径"},
				&cli.StringFlag{Name: "name", Value: "", Usage: "输出文件名"},
				&cli.StringFlag{Name: "os", Value: runtime.GOOS, Usage: "目标操作系统 (linux, windows, darwin)"},
				&cli.StringFlag{Name: "arch", Value: runtime.GOARCH, Usage: "目标架构 (amd64, arm64)"},
				&cli.StringFlag{Name: "version", Value: "", Usage: "版本号，会注入到二进制中"},
				&cli.StringFlag{Name: "main", Value: "main.go", Usage: "主文件路径"},
				&cli.StringFlag{Name: "ldflags", Value: "", Usage: "链接标志参数"},
				&cli.StringFlag{Name: "tags", Value: "", Usage: "构建标签"},
				&cli.BoolFlag{Name: "static", Usage: "启用静态链接"},
				&cli.BoolFlag{Name: "debug", Usage: "保留调试信息"},
				&cli.BoolFlag{Name: "clean", Usage: "清理旧文件重新构建"},
			},
			Action: buildCodeAction,
		},
		{
			Name:  "image",
			Usage: "构建 Docker 镜像",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "tag", Value: "", Usage: "镜像标签"},
				&cli.StringFlag{Name: "name", Value: "", Usage: "镜像名称"},
				&cli.StringFlag{Name: "file", Value: "Dockerfile", Usage: "Dockerfile路径"},
				&cli.StringFlag{Name: "version", Value: "", Usage: "版本号"},
				&cli.BoolFlag{Name: "no-cache", Usage: "禁用构建缓存"},
			},
			Action: buildImageAction,
		},
	},
}

func buildCodeAction(c *cli.Context) error {
	// 初始化配置
	if err := config.Init(""); err != nil {
		return fmt.Errorf("初始化配置失败: %w", err)
	}

	// 创建构建器
	b := builder.NewStandardBuilder()

	// 获取默认构建选项
	opts := builder.GetDefaultBuildOptions(builder.TypeBinary)

	// 更新构建选项
	opts.OutputPath = c.String("output")
	if c.String("name") != "" {
		opts.Name = c.String("name")
	}
	opts.GoOS = c.String("os")
	opts.GoArch = c.String("arch")
	if c.String("version") != "" {
		opts.Version = c.String("version")
	}
	opts.MainFile = c.String("main")
	opts.Tags = c.String("tags")
	opts.Debug = c.Bool("debug")
	opts.CleanBuild = c.Bool("clean")

	// 处理ldflags
	ldflags := c.String("ldflags")
	if c.Bool("static") {
		if ldflags != "" {
			ldflags += " "
		}
		ldflags += "-extldflags '-static'"
	}
	opts.LDFlags = ldflags

	// 执行构建
	ctx := context.Background()
	result, err := b.BuildBinary(ctx, opts)
	if err != nil {
		return fmt.Errorf("构建失败: %w", err)
	}

	// 输出构建结果
	fmt.Println(builder.FormatBuildResult(result, builder.TypeBinary))

	return nil
}

func buildImageAction(c *cli.Context) error {
	// 初始化配置
	if err := config.Init(""); err != nil {
		return fmt.Errorf("初始化配置失败: %w", err)
	}

	// 创建构建器
	b := builder.NewStandardBuilder()

	// 获取默认构建选项
	opts := builder.GetDefaultBuildOptions(builder.TypeDocker)

	// 更新构建选项
	opts.Dockerfile = c.String("file")
	if c.String("name") != "" {
		opts.Name = c.String("name")
	}
	if c.String("version") != "" {
		opts.Version = c.String("version")
	}

	// 处理Docker标签
	if c.String("tag") != "" {
		opts.DockerTags = []string{c.String("tag")}
		if opts.DockerTags[0] != "latest" {
			opts.DockerTags = append(opts.DockerTags, "latest")
		}
	}

	opts.CleanBuild = c.Bool("no-cache")

	// 执行构建
	ctx := context.Background()
	result, err := b.BuildDocker(ctx, opts)
	if err != nil {
		return fmt.Errorf("构建Docker镜像失败: %w", err)
	}

	// 输出构建结果
	fmt.Println(builder.FormatBuildResult(result, builder.TypeDocker))

	return nil
}
