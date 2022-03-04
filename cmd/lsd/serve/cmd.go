package serve

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/andrebq/learn-system-design/handler"
	"github.com/andrebq/learn-system-design/internal/logutil"
	"github.com/urfave/cli/v2"
)

func Cmd() *cli.Command {
	var bind string = "127.0.0.1:9000"
	var initFile string = "./scripts/init.lua"
	var handlerFile string = "./scripts/handler.lua"
	return &cli.Command{
		Name:  "serve",
		Usage: "Serve the configured handler at the designated port",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "bind",
				Usage:       "Address to bind for incoming connections",
				EnvVars:     []string{"LSD_SERVE_BIND"},
				Value:       bind,
				Destination: &bind,
			},
			&cli.StringFlag{
				Name:        "init-file",
				Usage:       "File with initialization logic",
				EnvVars:     []string{"LSD_SERVE_INIT_FILE"},
				Value:       initFile,
				Destination: &initFile,
			},
			&cli.StringFlag{
				Name:        "handler-file",
				Usage:       "File with the handler logic",
				EnvVars:     []string{"LSD_SERVE_HANDLER_FILE"},
				Value:       handlerFile,
				Destination: &handlerFile,
			},
		},
		Action: func(c *cli.Context) error {
			log := logutil.Acquire(c.Context)
			rootCtx, cancel := context.WithCancel(c.Context)
			defer cancel()
			h, err := handler.NewHandler(rootCtx, initFile, handlerFile)
			if err != nil {
				return err
			}
			server := &http.Server{Addr: bind, Handler: h}
			shutdown := make(chan struct{})
			go func() {
				defer close(shutdown)
				<-rootCtx.Done()
				ctx, cancel := context.WithTimeout(rootCtx, time.Minute)
				defer cancel()
				server.Shutdown(ctx)
			}()

			serveErr := make(chan error)
			go func() {
				defer cancel()
				log.Info().Str("binding", server.Addr).Msg("Starting server")
				serveErr <- server.ListenAndServe()
			}()

			<-rootCtx.Done()
			cancel()
			err = <-serveErr
			if errors.Is(err, http.ErrServerClosed) {
				err = nil
			}
			<-shutdown
			return err
		},
	}
}
