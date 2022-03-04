package handler

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/andrebq/learn-system-design/internal/logutil"
	"github.com/rs/zerolog"
	"github.com/steinfletcher/apitest"
)

func TestHandler(t *testing.T) {
	ctx := logutil.WithLogger(context.Background(), zerolog.Nop())
	initFile := filepath.Join("testdata", "fixture", "test-handler", "init.lua")
	handlerFile := filepath.Join("testdata", "fixture", "test-handler", "handler.lua")
	h, err := NewHandler(ctx, initFile, handlerFile)
	if err != nil {
		t.Fatal(err)
	}
	apitest.Handler(h).Debug().Put("/data.json").Body(`{"salute":"World"}`).Expect(t).Status(http.StatusOK).End()
}
