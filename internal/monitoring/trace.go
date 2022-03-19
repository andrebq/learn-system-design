package monitoring

import (
	"context"
	"fmt"
	"net/http"

	"github.com/andrebq/learn-system-design/internal/logutil"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// WrapHandler takes a normal http Handler and wraps it
// so that every request will be measured using otel
func WrapHandler(h http.Handler) http.Handler {
	tracer := otel.Tracer("monitoring.http")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Measure(r.Context(), tracer, "http.handler", func(ctx context.Context) {
			id := trace.SpanFromContext(ctx).SpanContext().TraceID()
			if id.IsValid() {
				log := logutil.Acquire(ctx).With().Stringer("trace-id", id).Logger()
				ctx = logutil.WithLogger(ctx, log)
			}
			r = r.WithContext(ctx)
			h.ServeHTTP(w, r)
		})
	})
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
