package monitoring

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

var (
	initProvider     sync.Once
	shutdownProvider sync.Once

	traceProvider *trace.TracerProvider
	shutdownErr   error
)

// InitTraceProvider using the given exporter and resource, this will be configured
// as the default provider for all otel calls
func InitTraceProvider(exp trace.SpanExporter, res *resource.Resource) bool {
	var firstInit bool
	initProvider.Do(func() {
		traceProvider = trace.NewTracerProvider(
			trace.WithBatcher(exp),
			trace.WithResource(res),
		)
		firstInit = true
		otel.SetTracerProvider(traceProvider)
	})
	return firstInit
}

func ShutdownProvider(ctx context.Context) error {
	shutdownProvider.Do(func() {
		shutdownErr = traceProvider.Shutdown(ctx)
	})
	return shutdownErr
}
