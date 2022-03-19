package grafana

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/rakyll/statik/fs"
)

var (
	grafanaTemplate = mustLoadTemplate()
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

func mustLoadTemplate() *template.Template {
	statikfs, err := fs.NewWithNamespace("github.com/andrebq/learn-system-design/internal/support/grafana")
	if err != nil {
		panic(err)
	}
	t := template.New("__root__")
	for _, fileName := range []string{
		"grafana-datasources.yaml",
		"prometheus-config.yaml",
		"template.docker-compose.yaml",
		"tempo-config.yaml",
	} {
		name := filepath.Base(fileName)[:len(fileName)-len(filepath.Ext(fileName))]
		file, err := statikfs.Open("/" + fileName)
		if err != nil {
			panic(fileName)
		}
		content, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		file.Close()
		_, err = t.New(name).Parse(string(content))
		if err != nil {
			panic(err)
		}
	}
	return t
}

func GenerateFiles(config Config) error {
	type genpair struct {
		template string
		output   string
	}
	parent := filepath.Dir(config.DockerComposeFile)
	for _, p := range []genpair{
		{template: "grafana-datasources", output: filepath.Join(parent, config.DatasourcesFile)},
		{template: "prometheus-config", output: filepath.Join(parent, config.PrometheusConfigFile)},
		{template: "tempo-config", output: filepath.Join(parent, config.TempoConfigFile)},
		{template: "template.docker-compose", output: config.DockerComposeFile},
	} {
		buf := bytes.Buffer{}
		err := grafanaTemplate.ExecuteTemplate(&buf, p.template, config)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(p.output, buf.Bytes(), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
