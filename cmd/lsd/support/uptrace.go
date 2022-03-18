package support

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/andrebq/learn-system-design/internal/cmdutil"
	"github.com/urfave/cli/v2"
)

var (
	uptraceDockerComposeTmpl = template.Must(template.New("__root__").Parse(`
{{ define "uptrace-compose" }}
version: '{{ .ComposeVersion }}'
services:
  clickhouse:
    image: 'yandex/clickhouse-server:21.12'
    environment:
      - CLICKHOUSE_DB=uptrace
    healthcheck:
      test: ['CMD', 'wget', '--spider', '-q', 'localhost:8123/ping']
      interval: 1s
      timeout: 1s
      retries: 30
    {{ if .CustomNetwork }}
    networks:
      - {{ .CustomNetworkName }}
    {{ end }}
  uptrace:
    image: 'uptrace/uptrace:latest'
    volumes:
      - {{ .UptraceConfigFile }}:/etc/uptrace/uptrace.yml
    ports:
      - '{{ .OtelGRPCPort }}:{{ .OtelGRPCPort }}' # OTLP
      - '{{ .OtelHTTPPort }}:{{ .OtelHTTPPort }}' # UI and HTTP API
    environment:
      - DEBUG=2
    depends_on:
      clickhouse:
        condition: service_healthy
    {{ if .CustomNetwork }}
    networks:
      - {{ .CustomNetworkName }}
    {{ end }}
{{ if .CustomNetwork }}
networks:
  {{ .CustomNetworkName }}:
{{ end }}

{{ end }}
{{ define "uptrace-config" }}
secret_key: {{ .JWTSecret }}

# Public URL for Vue-powered UI.
site:
  scheme: 'http'
  host: 'localhost'

listen:
  # OTLP/gRPC API
  grpc: ':{{ .OtelGRPCPort }}'
  # OTLP/HTTP API and Uptrace API
  http: ':{{ .OtelHTTPPort }}'

ch:
  # Connection string for ClickHouse database.
  # clickhouse://<user>:<password>@<host>:<port>/<database>?sslmode=disable
  dsn: 'clickhouse://default:@clickhouse:9000/uptrace?sslmode=disable'

retention:
  # Tell ClickHouse to delete data after 30 days.
  # Supports SQL interval syntax, for example, INTERVAL 30 DAY.
  ttl: 1 DAY

# Uncomment this section to require authentication.
#
# users:
#   - id: 1
#     username: uptrace
#     password: uptrace

projects:
  # First project is used for self-monitoring.
  - id: 1
    name: Uptrace
    token: project1_secret_token

  - id: 2
    name: LearnSystemDesign

# Various limits we apply to queries on spans_index table.
#
# - https://clickhouse.com/docs/en/operations/settings/query-complexity/
# - https://clickhouse.com/docs/en/sql-reference/statements/select/sample/
ch_select_limits:
  # ClickHouse sampling is disabled by default,
  # because it only makes queries faster with big ClickHouse clusters
  sample_rows: 0
  max_rows_to_read: 12e6 # read at most 12 million rows
  max_bytes_to_read: 4e9 # read at most 4 gigabytes of data

{{ end }}
`))
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
	var key [32]byte
	cmdutil.RandomKey(key[:])
	var projectToken [32]byte
	cmdutil.RandomKey(projectToken[:])
	var uptraceConfig = struct {
		ComposeVersion    string
		CustomNetwork     bool
		CustomNetworkName string
		SkipOtelCollector bool
		OtelHTTPPort      int
		OtelGRPCPort      int
		UptraceConfigFile string
		JWTSecret         string
		ProjectToken      string
	}{
		CustomNetwork:     false,
		CustomNetworkName: "uptrace-net",
		ComposeVersion:    "3.7",
		OtelHTTPPort:      4318,
		OtelGRPCPort:      4317,
		UptraceConfigFile: "./uptrace.yaml",
		JWTSecret:         hex.EncodeToString(key[:]),
		ProjectToken:      hex.EncodeToString(projectToken[:]),
	}
	var composeOutput = "uptrace.compose.yaml"
	var skipUptraceConfigFile = false
	return &cli.Command{
		Name:  "docker-compose",
		Usage: "Generates simple docker-compose file that contains a uptrace installation",
		Flags: []cli.Flag{
			cmdutil.StringFlag(&uptraceConfig.JWTSecret, "jwt-secret", "Random key used to sign tokens, each call will generate a new one. Open the compose-file to get its value"),
			cmdutil.StringFlag(&uptraceConfig.ProjectToken, "project-token", "Token for the learn-system-design project on Uptrace"),
			cmdutil.StringFlag(&uptraceConfig.ComposeVersion, "compose-version", "Version of docker compose to use"),
			cmdutil.BoolFlag(&uptraceConfig.CustomNetwork, "use-custom-network", "Indicates if the compose file should include settings for a specific network"),
			cmdutil.StringFlag(&uptraceConfig.CustomNetworkName, "custom-network", "Name of the custom network to use"),
			cmdutil.BoolFlag(&uptraceConfig.SkipOtelCollector, "skip-otel-collector", "Generate the docker-compose file without the Otel Collector"),
			cmdutil.IntFlag(&uptraceConfig.OtelGRPCPort, "otel-grpc-port", "Port to bind the Otel-Collector GRPC Server"),
			cmdutil.IntFlag(&uptraceConfig.OtelHTTPPort, "otel-http-port", "Port to bind the Otel-Collector HTTP Server"),
			cmdutil.StringFlag(&uptraceConfig.UptraceConfigFile, "uptrace-config-file", "File where uptrace config should be saved"),
			cmdutil.StringFlag(&composeOutput, "compose-output", "Location where the docker-compose should be saved"),
			cmdutil.BoolFlag(&skipUptraceConfigFile, "skip-uptrace-config-file", "When true, will not generate otel-collector-yaml (must be provided by the user)"),
		},
		Action: func(ctx *cli.Context) error {
			buf := &bytes.Buffer{}
			var err error
			if err = uptraceDockerComposeTmpl.ExecuteTemplate(buf, "uptrace-compose", uptraceConfig); err != nil {
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
			if !skipUptraceConfigFile {
				buf.Reset()
				if err = uptraceDockerComposeTmpl.ExecuteTemplate(buf, "uptrace-config", uptraceConfig); err != nil {
					return err
				}
				err = ioutil.WriteFile(uptraceConfig.UptraceConfigFile, buf.Bytes(), 0644)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
}
