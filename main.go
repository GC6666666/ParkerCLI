package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/parker/ParkerCli/cmd"
)

func main() {
	app := &cli.App{
		Name:    "ParkerCli",
		Usage:   "一款面向 Go 后端开发调试、部署、发布的全能 CLI 工具",
		Version: "0.1.0",
		Commands: []*cli.Command{
			cmd.DebugCommand,
			cmd.RunCommand,
			cmd.ConfigCommand,
			cmd.BuildCommand,
			cmd.TestCommand,
			cmd.LogsCommand,
			cmd.MigrateCommand,
			cmd.ReleaseCommand,
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "查看当前 ParkerCli 版本",
				Action: func(c *cli.Context) error {
					log.Printf("ParkerCli version: %s", c.App.Version)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
