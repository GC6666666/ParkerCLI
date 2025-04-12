package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

var ReleaseCommand = &cli.Command{
	Name:  "release",
	Usage: "版本发布管理",
	Subcommands: []*cli.Command{
		{
			Name:  "build",
			Usage: "生成正式发行版二进制",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "version", Value: "0.1.0", Usage: "版本号"},
				&cli.StringSliceFlag{Name: "os", Value: cli.NewStringSlice("linux", "darwin", "windows"), Usage: "目标系统"},
				&cli.StringSliceFlag{Name: "arch", Value: cli.NewStringSlice("amd64", "arm64"), Usage: "目标架构"},
				&cli.StringFlag{Name: "output", Value: "./dist", Usage: "输出目录"},
				&cli.BoolFlag{Name: "compress", Usage: "是否压缩二进制文件"},
			},
			Action: releaseBuildAction,
		},
		{
			Name:  "publish",
			Usage: "发布到远程仓库",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "tag", Usage: "发布标签"},
				&cli.StringFlag{Name: "repo", Value: "origin", Usage: "远程仓库名称"},
				&cli.StringFlag{Name: "message", Value: "New release", Usage: "发布说明"},
				&cli.BoolFlag{Name: "draft", Usage: "创建草稿"},
				&cli.BoolFlag{Name: "pre-release", Usage: "标记为预发布版本"},
			},
			Action: releasePublishAction,
		},
	},
}

func releaseBuildAction(c *cli.Context) error {
	version := c.String("version")
	targetOSList := c.StringSlice("os")
	targetArchList := c.StringSlice("arch")
	outputDir := c.String("output")
	compress := c.Bool("compress")

	fmt.Printf("构建发行版 v%s\n", version)

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 设置版本信息和构建时间的ldflags
	buildTime := time.Now().Format("2006-01-02T15:04:05")
	ldflags := fmt.Sprintf("-X 'main.Version=%s' -X 'main.BuildTime=%s'", version, buildTime)

	// 对每个OS/Arch组合构建二进制
	for _, goos := range targetOSList {
		for _, goarch := range targetArchList {
			fmt.Printf("构建 %s/%s...\n", goos, goarch)

			// 构建输出文件名
			outputName := fmt.Sprintf("ParkerCli-%s-%s-%s", version, goos, goarch)
			if goos == "windows" {
				outputName += ".exe"
			}
			outputPath := filepath.Join(outputDir, outputName)

			// 构建命令
			cmd := exec.Command("go", "build", "-ldflags", ldflags, "-trimpath", "-o", outputPath)

			// 设置环境变量
			env := append(os.Environ(),
				fmt.Sprintf("GOOS=%s", goos),
				fmt.Sprintf("GOARCH=%s", goarch),
				"CGO_ENABLED=0")
			cmd.Env = env

			// 执行构建
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				fmt.Printf("构建 %s/%s 失败: %v\n", goos, goarch, err)
				continue
			}

			// 如果需要压缩
			if compress {
				fmt.Printf("压缩 %s...\n", outputName)

				// 可以使用如upx等工具压缩二进制
				if runtime.GOOS == "windows" {
					// Windows下可能需要额外的工具
					fmt.Println("警告: 在Windows下跳过压缩")
				} else {
					// 检查是否安装了upx
					_, err := exec.LookPath("upx")
					if err == nil {
						// 使用upx压缩
						upxCmd := exec.Command("upx", "--best", outputPath)
						upxCmd.Stdout = os.Stdout
						upxCmd.Stderr = os.Stderr

						if err := upxCmd.Run(); err != nil {
							fmt.Printf("压缩失败: %v\n", err)
						}
					} else {
						fmt.Println("未找到upx，跳过压缩步骤")
					}
				}
			}

			fmt.Printf("成功构建: %s\n", outputPath)
		}
	}

	fmt.Println("所有平台构建完成，输出目录:", outputDir)
	return nil
}

func releasePublishAction(c *cli.Context) error {
	tag := c.String("tag")
	repo := c.String("repo")
	message := c.String("message")
	isDraft := c.Bool("draft")
	isPreRelease := c.Bool("pre-release")

	if tag == "" {
		return fmt.Errorf("必须提供发布标签")
	}

	fmt.Printf("发布版本 %s 到远程仓库 %s...\n", tag, repo)

	// 检查git是否安装
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("未找到git命令: %w", err)
	}

	// 检查当前git状态
	statusCmd := exec.Command("git", "status", "--porcelain")
	output, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("获取git状态失败: %w", err)
	}

	// 如果有未提交的更改，提示用户
	if len(strings.TrimSpace(string(output))) > 0 {
		fmt.Println("警告: 有未提交的更改，建议先提交所有更改再发布")
		fmt.Println(string(output))

		fmt.Print("是否继续发布? (y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" {
			return fmt.Errorf("发布被用户取消")
		}
	}

	// 创建和推送标签
	fmt.Printf("创建标签 %s...\n", tag)
	tagCmd := exec.Command("git", "tag", "-a", tag, "-m", message)
	tagCmd.Stdout = os.Stdout
	tagCmd.Stderr = os.Stderr

	if err := tagCmd.Run(); err != nil {
		return fmt.Errorf("创建标签失败: %w", err)
	}

	// 推送标签到远程仓库
	fmt.Printf("推送标签到 %s...\n", repo)
	pushCmd := exec.Command("git", "push", repo, tag)
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr

	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("推送标签失败: %w", err)
	}

	// 模拟创建GitHub Release
	// 实际应用中可以使用GitHub API或gh CLI
	fmt.Println("发布信息:")
	fmt.Printf("- 标签: %s\n", tag)
	fmt.Printf("- 说明: %s\n", message)
	fmt.Printf("- 草稿: %v\n", isDraft)
	fmt.Printf("- 预发布: %v\n", isPreRelease)

	fmt.Println("发布成功!")
	fmt.Println("注意: 此命令仅创建并推送Git标签。")
	fmt.Println("如需完整的GitHub Release功能，请使用'gh'命令行工具或访问GitHub网站。")

	return nil
}
