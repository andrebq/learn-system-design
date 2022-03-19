package cmdutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var (
	serviceVersion  string = "dev"
	serviceEnv      string = "unspecified"
	serviceName     string = mustStr(os.Executable())
	useGRPCExporter bool   = false
)

func mustStr(s string, e error) string {
	if e != nil {
		panic(e)
	}
	return filepath.Base(s)
}

func Exporter(ctx context.Context) (trace.SpanExporter, error) {
	var client otlptrace.Client
	if useGRPCExporter {
		client = otlptracegrpc.NewClient(otlptracegrpc.WithEndpoint("localhost:4317"), otlptracegrpc.WithInsecure())
	} else {
		client = otlptracehttp.NewClient(otlptracehttp.WithEndpoint("localhost:4318"), otlptracehttp.WithInsecure())
	}
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("cmdutil: unable to setup otel exporter")
	}
	return exporter, nil
}

func ExporterTypeFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:        "otel-exporter.use-grpc",
		Usage:       "Indicates if the service should use GRPC or not for Otel Export",
		EnvVars:     []string{"LSD_OTEL_EXPORTER_USE_GRPC"},
		Value:       useGRPCExporter,
		Destination: &useGRPCExporter,
	}
}

func ServiceName() string {
	return serviceName
}

func ServiceNameFlag() cli.Flag {
	return &cli.StringFlag{
		Name:        "otel.service.name",
		Usage:       "Indicates the name of the service",
		EnvVars:     []string{"APP_NAME", "OTEL_SERVICE_NAME"},
		Value:       serviceName,
		Destination: &serviceName,
	}
}

func ServiceVersionFlag() cli.Flag {
	return &cli.StringFlag{
		Name:        "otel.service.version",
		Usage:       "Indicates the service version.",
		EnvVars:     []string{"APP_VERSION", "OTEL_SERVICE_VERSION"},
		Value:       serviceVersion,
		Destination: &serviceVersion,
	}
}

func ServiceEnvFlag() cli.Flag {
	return &cli.StringFlag{
		Name:        "otel.deployment.environment",
		Usage:       "Indicates the environment where the service is running",
		EnvVars:     []string{"APP_ENV", "OTEL_DEPLOYMENT_ENVIRONMENT"},
		Value:       serviceEnv,
		Destination: &serviceEnv,
	}
}

// Resource returns a resource that was configured by the
// flags provided to this process
func Resource() *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			semconv.DeploymentEnvironmentKey.String(serviceEnv),
			semconv.ServiceInstanceIDKey.String(GetInstanceName()),
		),
	)
	return r
}
