package serve

import (
	"github.com/andrebq/learn-system-design/handler"
	"github.com/andrebq/learn-system-design/internal/cmdutil"
	"github.com/urfave/cli/v2"
)

func Cmd() *cli.Command {
	var bind string = "127.0.0.1:9000"
	var initFile string = "./scripts/init.lua"
	var handlerFile string = "./scripts/handler.lua"
	return &cli.Command{
		Name:  "serve",
		Usage: "Serve the configured handler at the designated port",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "bind",
				Usage:       "Address to bind for incoming connections",
				EnvVars:     []string{"LSD_SERVE_BIND"},
				Value:       bind,
				Destination: &bind,
			},
			&cli.StringFlag{
				Name:        "init-file",
				Usage:       "File with initialization logic",
				EnvVars:     []string{"LSD_SERVE_INIT_FILE"},
				Value:       initFile,
				Destination: &initFile,
			},
			&cli.StringFlag{
				Name:        "handler-file",
				Usage:       "File with the handler logic",
				EnvVars:     []string{"LSD_SERVE_HANDLER_FILE"},
				Value:       handlerFile,
				Destination: &handlerFile,
			},
		},
		Action: func(c *cli.Context) error {
			h, err := handler.NewHandler(c.Context, initFile, handlerFile)
			if err != nil {
				return err
			}
			return cmdutil.RunHTTPServer(c.Context, h, bind)
		},
	}
}
