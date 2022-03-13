package fleet

import (
	"os"
	"path/filepath"

	"github.com/andrebq/learn-system-design/fleet"
	"github.com/urfave/cli/v2"
)

func Cmd() *cli.Command {
	return &cli.Command{
		Name:  "fleet",
		Usage: "Controls a fleet of lsd tools (stressors, control-panel and services)",
		Subcommands: []*cli.Command{
			localFleetCmd(),
		},
	}
}

func localFleetCmd() *cli.Command {
	stringFlag := func(name, usage string, dest *string) *cli.StringFlag {
		return &cli.StringFlag{
			Name:        name,
			Usage:       usage,
			Destination: dest,
			Value:       *dest,
		}
	}
	var (
		scriptBase = filepath.Join(".", "scripts")
		bindIface  = "127.0.0.1"
		basePort   = 9000
		services   = cli.StringSlice{}
		baseBinary = os.Args[0]
		stressors  = 4
	)
	return &cli.Command{
		Name:  "serve-local",
		Usage: "Starts the local fleet",
		Flags: []cli.Flag{
			stringFlag("baseBinary", "Path to the lsd binary", &baseBinary),
			stringFlag("scriptBase", "Path to the folder which holds all handler scripts", &scriptBase),
			stringFlag("bindIface", "IP of the interface to bind fleet processes", &bindIface),
			&cli.IntFlag{
				Name:        "basePort",
				Usage:       "Lowest port to use when firing up new fleet processes",
				Destination: &basePort,
				Value:       basePort,
			},
			&cli.StringSliceFlag{
				Name:        "app",
				Usage:       "Applicaton (service) to spin up (add multiple times to start different types of services)",
				Destination: &services,
			},
			&cli.IntFlag{
				Name:        "stressors",
				Usage:       "How many stressors to start",
				Destination: &stressors,
				Value:       stressors,
			},
		},
		Action: func(ctx *cli.Context) error {
			m := fleet.NewManager(baseBinary, bindIface, basePort, scriptBase, stressors, services.Value())
			return m.Run(ctx.Context)
		},
	}
}
