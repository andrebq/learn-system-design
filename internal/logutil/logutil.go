package logutil

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type (
	ctxkey byte
)

const (
	loggerKey = ctxkey(1)
)

func Acquire(ctx context.Context) zerolog.Logger {
	l := ctx.Value(loggerKey)
	if l != nil {
		return l.(zerolog.Logger)
	}
	return log.Logger
}

func WithLogger(ctx context.Context, log zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}
