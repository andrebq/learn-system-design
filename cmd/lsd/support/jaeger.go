package support

import (
	"bytes"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/andrebq/learn-system-design/internal/cmdutil"
	"github.com/urfave/cli/v2"
)

var (
	jagerDockerComposeTmpl = template.Must(template.New("__root__").Parse(`
{{ define "otel-config" }}
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:{{ .OtelGRPCPort }}
      http:
        endpoint: 0.0.0.0:{{ .OtelHTTPPort }}

exporters:
  jaeger:
    endpoint: http://jaeger:14250
  logging:

processors:
  batch:

extensions:
  health_check:
  pprof:

service:
  extensions: [pprof, health_check]
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [jaeger, logging]
      processors: [batch]
{{ end }}

{{ define "jaeger-compose" }}
version: '{{ .ComposeVersion }}'
services:
  {{ if not .SkipOtelCollector }}
  otel-collector:
    image: otel/opentelemetry-collector:latest
    command: ["--config=/etc/otel-collector-config.yml"]
    volumes:
      - {{ .OtelConfigFile }}:/etc/otel-collector-config.yml
    ports:
      - "{{ .OtelHTTPPort }}:{{ .OtelHTTPPort }}" # Otel collector HTTP Port
      - "{{ .OtelGRPCPort }}:{{ .OtelGRPCPort }}" # Otel collector GRPC Port
      - "1888:1888"   # pprof extension
      - "8888:8888"   # Prometheus metrics exposed by the collector
      - "8889:8889"   # Prometheus exporter metrics
      - "13133:13133" # health_check extension
    depends_on:
      - jaeger
  {{ end }}
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
		SkipOtelCollector bool
		OtelHTTPPort      int
		OtelGRPCPort      int
		OtelConfigFile    string
	}{
		UDPPort:           6831,
		HTTPPort:          16686,
		IncludeSampleApp:  false,
		SampleAppPort:     8080,
		CustomNetwork:     false,
		CustomNetworkName: "jaeger-net",
		ComposeVersion:    "3.7",
		SkipOtelCollector: false,
		OtelHTTPPort:      4318,
		OtelGRPCPort:      4317,
		OtelConfigFile:    "./otel-collector-config.yaml",
	}
	var composeOutput = "jaeger.compose.yaml"
	var skipOtelConfigFile = false
	return &cli.Command{
		Name:  "docker-compose",
		Usage: "Generates simple docker-compose file that contains a jaeger installation",
		Flags: []cli.Flag{
			cmdutil.StringFlag(&jagerConfig.ComposeVersion, "compose-version", "Version of docker compose to use"),
			cmdutil.IntFlag(&jagerConfig.UDPPort, "udp-port", "Port to bind the UDP Jager Agent"),
			cmdutil.IntFlag(&jagerConfig.HTTPPort, "http-port", "Port to bind the Jaeger UI"),
			cmdutil.BoolFlag(&jagerConfig.IncludeSampleApp, "include-sample-app", "Include the hot-rod sample app from Jaeger"),
			cmdutil.IntFlag(&jagerConfig.SampleAppPort, "sample-app-port", "PORT used by the sample app"),
			cmdutil.BoolFlag(&jagerConfig.CustomNetwork, "use-custom-network", "Indicates if the compose file should include settings for a specific network"),
			cmdutil.StringFlag(&jagerConfig.CustomNetworkName, "custom-network", "Name of the custom network to use"),
			cmdutil.BoolFlag(&jagerConfig.SkipOtelCollector, "skip-otel-collector", "Generate the docker-compose file without the Otel Collector"),
			cmdutil.IntFlag(&jagerConfig.OtelGRPCPort, "otel-grpc-port", "Port to bind the Otel-Collector GRPC Server"),
			cmdutil.IntFlag(&jagerConfig.OtelHTTPPort, "otel-http-port", "Port to bind the Otel-Collector HTTP Server"),
			cmdutil.StringFlag(&jagerConfig.OtelConfigFile, "otel-config-file", "File where otel config should be saved"),
			cmdutil.StringFlag(&composeOutput, "compose-output", "Location where the docker-compose should be saved"),
			cmdutil.BoolFlag(&skipOtelConfigFile, "skip-otel-config-file", "When true, will not generate otel-collector-yaml (must be provided by the user)"),
		},
		Action: func(ctx *cli.Context) error {
			buf := &bytes.Buffer{}
			var err error
			if err = jagerDockerComposeTmpl.ExecuteTemplate(buf, "jaeger-compose", jagerConfig); err != nil {
				return err
			}
			if composeOutput == "-" {
				_, err = os.Stdout.Write(buf.Bytes())
				if err != nil {
					return err
				}
			} else {
				err = ioutil.WriteFile(composeOutput, buf.Bytes(), 0644)
				if err != nil {
					return err
				}
			}
			if !skipOtelConfigFile {
				buf.Reset()
				if err = jagerDockerComposeTmpl.ExecuteTemplate(buf, "otel-config", jagerConfig); err != nil {
					return err
				}
				err = ioutil.WriteFile(jagerConfig.OtelConfigFile, buf.Bytes(), 0644)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
}
