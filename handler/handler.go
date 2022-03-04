package handler

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/andrebq/learn-system-design/internal/bindings/handler"
	"github.com/andrebq/learn-system-design/internal/logutil"
	lua "github.com/yuin/gopher-lua"
)

type (
	h struct {
		initFile    string
		handlerFile string
	}
)

func NewHandler(ctx context.Context, initFile string, handlerFile string) (http.Handler, error) {
	log := logutil.Acquire(ctx)
	log.Info().Str("initFile", initFile).Str("handlerFile", filepath.Base(handlerFile)).Msg("Preparing new handler")
	return &h{
		handlerFile: handlerFile,
		initFile:    initFile,
	}, nil
}

func (h *h) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log := logutil.Acquire(req.Context()).With().Stringer("handler", h).Logger()
	ctx := logutil.WithLogger(req.Context(), log)
	req = req.WithContext(ctx)
	_ = req
	state := h.newState(w, req)
	err := state.DoFile(h.handlerFile)
	if err != nil {
		log.Error().Err(err).Str("method", req.Method).Stringer("url", req.URL).Msg("Error while processing request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *h) newState(res http.ResponseWriter, req *http.Request) *lua.LState {
	L := lua.NewState(lua.Options{
		SkipOpenLibs:        true,
		RegistryMaxSize:     1_000_000,
		IncludeGoStackTrace: false,
		CallStackSize:       100,
	})
	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage}, // Must be first
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
	} {
		if err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(pair.f),
			NRet:    0,
			Protect: true,
		}, lua.LString(pair.n)); err != nil {
			panic(err)
		}
	}
	L.PreloadModule("handler", handler.Loader(req, res))
	return L
}

func (h *h) String() string {
	return fmt.Sprintf("handler init: %v / handler: %v", h.initFile, filepath.Base(h.handlerFile))
}
