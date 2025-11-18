package middlewares

import (
	"strings"
	"time"

	"github.com/RowenTey/JustJio/server/api/pkg/app"
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func Fiber(ctx *app.Context) {
	a := ctx.App
	env := ctx.Config.Environment

	serviceName := "justjio-api"
	if env != "production" {
		serviceName += "-" + env
	}

	prometheus := fiberprometheus.New(serviceName)
	prometheus.RegisterAt(a, "/metrics")
	// Skip the root path (healthcheck) and favicon path for metrics
	prometheus.SetSkipPaths([]string{"/", "/favicon.ico"})

	a.Use(
		// CORS setting
		cors.New(cors.Config{
			AllowOrigins: ctx.Config.AllowedOrigins,
			AllowHeaders: "Origin, Content-Type, Accept, Authorization, CF-Access-Client-Id, CF-Access-Client-Secret",
			AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		}),

		// OpenTelemetry tracing
		Tracing(TracingConfig{
			Next: func(c *fiber.Ctx) bool {
				// Skip healthcheck, metrics, Swagger endpoints
				return c.Path() == "/" || c.Path() == "/metrics" || c.Path() == "/favicon.ico" || strings.HasPrefix(c.Path(), "/docs")
			},
			TracerName: serviceName,
			SpanNameFormatter: func(c *fiber.Ctx) string {
				return c.Method() + " " + c.Path()
			},
		}),

		// Prometheus metrics
		prometheus.Middleware,

		// Rate limiting
		// TODO: Make this configurable per environment
		limiter.New(limiter.Config{
			Next: func(c *fiber.Ctx) bool {
				return c.IP() == "127.0.0.1" // Don't limit from localhost
			},
			Max:        50,
			Expiration: 30 * time.Second,
			LimitReached: func(c *fiber.Ctx) error {
				return c.
					Status(fiber.StatusTooManyRequests).
					SendString("Rate Limit Exceeded! Please wait 30s before making a request again...")
			},
		}),

		// Logging
		logger.New(logger.Config{
			Next: func(c *fiber.Ctx) bool {
				return !strings.HasPrefix(c.Path(), "/v1")
			},
			Format:     "time=${time} level=info | ${latency} | ${status} - ${method} ${path}\n",
			TimeZone:   "Asia/Singapore",
			TimeFormat: time.RFC3339,
		}),

		// Accept application/json
		func(c *fiber.Ctx) error {
			c.Accepts("application/json")
			return c.Next()
		},

		// JWT middleware
		Authenticated(ctx.Config.JwtSecret),
	)
}
