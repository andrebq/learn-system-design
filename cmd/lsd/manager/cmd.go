package manager

import (
	"fmt"

	"github.com/andrebq/learn-system-design/internal/cmdutil"
	"github.com/andrebq/learn-system-design/internal/manager"
	"github.com/urfave/cli/v2"
)

func Cmd() *cli.Command {
	var bindAddr string = "127.0.0.1"
	var bindPort uint = 9001
	return &cli.Command{
		Name:  "manager",
		Usage: "Starts the manager that provides logic to all other services on the fleet",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "bind-interface",
				Aliases:     []string{"bind", "iface", "interface"},
				Usage:       "IP to bind for incoming connections",
				EnvVars:     []string{"LSD_MANAGER_BIND"},
				Value:       bindAddr,
				Destination: &bindAddr,
			},
			&cli.UintFlag{
				Name:        "bind-port",
				Aliases:     []string{"port"},
				Usage:       "Port to bind for incoming connections",
				EnvVars:     []string{"LSD_MANAGER_BIND_PORT"},
				Value:       bindPort,
				Destination: &bindPort,
			},
		},
		Action: func(c *cli.Context) error {
			return cmdutil.RunHTTPServer(c.Context, manager.Handler(), fmt.Sprintf("%v:%v", bindAddr, bindPort))
		},
	}
}
