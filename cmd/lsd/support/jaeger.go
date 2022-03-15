package support

import (
	"os"
	"text/template"

	"github.com/andrebq/learn-system-design/internal/cmdutil"
	"github.com/urfave/cli/v2"
)

var (
	jagerDockerComposeTmpl = template.Must(template.New("__root__").Parse(`
version: '{{ .ComposeVersion }}'
services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "{{ .UDPPort }}:{{ .UDPPort }}/udp"
      - "{{ .HTTPPort }}:{{ .HTTPPort }}"
    {{ if .CustomNetwork }}
    networks:
      - {{ .CustomNetworkName }}
    {{ end }}
{{ if .IncludeSampleApp }}
  hotrod:
    image: jaegertracing/example-hotrod:latest
    ports:
      - "{{ .SampleAppPort }}:{{ .SampleAppPort}}"
    command: ["all"]
    environment:
      - JAEGER_AGENT_HOST=jaeger
      - JAEGER_AGENT_PORT=6831
      {{ if .CustomNetwork }}
      networks:
        - {{ .CustomNetworkName }}
      {{ end }}
    depends_on:
      - jaeger
{{ end }}
{{ if .CustomNetwork }}
networks:
  {{ .CustomNetworkName }}:
{{ end }}
`))
)

func jaegerCmd() *cli.Command {
	return &cli.Command{
		Name:  "jaeger",
		Usage: "Commands to spin-up jaeger in container or kubernetes",
		Subcommands: []*cli.Command{
			jaegerDockerComposeCmd(),
		},
	}
}

func jaegerDockerComposeCmd() *cli.Command {
	var jagerConfig = struct {
		UDPPort           int
		HTTPPort          int
		ComposeVersion    string
		IncludeSampleApp  bool
		SampleAppPort     int
		CustomNetwork     bool
		CustomNetworkName string
	}{
		UDPPort:           6831,
		HTTPPort:          16686,
		IncludeSampleApp:  false,
		SampleAppPort:     8080,
		CustomNetwork:     false,
		CustomNetworkName: "jaeger-net",
		ComposeVersion:    "3.7",
	}
	return &cli.Command{
		Name:  "docker-compose",
		Usage: "Print a simple docker-compose file that contains a jaeger installation",
		Flags: []cli.Flag{
			cmdutil.StringFlag(&jagerConfig.ComposeVersion, "compose-version", "Version of docker compose to use"),
			cmdutil.IntFlag(&jagerConfig.UDPPort, "udp-port", "Port to bind the UDP Jager Agent"),
			cmdutil.IntFlag(&jagerConfig.HTTPPort, "http-port", "Port to bind the HTTP/OTEL collector"),
			cmdutil.BoolFlag(&jagerConfig.IncludeSampleApp, "include-sample-app", "Include the hot-rod sample app from Jaeger"),
			cmdutil.IntFlag(&jagerConfig.SampleAppPort, "sample-app-port", "PORT used by the sample app"),
			cmdutil.BoolFlag(&jagerConfig.CustomNetwork, "use-custom-network", "Indicates if the compose file should include settings for a specific network"),
			cmdutil.StringFlag(&jagerConfig.CustomNetworkName, "custom-network", "Name of the custom network to use"),
		},
		Action: func(ctx *cli.Context) error {
			return jagerDockerComposeTmpl.Execute(os.Stdout, jagerConfig)
		},
	}
}
