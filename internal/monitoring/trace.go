package monitoring

import (
	"context"
	"fmt"
	"net/http"

	"github.com/andrebq/learn-system-design/internal/logutil"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

type (
	saveStatusResponse struct {
		http.ResponseWriter
		status int
	}
)

func (s *saveStatusResponse) WriteHeader(st int) {
	if s.status != 0 {
		return
	}
	s.ResponseWriter.WriteHeader(st)
	s.status = st
}

func WithRouteTag(tag string, h http.Handler) http.Handler {
	return otelhttp.WithRouteTag(tag, h)
}

// WrapRootHandler takes a normal http Handler and wraps it
// so that every request will be measured using otel.
//
// This method should be called on Root HTTP handlers instead of using individual handlers.
func WrapRootHandler(ctx context.Context, h http.Handler) http.Handler {
	sampled := logutil.Acquire(ctx).Sample(zerolog.Sometimes)
	sub := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := &saveStatusResponse{w, 0}
		log := logutil.Acquire(r.Context()).With().Stringer("trace-id", trace.SpanContextFromContext(r.Context()).TraceID()).Logger()
		defer func() {
			sampled.Debug().Str(string(semconv.HTTPMethodKey), r.Method).Int(string(semconv.HTTPStatusCodeKey), s.status).Send()
		}()
		r = r.WithContext(logutil.WithLogger(r.Context(), log))
		h.ServeHTTP(s, r)
	})
	return otelhttp.NewHandler(sub, "root-handler")
}

// Tracer is just a syntatic sugar for otel.Tracer(...),
// that way users don't need to worry about otel on most use-cases
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// Measure method implemented by fn using the given tracer
func Measure(ctx context.Context, t trace.Tracer, method string, fn func(context.Context)) {
	ctx, span := t.Start(ctx, method)
	defer recordResult(span)()
	fn(ctx)
}

// MeasureErr works like Measure but returns an error
func MeasureErr(ctx context.Context, t trace.Tracer, method string, fn func(context.Context) error) error {
	ctx, span := t.Start(ctx, method)
	defer recordResult(span)
	return fn(ctx)
}

func recordResult(span trace.Span) func() {
	return func() {
		err := recover()
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%v", err))
			span.SetAttributes(attribute.Bool("github.com.andrebq.panic", true))
			span.End()
			panic(err)
		} else {
			span.End()
		}
	}
}
