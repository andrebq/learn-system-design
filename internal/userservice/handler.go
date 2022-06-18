package userservice

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"

	handlerBindings "github.com/andrebq/learn-system-design/internal/bindings/handler"
	"github.com/andrebq/learn-system-design/internal/logutil"
	"github.com/rs/zerolog"

	lua "github.com/yuin/gopher-lua"
)

type (
	handler struct {
		sync.RWMutex
		code        string
		manager     *url.URL
		serviceType string
	}
)

func Handler(rootCtx context.Context, managerURL *url.URL, serviceType string) http.Handler {
	h := &handler{
		manager:     managerURL,
		serviceType: serviceType,
	}
	go h.checkManager(rootCtx)
	return h
}

func (h *handler) checkManager(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cli := http.Client{
		Transport: http.DefaultTransport,
	}

	codeurl := *h.manager
	codeurl.Path = path.Join(codeurl.Path, fmt.Sprintf("/code/%v.lua", h.serviceType))
	loadCode := func(ctx context.Context) (string, error) {
		req, err := http.NewRequest("GET", codeurl.String(), nil)
		if err != nil {
			return "", err
		}
		req = req.WithContext(ctx)
		res, err := cli.Do(req)
		if err != nil {
			return "", err
		}
		if res.StatusCode != 200 {
			return "", fmt.Errorf("unexpected status code: %v", res.StatusCode)
		}
		defer res.Body.Close()
		buf, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}
	code, err := loadCode(ctx)
	if err == nil {
		h.replaceCode(code)
	}
	log := logutil.Acquire(ctx).With().Stringer("manager-endpoint", &codeurl).Str("service-type", h.serviceType).Logger()
	sample := log.Sample(zerolog.Often)
	tick := time.NewTicker(time.Second)
	oldCode := code
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			code, err := loadCode(ctx)
			if err != nil {
				sample.Error().Err(err).Msg("Unable to fetch code from manager")
				continue
			}
			if code != oldCode {
				log.Info().Msg("Code replacement!")
				h.replaceCode(code)
				oldCode = code
			}
		}
	}
}

func (h *handler) replaceCode(newCode string) {
	h.Lock()
	h.code = newCode
	h.Unlock()
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.RLock()
	code := h.code
	h.RUnlock()
	ctx := req.Context()
	log := logutil.Acquire(ctx)
	if len(code) == 0 {
		http.Error(w, "User-Service has not code attached to it", http.StatusInternalServerError)
		return
	}
	st := h.newState(req.Context(), req, w)
	err := st.DoString(code)
	if err != nil {
		log.Error().Err(err).Str("url", req.URL.Path).Msg("Error while handling request")
		http.Error(w, fmt.Sprintf("Internal server error while handling request: \n%v", err), http.StatusInternalServerError)
		return
	}
}

func (h *handler) newState(ctx context.Context, req *http.Request, res http.ResponseWriter) *lua.LState {
	L := lua.NewState(lua.Options{
		SkipOpenLibs:        true,
		RegistryMaxSize:     1_000_000,
		IncludeGoStackTrace: false,
		CallStackSize:       100,
	})
	L.SetContext(ctx)
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
	L.PreloadModule("handler", handlerBindings.Loader(req, res))
	L.PreloadModule("services", handlerBindings.ServicesLoader(ctx))
	L.PreloadModule("computations", handlerBindings.FakeComputations(ctx))
	return L
}
