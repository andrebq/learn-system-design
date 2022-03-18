package handler

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/andrebq/learn-system-design/control"
	"github.com/andrebq/learn-system-design/internal/bindings/handler"
	"github.com/andrebq/learn-system-design/internal/logutil"
	"github.com/andrebq/learn-system-design/internal/monitoring"
	"github.com/andrebq/learn-system-design/internal/mutex"
	lua "github.com/yuin/gopher-lua"
)

type (
	h struct {
		mutex.Zone
		initFile        string
		handlerFile     string
		handlerCode     string
		service         string
		publicEndpoint  string
		controlEndpoint string
		name            string

		instanceData control.Instance

		servers []*control.Server
	}
)

func NewHandler(ctx context.Context, initFile string, handlerFile string, name string, serviceName string, publicEndpoint string, controlEndpoint string) (http.Handler, error) {
	handlerCode, err := ioutil.ReadFile(handlerFile)
	if err != nil {
		return nil, fmt.Errorf("handler: unable to open %v, cause %w", handlerFile, err)
	}
	log := logutil.Acquire(ctx)
	log.Info().Str("initFile", initFile).Str("handlerFile", filepath.Base(handlerFile)).Msg("Preparing new handler")
	h := &h{
		handlerFile:     handlerFile,
		handlerCode:     string(handlerCode),
		initFile:        initFile,
		service:         serviceName,
		publicEndpoint:  publicEndpoint,
		controlEndpoint: controlEndpoint,
		name:            name,
		instanceData: control.Instance{
			Name:     name,
			Services: map[string]string{serviceName: publicEndpoint},
		},
	}
	go h.registration(ctx)
	return h, nil
}

func (h *h) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log := logutil.Acquire(req.Context()).With().Stringer("handler", h).Logger()
	ctx := logutil.WithLogger(req.Context(), log)
	req = req.WithContext(ctx)
	_ = req
	tracer := monitoring.Tracer("lsd-handler")
	monitoring.Measure(ctx, tracer, "handler", func(ctx context.Context) {
		state := h.newState(ctx, w, req)
		state.SetContext(req.Context())
		atomic.AddInt64(&h.instanceData.Metrics.Requests, 1)
		err := state.DoString(h.handlerCode)
		if err != nil {
			log.Error().Err(err).Str("method", req.Method).Stringer("url", req.URL).Msg("Error while processing request")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func (h *h) newState(ctx context.Context, res http.ResponseWriter, req *http.Request) *lua.LState {
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

	var availableServers []*control.Server
	mutex.Run(h.Shared(), func() {
		availableServers = append(availableServers, h.servers...)
	})
	L.PreloadModule("services", handler.ServicesLoader(ctx, availableServers))
	L.PreloadModule("computations", handler.FakeComputations(ctx))
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
		err := control.Register(ctx, h.controlEndpoint, h.service, h.publicEndpoint)
		if err != nil {
			sampled.Error().
				Str("control", h.controlEndpoint).
				Str("name", h.name).
				Str("service", h.service).
				Str("endpoint", h.publicEndpoint).
				Err(err).
				Msg("Unable to register")
		}
		err = control.RegisterInstance(ctx, h.controlEndpoint, h.instanceData)
		if err != nil {
			sampled.Error().
				Str("control", h.controlEndpoint).
				Str("name", h.name).
				Str("service", h.service).
				Str("endpoint", h.publicEndpoint).
				Err(err).
				Msg("Unable to register instance")
		}

		servers, err := control.Services(ctx, h.controlEndpoint)
		if err != nil {
			sampled.Error().
				Str("control", h.controlEndpoint).
				Str("name", h.name).
				Str("service", h.service).
				Str("endpoint", h.publicEndpoint).
				Err(err).
				Msg("Unable to register")
		} else {
			for i, v := range servers {
				if v.Endpoint == h.publicEndpoint {
					servers[i] = nil
				}
			}
			mutex.Run(h.Exclusive(), func() {
				for i := range h.servers {
					h.servers[i] = nil
				}
				h.servers = h.servers[:0]
				for _, v := range servers {
					if v == nil {
						continue
					}
					h.servers = append(h.servers, v)
				}
			})
		}
		select {
		case <-tick.C:
		case <-ctx.Done():
			return
		}
	}
}
