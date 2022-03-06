package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/andrebq/learn-system-design/cmd/lsd/control"
	"github.com/andrebq/learn-system-design/cmd/lsd/serve"
	"github.com/andrebq/learn-system-design/cmd/lsd/stress"
	"github.com/andrebq/learn-system-design/internal/logutil"
	"github.com/rs/zerolog"
	logpkg "github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	log := logutil.Acquire(logutil.WithLogger(ctx, logpkg.Logger))
	var logLevel string = zerolog.InfoLevel.String()
	app := &cli.App{
		Name:  "lsd - learn system design",
		Usage: "Simple application to help teach system design various audiences",
		Commands: []*cli.Command{
			serve.Cmd(),
			stress.Cmd(),
			control.Cmd(),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log-level",
				Usage:       "Controls how verbose the log will be",
				Destination: &logLevel,
				Value:       logLevel,
			},
		},
		Before: func(ctx *cli.Context) error {
			var ll zerolog.Level
			switch logLevel {
			case zerolog.LevelDebugValue, zerolog.LevelTraceValue:
				ll = zerolog.DebugLevel
			case zerolog.LevelErrorValue, zerolog.LevelFatalValue, zerolog.LevelPanicValue:
				ll = zerolog.ErrorLevel
			case zerolog.LevelWarnValue:
				ll = zerolog.WarnLevel
			}
			log = log.Level(ll)
			ctx.Context = logutil.WithLogger(ctx.Context, log)
			logpkg.Logger = log
			return nil
		},
	}

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("Application failed")
	}
}
