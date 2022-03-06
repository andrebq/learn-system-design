package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"time"

	"github.com/andrebq/learn-system-design/control"
	"github.com/andrebq/learn-system-design/internal/bindings/handler"
	"github.com/andrebq/learn-system-design/internal/logutil"
	lua "github.com/yuin/gopher-lua"
)

type (
	h struct {
		initFile        string
		handlerFile     string
		service         string
		publicEndpoint  string
		controlEndpoint string
		name            string
	}
)

func NewHandler(ctx context.Context, initFile string, handlerFile string, name string, publicEndpoint string, controlEndpoint string) (http.Handler, error) {
	if len(name) == 0 {
		u, err := url.Parse(publicEndpoint)
		if err != nil {
			return nil, err
		}
		name = u.Host
	}
	log := logutil.Acquire(ctx)
	log.Info().Str("initFile", initFile).Str("handlerFile", filepath.Base(handlerFile)).Msg("Preparing new handler")
	h := &h{
		handlerFile:     handlerFile,
		initFile:        initFile,
		service:         filepath.Base(filepath.Dir(initFile)),
		publicEndpoint:  publicEndpoint,
		controlEndpoint: controlEndpoint,
		name:            name,
	}
	go h.registration(ctx)
	return h, nil
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

func (h *h) registration(ctx context.Context) {
	if h.controlEndpoint == "" {
		return
	}
	runtime.Gosched()
	sampled := logutil.Acquire(ctx) //.Sample(zerolog.Sometimes)
	tick := time.NewTicker(time.Second * 5)
	for {
		err := control.Register(ctx, h.controlEndpoint, h.name, h.service, h.publicEndpoint)
		if err != nil {
			sampled.Error().
				Str("control", h.controlEndpoint).
				Str("name", h.name).
				Str("service", h.service).
				Str("endpoint", h.publicEndpoint).
				Err(err).
				Msg("Unable to register")
		}
		select {
		case <-tick.C:
		case <-ctx.Done():
			return
		}
	}
}
