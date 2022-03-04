package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/andrebq/learn-system-design/cmd/lsd/serve"
	"github.com/andrebq/learn-system-design/internal/logutil"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	log := logutil.Acquire(logutil.WithLogger(ctx, log.Logger))
	app := &cli.App{
		Name:  "lsd - learn system design",
		Usage: "Simple application to help teach system design various audiences",
		Commands: []*cli.Command{
			serve.Cmd(),
		},
	}

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("Application failed")
	}
}
