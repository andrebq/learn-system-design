package cmdutil

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/andrebq/learn-system-design/internal/monitoring"
	"github.com/rs/zerolog/log"
)

func RunHTTPServer(parentCtx context.Context, h http.Handler, bind string) error {
	h = monitoring.WrapHandler(parentCtx, h)
	rootCtx, cancel := context.WithCancel(parentCtx)
	defer cancel()
	server := &http.Server{Addr: bind, Handler: h, BaseContext: func(l net.Listener) context.Context { return parentCtx }}
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
	err := <-serveErr
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	<-shutdown
	return err
}
