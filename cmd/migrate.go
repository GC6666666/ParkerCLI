package cmd

import (
	"fmt"

	"github.com/parker/ParkerCli/internal/migrator"
	"github.com/parker/ParkerCli/pkg/logger"
	"github.com/urfave/cli/v2"
)

var MigrateCommand = &cli.Command{
	Name:  "migrate",
	Usage: "数据库迁移管理",
	Subcommands: []*cli.Command{
		{
			Name:  "up",
			Usage: "升级数据库到最新版本",
			Flags: []cli.Flag{
				&cli.IntFlag{Name: "step", Value: 0, Usage: "迁移步数，0表示全部"},
			},
			Action: migrateUpAction,
		},
		{
			Name:  "down",
			Usage: "回退到上一个数据库版本",
			Flags: []cli.Flag{
				&cli.IntFlag{Name: "step", Value: 1, Usage: "回退步数"},
			},
			Action: migrateDownAction,
		},
		{
			Name:   "status",
			Usage:  "查看当前数据库迁移状态",
			Action: migrateStatusAction,
		},
		{
			Name:  "create",
			Usage: "创建新的迁移文件",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "name", Required: true, Usage: "迁移名称"},
				&cli.StringFlag{Name: "type", Value: "sql", Usage: "迁移类型 (sql 或 go)"},
			},
			Action: migrateCreateAction,
		},
		{
			Name:   "reset",
			Usage:  "重置所有迁移（先回退后升级）",
			Action: migrateResetAction,
		},
		{
			Name:   "refresh",
			Usage:  "刷新所有迁移（与reset相同）",
			Action: migrateRefreshAction,
		},
	},
}

// 获取迁移器实例
func getMigrator() *migrator.StandardMigrator {
	return migrator.NewStandardMigrator()
}

func migrateUpAction(c *cli.Context) error {
	steps := c.Int("step")

	logger.Info("执行数据库升级，步数: %d", steps)

	m := getMigrator()
	if err := m.Up(steps); err != nil {
		logger.Error("数据库升级失败: %v", err)
		return err
	}

	logger.Info("数据库升级完成")
	return nil
}

func migrateDownAction(c *cli.Context) error {
	steps := c.Int("step")

	logger.Info("执行数据库回退，步数: %d", steps)

	m := getMigrator()
	if err := m.Down(steps); err != nil {
		logger.Error("数据库回退失败: %v", err)
		return err
	}

	logger.Info("数据库回退完成")
	return nil
}

func migrateStatusAction(c *cli.Context) error {
	logger.Info("查看数据库迁移状态")

	m := getMigrator()
	migrations, err := m.Status()
	if err != nil {
		logger.Error("获取迁移状态失败: %v", err)
		return err
	}

	// 显示格式化的迁移状态
	fmt.Println(migrator.FormatMigrationStatus(migrations))

	return nil
}

func migrateCreateAction(c *cli.Context) error {
	name := c.String("name")
	typeStr := c.String("type")

	if name == "" {
		return fmt.Errorf("必须提供迁移名称")
	}

	var migrationType migrator.MigrationType
	if typeStr == "go" {
		migrationType = migrator.TypeGoFn
	} else {
		migrationType = migrator.TypeSQL
	}

	logger.Info("创建新迁移: %s, 类型: %s", name, typeStr)

	m := getMigrator()
	migration, err := m.Create(name, migrationType)
	if err != nil {
		logger.Error("创建迁移失败: %v", err)
		return err
	}

	logger.Info("创建迁移成功: %s (ID: %s)", migration.Name, migration.ID)
	logger.Info("迁移文件位置: %s", migration.FilePath)

	return nil
}

func migrateResetAction(c *cli.Context) error {
	logger.Info("重置所有迁移")

	m := getMigrator()
	if err := m.Reset(); err != nil {
		logger.Error("重置迁移失败: %v", err)
		return err
	}

	logger.Info("重置迁移完成")
	return nil
}

func migrateRefreshAction(c *cli.Context) error {
	logger.Info("刷新所有迁移")

	m := getMigrator()
	if err := m.Refresh(); err != nil {
		logger.Error("刷新迁移失败: %v", err)
		return err
	}

	logger.Info("刷新迁移完成")
	return nil
}
