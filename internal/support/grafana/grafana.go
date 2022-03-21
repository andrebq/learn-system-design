package grafana

import (
	"path/filepath"
	"text/template"

	"github.com/andrebq/learn-system-design/internal/support/tmplhelper"
)

var (
	grafanaTemplate = template.Must(tmplhelper.LoadAll("github.com/andrebq/learn-system-design/internal/support/grafana",
		"grafana-datasources.yaml",
		"prometheus-config.yaml",
		"template.docker-compose.yaml",
		"tempo-config.yaml",
	))
)

type (
	Config struct {
		DatasourcesFile      string
		PrometheusConfigFile string
		TempoConfigFile      string
		DockerComposeFile    string
	}
)

func DefaultConfig() Config {
	return Config{
		DatasourcesFile:      "./grafana-datasources.yaml",
		PrometheusConfigFile: "./prometheus-config.yaml",
		TempoConfigFile:      "./tempo-config.yaml",
		DockerComposeFile:    "tempo.docker-compose.yaml",
	}
}

func GenerateFiles(config Config) error {
	basepath := filepath.Dir(config.DockerComposeFile)
	computePath := func(child string) string {
		return filepath.Join(basepath, child)
	}
	return tmplhelper.RenderAll(grafanaTemplate, config,
		tmplhelper.RenderConfig{
			Name:       "grafana-datasources",
			OutputFile: computePath(config.DatasourcesFile)},
		tmplhelper.RenderConfig{
			Name:       "prometheus-config",
			OutputFile: computePath(config.PrometheusConfigFile),
		},
		tmplhelper.RenderConfig{
			Name:       "tempo-config",
			OutputFile: computePath(config.TempoConfigFile)},
		tmplhelper.RenderConfig{
			Name:       "template.docker-compose",
			OutputFile: config.DockerComposeFile})
}
