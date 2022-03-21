package uptrace

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"path/filepath"
	"text/template"

	"github.com/andrebq/learn-system-design/internal/support/tmplhelper"
)

type (
	Config struct {
		UptraceConfigFile string
		OtelConfigFile    string
		OtelGRPCPort      int
		OtelHTTPPort      int
		JWTSecret         string
		DockerComposeFile string
		UptraceUIPort     int
	}
)

var (
	uptraceTemplate = template.Must(tmplhelper.LoadAll("github.com/andrebq/learn-system-design/internal/support/uptrace",
		"template.docker-compose.yaml",
		"uptrace-config.yaml",
		"otel-config.yaml",
	))
)

func DefaultConfig() (Config, error) {
	var key [32]byte
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		return Config{}, err
	}
	return Config{
		UptraceConfigFile: "./uptrace.yaml",
		OtelConfigFile:    "./otel-config.yaml",
		OtelGRPCPort:      4317,
		OtelHTTPPort:      4318,
		UptraceUIPort:     4319,
		JWTSecret:         hex.EncodeToString(key[:]),
		DockerComposeFile: "uptrace.docker-compose.yaml",
	}, nil
}

func GenerateFiles(config Config) error {
	basepath := filepath.Dir(config.DockerComposeFile)
	computePath := func(child string) string {
		return filepath.Join(basepath, child)
	}
	return tmplhelper.RenderAll(uptraceTemplate, config,
		tmplhelper.RenderConfig{
			Name:       "uptrace-config",
			OutputFile: computePath(config.UptraceConfigFile)},
		tmplhelper.RenderConfig{
			Name:       "template.docker-compose",
			OutputFile: config.DockerComposeFile},
		tmplhelper.RenderConfig{
			Name:       "otel-config",
			OutputFile: computePath(config.OtelConfigFile)},
	)
}
