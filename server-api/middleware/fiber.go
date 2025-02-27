package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func Fiber(a *fiber.App) {
	a.Use(
		// CORS setting
		cors.New(cors.Config{
			AllowOrigins:     "http://localhost:5173, https://justjio-staging.rowentey.xyz",
			AllowHeaders:     "Origin, Content-Type, Accept, Authorization, CF-Access-Client-Id, CF-Access-Client-Secret",
			AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
			AllowCredentials: true,
		}),

		// Rate limiting
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
			Format: "[${ip}:${port}] ${status} - ${method} ${path}\n",
		}),

		// Accept application/json
		func(c *fiber.Ctx) error {
			c.Accepts("application/json")
			return c.Next()
		},

		// JWT middleware
		Authenticated(),
	)
}
