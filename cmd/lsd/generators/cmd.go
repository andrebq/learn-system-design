package generators

import (
	"github.com/andrebq/learn-system-design/cmd/lsd/generators/kubegen"
	"github.com/urfave/cli/v2"
)

func Cmd() *cli.Command {
	return &cli.Command{
		Name:    "generators",
		Aliases: []string{"gen"},
		Usage:   "Generate files for docker-compose, kubernetes, etc...",
		Subcommands: []*cli.Command{
			kubegen.Cmd(),
		},
	}
}
