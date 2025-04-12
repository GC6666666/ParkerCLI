package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

var TestCommand = &cli.Command{
	Name:  "test",
	Usage: "测试相关操作: 单元测试、集成测试",
	Subcommands: []*cli.Command{
		{
			Name:  "unit",
			Usage: "运行单元测试",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "package", Value: "./...", Usage: "测试包路径"},
				&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "显示详细输出"},
				&cli.BoolFlag{Name: "cover", Usage: "生成测试覆盖率"},
				&cli.StringFlag{Name: "coverprofile", Value: "coverage.out", Usage: "覆盖率报告文件"},
			},
			Action: testUnitAction,
		},
		{
			Name:  "integration",
			Usage: "运行集成测试",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "tags", Value: "integration", Usage: "构建标签"},
				&cli.BoolFlag{Name: "setup", Usage: "测试前运行环境设置"},
				&cli.BoolFlag{Name: "cleanup", Usage: "测试后清理环境"},
			},
			Action: testIntegrationAction,
		},
	},
}

func testUnitAction(c *cli.Context) error {
	pkg := c.String("package")
	verbose := c.Bool("verbose")
	cover := c.Bool("cover")
	coverProfile := c.String("coverprofile")

	fmt.Printf("运行单元测试，包路径: %s\n", pkg)

	// 构建测试命令参数
	args := []string{"test"}

	// 添加详细输出
	if verbose {
		args = append(args, "-v")
	}

	// 添加覆盖率
	if cover {
		args = append(args, "-cover")
		args = append(args, fmt.Sprintf("-coverprofile=%s", coverProfile))
	}

	// 添加包路径
	args = append(args, pkg)

	// 创建命令
	cmd := exec.Command("go", args...)

	// 将命令的标准输出和错误输出连接到当前进程
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 设置工作目录为当前目录
	cmd.Dir, _ = os.Getwd()

	// 运行命令
	fmt.Printf("执行: go %s\n", strings.Join(args, " "))
	err := cmd.Run()

	// 如果生成了覆盖率报告，显示摘要
	if err == nil && cover {
		absPath, _ := filepath.Abs(coverProfile)
		fmt.Printf("覆盖率报告生成在: %s\n", absPath)

		// 使用go tool cover生成HTML报告
		htmlFile := strings.TrimSuffix(coverProfile, filepath.Ext(coverProfile)) + ".html"
		coverCmd := exec.Command("go", "tool", "cover", "-html="+coverProfile, "-o", htmlFile)
		coverCmd.Stdout = os.Stdout
		coverCmd.Stderr = os.Stderr

		if err := coverCmd.Run(); err == nil {
			absHtmlPath, _ := filepath.Abs(htmlFile)
			fmt.Printf("HTML覆盖率报告: %s\n", absHtmlPath)
		}
	}

	if err != nil {
		return fmt.Errorf("单元测试失败: %w", err)
	}

	fmt.Println("单元测试完成")
	return nil
}

func testIntegrationAction(c *cli.Context) error {
	tags := c.String("tags")
	setup := c.Bool("setup")
	cleanup := c.Bool("cleanup")

	fmt.Printf("运行集成测试，标签: %s\n", tags)

	// 运行测试前的环境设置
	if setup {
		fmt.Println("设置集成测试环境...")
		// 这里可以启动数据库容器、初始化环境等

		// 示例: 启动测试用的Docker容器
		if err := setupTestEnvironment(); err != nil {
			return fmt.Errorf("设置测试环境失败: %w", err)
		}
	}

	// 延迟执行清理操作
	if cleanup {
		defer func() {
			fmt.Println("清理集成测试环境...")
			// 这里可以清理数据库、删除临时文件等

			// 示例: 停止并删除测试容器
			if err := cleanupTestEnvironment(); err != nil {
				fmt.Printf("清理测试环境时出错: %s\n", err)
			}
		}()
	}

	// 构建测试命令
	args := []string{"test", "-tags", tags, "./..."}

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("执行: go %s\n", strings.Join(args, " "))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("集成测试失败: %w", err)
	}

	fmt.Println("集成测试完成")
	return nil
}

// 设置测试环境
func setupTestEnvironment() error {
	fmt.Println("模拟环境设置: 创建测试数据库、启动Docker容器等")
	// 在实际应用中，你可以在这里使用Docker API启动测试容器
	// 这里仅做演示
	return nil
}

// 清理测试环境
func cleanupTestEnvironment() error {
	fmt.Println("模拟环境清理: 删除测试数据库、停止Docker容器等")
	// 在实际应用中，你可以在这里停止并删除测试容器
	// 这里仅做演示
	return nil
}
