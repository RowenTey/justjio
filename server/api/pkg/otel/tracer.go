package otel

import (
	"context"
	"fmt"
	"time"

	"github.com/RowenTey/JustJio/server/api/pkg/app"
	"github.com/RowenTey/JustJio/server/api/pkg/config"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Tracer manages OpenTelemetry tracing with OTLP HTTP exporter for Grafana Tempo
type Tracer struct {
	provider *sdktrace.TracerProvider
	config   *config.Config
	logger   *logrus.Entry
}

func New(appCtx *app.Context) (*Tracer, error) {
	ctx := context.Background()
	conf := appCtx.Config

	otlpEndpoint := conf.TracingEndpoint
	if otlpEndpoint == "" {
		otlpEndpoint = "localhost:4318"
	}

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

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(conf.Version),
			semconv.DeploymentEnvironment(conf.Environment),
		),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
		// Sample all traces in dev/staging, 10% in production
		sdktrace.WithSampler(getSampler(conf.Environment)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	appCtx.Logger.Info("OpenTelemetry tracer initialized")
	return &Tracer{
		provider: tp,
		config:   conf,
		logger:   appCtx.Logger.WithField("component", "OtelTracer"),
	}, nil
}

// Shutdown gracefully shuts down the tracer provider
func (t *Tracer) Shutdown(ctx context.Context) {
	if t.provider == nil {
		return
	}

	// Allow 5 seconds for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := t.provider.Shutdown(shutdownCtx); err != nil {
		t.logger.Errorf("failed to shutdown tracer provider: %v", err)
	}
}

// Provider returns the underlying TracerProvider
func (t *Tracer) Provider() *sdktrace.TracerProvider {
	return t.provider
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
