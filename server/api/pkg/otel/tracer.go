package otel

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/RowenTey/JustJio/server/api/pkg/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// InitTracer initializes OpenTelemetry tracer with OTLP HTTP exporter for Grafana Tempo
// Returns the TracerProvider which must be shut down gracefully on application exit
func InitTracer(conf *config.Config) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	otlpEndpoint := conf.TracingEndpoint
	if otlpEndpoint == "" {
		otlpEndpoint = "localhost:4318"
	}

	// Create OTLP HTTP exporter
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(otlpEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	serviceName := "justjio-api"
	if conf.Environment != "production" {
		serviceName += "-" + conf.Environment
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(os.Getenv("VERSION")),
			semconv.DeploymentEnvironment(conf.Environment),
		),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider with batch span processor
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
		// Sample all traces in dev/staging, 10% in production
		sdktrace.WithSampler(getSampler(conf.Environment)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator to W3C Trace Context (standard for distributed tracing)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// getSampler returns appropriate sampler based on environment
func getSampler(environment string) sdktrace.Sampler {
	if environment == "production" {
		// Sample 10% of traces in production to reduce overhead
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1))
	}
	// Sample all traces in dev/staging
	return sdktrace.AlwaysSample()
}

// ShutdownTracer gracefully shuts down the tracer provider
func ShutdownTracer(ctx context.Context, tp *sdktrace.TracerProvider) error {
	if tp == nil {
		return nil
	}

	// Allow 5 seconds for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := tp.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown tracer provider: %w", err)
	}

	return nil
}
