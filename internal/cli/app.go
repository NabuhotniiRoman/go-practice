package cli

import (
	"github.com/urfave/cli/v2"
)

// NewApp створює новий CLI додаток
func NewApp() *cli.App {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "configure",
				Usage: "Generate configuration from template",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "template",
						Aliases: []string{"t"},
						Usage:   "Path to HCL template file",
						Value:   "configs/oidc-api.hcl.tmpl",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output configuration file path",
						Value:   "_local.hcl",
					},
					&cli.StringFlag{
						Name:    "version",
						Aliases: []string{"v"},
						Usage:   "Build version",
						Value:   "dev",
					},
					&cli.StringFlag{
						Name:    "mode",
						Aliases: []string{"m"},
						Usage:   "Configuration mode (local, staging, production)",
						Value:   "local",
					},
				},
				Action: configureAction,
			},
			{
				Name:  "server",
				Usage: "Start the API server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "Configuration file path",
						Value:   "_local.hcl",
					},
				},
				Action: serverAction,
			},
			{
				Name:   "version",
				Usage:  "Show version information",
				Action: versionAction,
			},
		},
	}

	return app
}
