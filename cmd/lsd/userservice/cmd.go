package userservice

import (
	"fmt"
	"net/url"

	"github.com/andrebq/learn-system-design/internal/cmdutil"
	"github.com/andrebq/learn-system-design/internal/userservice"
	"github.com/urfave/cli/v2"
)

func Cmd() *cli.Command {
	var bindAddr string = "127.0.0.1"
	var bindPort uint = 9000
	var manager string = "http://127.0.0.1:9001/"
	var serviceType string = "user-service"
	return &cli.Command{
		Name:  "user-service",
		Usage: "Starts a user-service handler that will execute Lua scripts",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "bind-interface",
				Aliases:     []string{"bind", "iface", "interface"},
				Usage:       "IP to bind for incoming connections",
				EnvVars:     []string{"LSD_USER_SERVICE_BIND"},
				Value:       bindAddr,
				Destination: &bindAddr,
			},
			&cli.UintFlag{
				Name:        "bind-port",
				Aliases:     []string{"port"},
				Usage:       "Port to bind for incoming connections",
				EnvVars:     []string{"LSD_USER_SERVICE_BIND_PORT"},
				Value:       bindPort,
				Destination: &bindPort,
			},
			&cli.StringFlag{
				Name:        "manager",
				Usage:       "Base endpoint where the manager is located",
				EnvVars:     []string{"LSD_USER_SERVICE_MANAGER_ENDPOINT"},
				Value:       manager,
				Destination: &manager,
			},
			&cli.StringFlag{
				Name:        "service-type",
				Usage:       "Used by the handler to lookup code on the manager",
				EnvVars:     []string{"LSD_USER_SERVICE_SERVICE_TYPE"},
				Destination: &serviceType,
				Value:       serviceType,
			},
		},
		Action: func(c *cli.Context) error {
			managerurl, err := url.Parse(manager)
			if err != nil {
				return err
			}
			return cmdutil.RunHTTPServer(c.Context, userservice.Handler(c.Context, managerurl, serviceType), fmt.Sprintf("%v:%v", bindAddr, bindPort))
		},
	}
}
