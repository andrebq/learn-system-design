package support

import (
	"github.com/andrebq/learn-system-design/internal/cmdutil"
	"github.com/andrebq/learn-system-design/internal/support/uptrace"
	"github.com/urfave/cli/v2"
)

func uptraceCmd() *cli.Command {
	return &cli.Command{
		Name:  "uptrace",
		Usage: "Commands to spin-up uptrace in container or kubernetes",
		Subcommands: []*cli.Command{
			uptraceDockerComposeCmd(),
		},
	}
}

func uptraceDockerComposeCmd() *cli.Command {
	uptraceConfig, err := uptrace.DefaultConfig()
	if err != nil {
		panic(err)
	}
	var skipUptraceConfigFile = false
	return &cli.Command{
		Name:  "docker-compose",
		Usage: "Generates simple docker-compose file that contains a uptrace installation",
		Flags: []cli.Flag{
			cmdutil.StringFlag(&uptraceConfig.JWTSecret, "jwt-secret", "Random key used to sign tokens, each call will generate a new one. Open the compose-file to get its value"),
			cmdutil.IntFlag(&uptraceConfig.OtelGRPCPort, "otel-grpc-port", "Port to bind the Otel-Collector GRPC Server"),
			cmdutil.IntFlag(&uptraceConfig.OtelHTTPPort, "otel-http-port", "Port to bind the Otel-Collector HTTP Server"),
			cmdutil.IntFlag(&uptraceConfig.UptraceUIPort, "uptrace-ui-port", "Port to bind on the host to access Uptrace Otel and UI services"),
			cmdutil.StringFlag(&uptraceConfig.OtelConfigFile, "otel-config-file", "File where otel-collector config should be saved"),
			cmdutil.StringFlag(&uptraceConfig.UptraceConfigFile, "uptrace-config-file", "File where uptrace config should be saved"),
			cmdutil.StringFlag(&uptraceConfig.DockerComposeFile, "compose-file", "Location where the docker-compose should be saved"),
			cmdutil.BoolFlag(&skipUptraceConfigFile, "skip-uptrace-config-file", "When true, will not generate otel-collector-yaml (must be provided by the user)"),
		},
		Action: func(ctx *cli.Context) error {
			return uptrace.GenerateFiles(uptraceConfig)
		},
	}
}
