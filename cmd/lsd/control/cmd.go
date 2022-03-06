package control

import (
	"github.com/andrebq/learn-system-design/control"
	"github.com/andrebq/learn-system-design/internal/cmdutil"
	"github.com/urfave/cli/v2"
)

func Cmd() *cli.Command {
	return &cli.Command{
		Name:        "control-plane",
		Usage:       "Commands to interact with the control-plane.",
		Subcommands: []*cli.Command{serveCmd()},
	}
}

func serveCmd() *cli.Command {
	var bind string = "127.0.0.1:9002"
	return &cli.Command{
		Name:  "serve",
		Usage: "Runs the control plane that is used to run and configure simulations",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "bind",
				Usage:       "Address to bind and listen for incoming connections",
				EnvVars:     []string{"LSD_CONTROL_PLANE_BIND"},
				Destination: &bind,
				Value:       bind,
			},
		},
		Action: func(ctx *cli.Context) error {
			h := control.Handler()
			return cmdutil.RunHTTPServer(ctx.Context, h, bind)
		},
	}
}
