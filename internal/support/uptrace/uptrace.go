package uptrace

import (
	"crypto/rand"
	"io"
)

type (
	Config struct {
		UptraceConfigFile string
		OtelGRPCPort      int
		OtelHTTPPort      int
		JWTSecret         string
	}
)

func DefaultConfig() (Config, error) {
	var key [32]byte
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		return Config{}, err
	}
	return Config{
		UptraceConfigFile: "uptrace.yaml",
		OtelGRPCPort:      4317,
		OtelHTTPPort:      4318,
	}, nil
}
