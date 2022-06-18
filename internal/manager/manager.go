package manager

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"sync"
	"unicode/utf8"

	"github.com/julienschmidt/httprouter"
)

type (
	manager struct {
		sync.RWMutex
		codebase map[string]string
	}
)

func Handler() http.Handler {
	m := &manager{
		codebase: map[string]string{},
	}
	router := httprouter.New()
	router.HandlerFunc("PUT", "/code/:serviceType", m.newCode)
	router.HandlerFunc("GET", "/code/:serviceType", m.getCode)
	return router
}

func (m *manager) newCode(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	st := httprouter.ParamsFromContext(ctx).ByName("serviceType")
	ext := path.Ext(st)
	if len(ext) == 0 {
		http.Error(w, "Missing extension, must be .lua", http.StatusNotFound)
		return
	}
	st = st[:len(st)-len(ext)]
	code, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}
	if len(code) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}
	if !utf8.Valid(code) {
		http.Error(w, "Code must be utf-8 encoded", http.StatusNotAcceptable)
		return
	}
	m.Lock()
	m.codebase[st] = string(code)
	m.Unlock()
	w.WriteHeader(http.StatusOK)
}

func (m *manager) getCode(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	st := httprouter.ParamsFromContext(ctx).ByName("serviceType")
	ext := path.Ext(st)
	if len(ext) == 0 {
		http.Error(w, "Missing extension, must be .lua", http.StatusNotFound)
		return
	}
	st = st[:len(st)-len(ext)]
	m.RLock()
	code, ok := m.codebase[st]
	m.RUnlock()
	if !ok {
		http.Error(w, fmt.Sprintf("No code for %v", st), http.StatusNotFound)
		return
	}
	w.Header().Add("Content-Length", strconv.Itoa(len(code)))
	w.Header().Add("Content-Type", "text/x-lua; charset=utf-8")
	io.WriteString(w, code)
}
