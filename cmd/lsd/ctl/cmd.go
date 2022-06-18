package ctl

import (
	"github.com/andrebq/learn-system-design/cmd/lsd/ctl/managerctl"
	"github.com/urfave/cli/v2"
)

func Cmd() *cli.Command {
	return &cli.Command{
		Name:  "ctl",
		Usage: "Subcommands to interact with various components of the system",
		Subcommands: []*cli.Command{
			managerctl.Cmd(),
		},
	}
}
