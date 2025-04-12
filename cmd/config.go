package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/parker/ParkerCli/internal/config"
	"github.com/urfave/cli/v2"
)

var ConfigCommand = &cli.Command{
	Name:  "config",
	Usage: "管理项目配置文件，例如初始化、查看、更新",
	Subcommands: []*cli.Command{
		{
			Name:  "init",
			Usage: "初始化配置文件",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "env", Value: "development", Usage: "环境: development, testing, production"},
			},
			Action: configInitAction,
		},
		{
			Name:   "show",
			Usage:  "展示当前配置",
			Action: configShowAction,
		},
		{
			Name:  "set",
			Usage: "更新某配置字段，例如 --key=DB_HOST --value=127.0.0.1",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "key", Usage: "配置键"},
				&cli.StringFlag{Name: "value", Usage: "配置值"},
			},
			Action: configSetAction,
		},
	},
}

func configInitAction(c *cli.Context) error {
	env := c.String("env")
	fmt.Printf("正在生成默认配置文件...\n")

	// 获取工作目录
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前目录失败: %w", err)
	}

	// 配置文件路径
	configFilePath := filepath.Join(dir, "config.yaml")

	// 检查文件是否已存在
	if _, err := os.Stat(configFilePath); err == nil {
		fmt.Printf("配置文件已存在: %s\n", configFilePath)
		fmt.Println("如需重新初始化，请先删除现有配置文件")
		return nil
	}

	// 初始化配置模块
	if err := config.Init(configFilePath); err != nil {
		return fmt.Errorf("初始化配置模块失败: %w", err)
	}

	// 设置环境
	config.Set("environment", env)

	// 使用默认配置结构中的值
	if err := config.Save(); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	fmt.Printf("配置文件已成功创建: %s\n", configFilePath)
	return nil
}

func configShowAction(c *cli.Context) error {
	// 初始化配置模块
	if err := config.Init(""); err != nil {
		return err
	}

	// 如果配置文件不存在
	if !config.ConfigFileExists() {
		fmt.Println("当前没有配置文件，请先运行 'ParkerCli config init' 初始化配置")
		return nil
	}

	fmt.Printf("配置文件: %s\n", config.ConfigFileUsed())
	fmt.Println("当前配置如下:")

	// 获取所有配置
	cfg := config.GetAll()

	// 打印应用配置
	fmt.Printf("app_name=%s\n", cfg.AppName)
	fmt.Printf("version=%s\n", cfg.Version)
	fmt.Printf("environment=%s\n", cfg.Environment)
	fmt.Printf("debug=%v\n", cfg.Debug)

	// 打印服务器配置
	fmt.Printf("server.port=%d\n", cfg.Server.Port)
	fmt.Printf("server.host=%s\n", cfg.Server.Host)
	fmt.Printf("server.read_timeout=%d\n", cfg.Server.ReadTimeout)
	fmt.Printf("server.write_timeout=%d\n", cfg.Server.WriteTimeout)

	// 打印数据库配置
	fmt.Printf("database.driver=%s\n", cfg.Database.Driver)
	fmt.Printf("database.host=%s\n", cfg.Database.Host)
	fmt.Printf("database.port=%d\n", cfg.Database.Port)
	fmt.Printf("database.name=%s\n", cfg.Database.Name)
	fmt.Printf("database.user=%s\n", cfg.Database.User)

	// 打印日志配置
	fmt.Printf("log.level=%s\n", cfg.Log.Level)
	fmt.Printf("log.format=%s\n", cfg.Log.Format)
	fmt.Printf("log.output=%s\n", cfg.Log.Output)

	// 打印Docker配置
	fmt.Printf("docker.registry=%s\n", cfg.Docker.Registry)
	fmt.Printf("docker.namespace=%s\n", cfg.Docker.Namespace)

	// 打印路径配置
	for k, v := range cfg.Paths {
		fmt.Printf("paths.%s=%s\n", k, v)
	}

	return nil
}

func configSetAction(c *cli.Context) error {
	key := c.String("key")
	value := c.String("value")

	if key == "" || value == "" {
		return fmt.Errorf("key 或 value 不可为空")
	}

	// 初始化配置模块
	if err := config.Init(""); err != nil {
		return err
	}

	// 如果配置文件不存在
	if !config.ConfigFileExists() {
		fmt.Println("当前没有配置文件，请先运行 'ParkerCli config init' 初始化配置")
		return nil
	}

	// 设置新值
	config.Set(key, value)

	// 保存配置
	if err := config.Save(); err != nil {
		return fmt.Errorf("更新配置文件失败: %w", err)
	}

	fmt.Printf("成功更新配置: %s=%s\n", key, value)
	return nil
}
