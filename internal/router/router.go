package router

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type (
	R interface {
		HandlerFunc(method string, route string, handler http.HandlerFunc)
		ServeHTTP(http.ResponseWriter, *http.Request)
	}

	wrap struct {
		*httprouter.Router
	}
)

func New() R {
	return wrap{httprouter.New()}
}

func (r wrap) HandlerFunc(method string, route string, handler http.HandlerFunc) {
	r.Handler(method, route, otelhttp.WithRouteTag(route, handler))
}
