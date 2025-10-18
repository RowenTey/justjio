package utils

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

// GetOtelContext extracts the OpenTelemetry context from Fiber context
// This allows propagating trace context from HTTP handlers to services and repositories
// so that database queries and service operations appear as child spans in the trace
func GetOtelContext(c *fiber.Ctx) context.Context {
	if ctx, ok := c.Locals("otel-context").(context.Context); ok {
		return ctx
	}
	// Fallback to Fiber's context if OTEL context is not set
	return c.Context()
}
