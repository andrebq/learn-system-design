package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/andrebq/learn-system-design/cmd/lsd/ctl"
	"github.com/andrebq/learn-system-design/cmd/lsd/generators"
	"github.com/andrebq/learn-system-design/cmd/lsd/manager"
	"github.com/andrebq/learn-system-design/cmd/lsd/userservice"
	"github.com/andrebq/learn-system-design/internal/cmdutil"
	"github.com/andrebq/learn-system-design/internal/logutil"
	"github.com/andrebq/learn-system-design/internal/monitoring"
	"github.com/rs/zerolog"
	logpkg "github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func newApp(parentCtx context.Context) *cli.App {
	var sigCancel context.CancelFunc
	var logLevel string = zerolog.InfoLevel.String()
	app := &cli.App{
		Name:  "lsd - learn system design",
		Usage: "Simple application to help teach system design various audiences",
		Commands: []*cli.Command{
			userservice.Cmd(),
			manager.Cmd(),
			/*
				stress.Cmd(),
				control.Cmd(),
				fleet.Cmd(),
				support.Cmd(),
			*/
			ctl.Cmd(),
			generators.Cmd(),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log-level",
				Usage:       "Controls how verbose the log will be",
				Destination: &logLevel,
				Value:       logLevel,
			},
			cmdutil.InstanceNameFlag(),
			cmdutil.ServiceNameFlag(),
			cmdutil.ServiceEnvFlag(),
			cmdutil.ServiceVersionFlag(),
			cmdutil.ExporterTypeFlag(),
		},
		After: func(ctx *cli.Context) error {
			if sigCancel != nil {
				sigCancel()
			}
			shutdownCtx, cancel := context.WithTimeout(parentCtx, time.Minute)
			defer cancel()
			monitoring.ShutdownProvider(shutdownCtx)
			return nil
		},
		Before: func(ctx *cli.Context) error {
			exp, err := cmdutil.Exporter(parentCtx)
			if err != nil {
				return err
			}
			monitoring.InitTraceProvider(exp, cmdutil.Resource())
			var ll zerolog.Level
			switch logLevel {
			case zerolog.LevelDebugValue, zerolog.LevelTraceValue:
				ll = zerolog.DebugLevel
			case zerolog.LevelErrorValue, zerolog.LevelFatalValue, zerolog.LevelPanicValue:
				ll = zerolog.ErrorLevel
			case zerolog.LevelWarnValue:
				ll = zerolog.WarnLevel
			}
			log := logpkg.Level(ll)
			appCtx := logutil.WithLogger(ctx.Context, log)
			var sigCtx context.Context
			sigCtx, sigCancel = signal.NotifyContext(appCtx, os.Interrupt)

			ctx.Context = sigCtx
			logpkg.Logger = log
			return nil
		},
	}
	return app
}

func main() {
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()
	log := logutil.Acquire(logutil.WithLogger(rootCtx, logpkg.Logger))
	app := newApp(rootCtx)
	err := app.RunContext(rootCtx, os.Args)
	if err != nil {
		log.Fatal().Strs("args", os.Args).Err(err).Msg("Application failed")
	}
}
