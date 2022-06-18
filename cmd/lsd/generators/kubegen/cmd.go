package kubegen

import (
	"os"

	"github.com/andrebq/learn-system-design/internal/generators/kubegen"
	"github.com/urfave/cli/v2"
)

func Cmd() *cli.Command {
	lsdRepo := "andrebq/lsd"
	lsdLabel := "latest"
	var services cli.StringSlice
	var ingresses cli.StringSlice
	var pullIfNotPresent bool
	return &cli.Command{
		Name:    "kubegen",
		Aliases: []string{"k8s"},
		Usage:   "Generate a full k8s deployment manifest (as yaml)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "lsd-repo",
				Value:       lsdRepo,
				Destination: &lsdRepo,
				Usage:       "Repository where lsd images are hosted",
			},
			&cli.StringFlag{
				Name:        "lsd-label",
				Value:       lsdLabel,
				Destination: &lsdLabel,
				Usage:       "Label to use for LSD images",
			},
			&cli.BoolFlag{
				Name:        "pull-if-not-present",
				Value:       pullIfNotPresent,
				Destination: &pullIfNotPresent,
				Usage:       "Useful when deploying to local k8s, as it allows for docker-desktop image to be used by the local cluster",
			},
			&cli.StringSliceFlag{
				Name:        "service",
				Aliases:     []string{"s"},
				Destination: &services,
				Usage:       "Name of a `service` to generate. Actual service code is loaded from the lsd manager",
			},
			&cli.StringSliceFlag{
				Name:        "ingresses",
				Aliases:     []string{"i"},
				Destination: &ingresses,
				Usage:       "List of simple ingress definitions (service:port:hostname:path)",
			},
		},
		Action: func(ctx *cli.Context) error {
			return kubegen.Generate(ctx.Context, os.Stdout, lsdRepo, lsdLabel, pullIfNotPresent, services.Value(), ingresses.Value())
		},
	}
}
