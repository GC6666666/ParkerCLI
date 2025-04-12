package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/parker/ParkerCli/internal/debug"
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

// 扩展StandardDebugger以支持测试功能
type TestDebugger struct {
	*debug.StandardDebugger
}

func NewTestDebugger() *TestDebugger {
	return &TestDebugger{
		StandardDebugger: debug.NewStandardDebugger(),
	}
}

// RunUnitTest 执行单元测试
func (d *TestDebugger) RunUnitTest(pkg string, verbose bool, cover bool, coverProfile string) error {
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

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 使用调试器的TestEndpoint方法执行命令
	resp, err := d.TestEndpoint(ctx, "go", strings.Join(args, " "), "EXEC", 600)
	if err != nil {
		return fmt.Errorf("单元测试失败: %w", err)
	}

	// 如果生成了覆盖率报告，显示摘要
	if cover {
		absPath, _ := filepath.Abs(coverProfile)
		fmt.Printf("覆盖率报告生成在: %s\n", absPath)

		// 使用go tool cover生成HTML报告
		htmlFile := strings.TrimSuffix(coverProfile, filepath.Ext(coverProfile)) + ".html"

		coverCtx, coverCancel := context.WithTimeout(context.Background(), time.Minute)
		defer coverCancel()

		coverArgs := []string{"tool", "cover", "-html=" + coverProfile, "-o", htmlFile}
		_, err := d.TestEndpoint(coverCtx, "go", strings.Join(coverArgs, " "), "EXEC", 60)
		if err == nil {
			absHtmlPath, _ := filepath.Abs(htmlFile)
			fmt.Printf("HTML覆盖率报告: %s\n", absHtmlPath)
		}
	}

	fmt.Println(debug.FormatHTTPResponse(resp))
	fmt.Println("单元测试完成")
	return nil
}

// RunIntegrationTest 执行集成测试
func (d *TestDebugger) RunIntegrationTest(tags string, setup bool, cleanup bool) error {
	fmt.Printf("运行集成测试，标签: %s\n", tags)

	// 运行测试前的环境设置
	if setup {
		fmt.Println("设置集成测试环境...")
		if err := d.SetupTestEnvironment(); err != nil {
			return fmt.Errorf("设置测试环境失败: %w", err)
		}
	}

	// 延迟执行清理操作
	if cleanup {
		defer func() {
			fmt.Println("清理集成测试环境...")
			if err := d.CleanupTestEnvironment(); err != nil {
				fmt.Printf("清理测试环境时出错: %s\n", err)
			}
		}()
	}

	// 构建测试命令
	args := []string{"test", "-tags", tags, "./..."}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 执行测试
	resp, err := d.TestEndpoint(ctx, "go", strings.Join(args, " "), "EXEC", 600)
	if err != nil {
		return fmt.Errorf("集成测试失败: %w", err)
	}

	fmt.Println(debug.FormatHTTPResponse(resp))
	fmt.Println("集成测试完成")
	return nil
}

// SetupTestEnvironment 设置测试环境
func (d *TestDebugger) SetupTestEnvironment() error {
	fmt.Println("模拟环境设置: 创建测试数据库、启动Docker容器等")
	// 在实际应用中，你可以在这里使用Docker API启动测试容器
	return nil
}

// CleanupTestEnvironment 清理测试环境
func (d *TestDebugger) CleanupTestEnvironment() error {
	fmt.Println("模拟环境清理: 删除测试数据库、停止Docker容器等")
	// 在实际应用中，你可以在这里停止并删除测试容器
	return nil
}

func testUnitAction(c *cli.Context) error {
	pkg := c.String("package")
	verbose := c.Bool("verbose")
	cover := c.Bool("cover")
	coverProfile := c.String("coverprofile")

	// 创建测试调试器
	tester := NewTestDebugger()
	return tester.RunUnitTest(pkg, verbose, cover, coverProfile)
}

func testIntegrationAction(c *cli.Context) error {
	tags := c.String("tags")
	setup := c.Bool("setup")
	cleanup := c.Bool("cleanup")

	// 创建测试调试器
	tester := NewTestDebugger()
	return tester.RunIntegrationTest(tags, setup, cleanup)
}
