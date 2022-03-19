package support

import (
	"github.com/andrebq/learn-system-design/internal/cmdutil"
	"github.com/andrebq/learn-system-design/internal/support/grafana"
	"github.com/urfave/cli/v2"
)

func grafanaCmd() *cli.Command {
	return &cli.Command{
		Name:  "grafana",
		Usage: "Commands to setup a complete monitoring stack using software from Grafana Labs",
		Subcommands: []*cli.Command{
			grafanaDockerComposeCmd(),
		},
	}
}

func grafanaDockerComposeCmd() *cli.Command {
	config := grafana.DefaultConfig()
	return &cli.Command{
		Name:  "docker-compose",
		Usage: "Generates a docker-compose (and any other support files)",
		Flags: []cli.Flag{
			cmdutil.StringFlag(&config.DockerComposeFile, "compose-file", "Destination of the docker-compose file"),
			cmdutil.StringFlag(&config.DatasourcesFile, "grafana-datasources", "Destination of the Grafana datasources config (relative to the docker-compose)"),
			cmdutil.StringFlag(&config.PrometheusConfigFile, "prometheus-file", "Destination of the prometheus compose file (relative to docker-compose)"),
			cmdutil.StringFlag(&config.TempoConfigFile, "tempo-file", "Destination of the tempo config file (relative to docker-compose)"),
		},
		Action: func(ctx *cli.Context) error {
			return grafana.GenerateFiles(config)
		},
	}
}
