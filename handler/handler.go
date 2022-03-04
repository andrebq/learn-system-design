package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/andrebq/learn-system-design/internal/logutil"
)

type (
	h struct {
		initFile    string
		handlerFile string
	}
)

func NewHandler(ctx context.Context, initFile string, handlerFile string) (http.Handler, error) {
	log := logutil.Acquire(ctx)
	log.Info().Str("initFile", initFile).Str("handlerFile", handlerFile).Msg("Preparing new handler")
	return &h{}, nil
}

func (h *h) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := logutil.WithLogger(req.Context(), logutil.Acquire(req.Context()).With().Stringer("handler", h).Logger())
	req = req.WithContext(ctx)
	_ = req
	http.Error(w, "Not Implemented", http.StatusInternalServerError)
}

func (h *h) String() string {
	return fmt.Sprintf("handler init: %v / handler: %v", h.initFile, h.handlerFile)
}
