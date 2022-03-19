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
	var publicEndpoint string = ""
	var controlEndpoint string = "http://127.0.0.1:9002/"
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
			&cli.StringFlag{
				Name: "public-endpoint",
				Usage: `Endpoint used when sending registration information to control plane.

Then endpoint of the control plane itself is defined by the 'control-endpoint' argument.

This argument is mandatory!
`,
				EnvVars:     []string{"LSD_SERVE_PUBLIC_ENDPOINT"},
				Value:       publicEndpoint,
				Destination: &publicEndpoint,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "control-endpoint",
				Usage:       "Base endpoint used to register this instance in the control plane",
				EnvVars:     []string{"LSD_SERVE_CONTROL_ENDPOINT"},
				Value:       controlEndpoint,
				Destination: &controlEndpoint,
			},
		},
		Action: func(c *cli.Context) error {
			h, err := handler.NewHandler(c.Context, initFile, handlerFile, cmdutil.GetInstanceName(), cmdutil.ServiceName(), publicEndpoint, controlEndpoint)
			if err != nil {
				return err
			}
			return cmdutil.RunHTTPServer(c.Context, h, bind)
		},
	}
}
