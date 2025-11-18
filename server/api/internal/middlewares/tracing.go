package middlewares

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig defines the config for tracing middleware
type TracingConfig struct {
	Next              func(c *fiber.Ctx) bool
	TracerName        string
	SpanNameFormatter func(c *fiber.Ctx) string
}

// Tracing returns a Fiber middleware that creates OpenTelemetry spans for HTTP requests
func Tracing(config ...TracingConfig) fiber.Handler {
	cfg := TracingConfig{
		TracerName: "justjio-api",
	}

	if len(config) > 0 {
		if config[0].Next != nil {
			cfg.Next = config[0].Next
		}
		if config[0].TracerName != "" {
			cfg.TracerName = config[0].TracerName
		}
		if config[0].SpanNameFormatter != nil {
			cfg.SpanNameFormatter = config[0].SpanNameFormatter
		}
	}

	tracer := otel.Tracer(cfg.TracerName)
	propagator := otel.GetTextMapPropagator()

	return func(c *fiber.Ctx) error {
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		ctx := propagator.Extract(c.Context(), &fiberCarrier{c: c})

		spanName := cfg.SpanNameFormatter(c)
		ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		c.Locals("otel-context", ctx)

		port, err := strconv.Atoi(c.Port())
		if err != nil {
			port = 0
		}

		span.SetAttributes(
			semconv.HTTPMethod(c.Method()),
			semconv.HTTPTarget(c.OriginalURL()),
			semconv.HTTPRoute(c.Route().Path),
			semconv.HTTPScheme(c.Protocol()),
			attribute.String("http.user_agent", c.Get("User-Agent")),
			semconv.HTTPRequestContentLength(len(c.Body())),
			semconv.NetHostName(c.Hostname()),
			semconv.NetHostPort(port),
			attribute.String("http.client_ip", c.IP()),
		)

		err = c.Next()

		statusCode := c.Response().StatusCode()
		span.SetAttributes(
			semconv.HTTPStatusCode(statusCode),
			semconv.HTTPResponseContentLength(len(c.Response().Body())),
		)

		// TODO: Check out what is going on with status codes and spans
		span.SetStatus(codes.Ok, "")
		if statusCode >= 400 {
			span.SetStatus(codes.Error, fiber.ErrBadRequest.Message)
		}

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		return err
	}
}

// fiberCarrier adapts Fiber context to OpenTelemetry propagation.TextMapCarrier
type fiberCarrier struct {
	c *fiber.Ctx
}

func (fc *fiberCarrier) Get(key string) string {
	return fc.c.Get(key)
}

func (fc *fiberCarrier) Set(key, value string) {
	fc.c.Set(key, value)
}

func (fc *fiberCarrier) Keys() []string {
	keys := make([]string, 0)
	fc.c.Request().Header.VisitAll(func(key, _ []byte) {
		keys = append(keys, string(key))
	})
	return keys
}

// AddSpanAttribute adds an attribute to the current span
func AddSpanAttribute(c *fiber.Ctx, key string, value interface{}) {
	ctx := c.Locals("otel-context")
	if ctx == nil {
		return
	}

	otelCtx, ok := ctx.(context.Context)
	if !ok {
		return
	}

	span := trace.SpanFromContext(otelCtx)
	if span != nil {
		switch v := value.(type) {
		case string:
			span.SetAttributes(attribute.String(key, v))
		case int:
			span.SetAttributes(attribute.Int(key, v))
		case int64:
			span.SetAttributes(attribute.Int64(key, v))
		case bool:
			span.SetAttributes(attribute.Bool(key, v))
		case float64:
			span.SetAttributes(attribute.Float64(key, v))
		}
	}
}
