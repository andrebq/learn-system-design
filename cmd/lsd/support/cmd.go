package support

import (
	"github.com/urfave/cli/v2"
)

func Cmd() *cli.Command {
	return &cli.Command{
		Name:  "support",
		Usage: "Commands to control support tooling (tracing, logging, etc...)",
		Subcommands: []*cli.Command{
			jaegerCmd(),
			grafanaCmd(),
			uptraceCmd(),
		},
	}
}
