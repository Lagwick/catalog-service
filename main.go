package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/Lagwick/catalog-service/cmd"
	"github.com/Lagwick/catalog-service/internal/app/constant"
	msentry "github.com/Lagwick/catalog-service/internal/app/monitor/sentry"
)

func main() {
	app := &cli.App{
		Name:    constant.AppName,
		Version: constant.Version,
		Usage:   "Catalog management service",
		Commands: []*cli.Command{
			cmd.WebServer(),
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "no-json",
				Usage: "Enable console logger instead of JSON",
			},
		},
	}

	defer msentry.Flush()

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
