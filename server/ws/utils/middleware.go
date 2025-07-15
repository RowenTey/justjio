package utils

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/sirupsen/logrus"
)

// custom CORS middleware for websocket endpoint
func IsAllowedOrigins(conf *Config, logger *logrus.Logger, c *fiber.Ctx) error {
	origin := c.Get("Origin")
	if origin == "" {
		return c.Status(403).SendString("Forbidden")
	}
	logger.WithField("service", "CORS").Debug("Origin:", origin)

	allowedOrigins := strings.Split(conf.AllowedOrigins, ",")
	for _, allowedOrigin := range allowedOrigins {
		if origin == strings.TrimSpace(allowedOrigin) {
			return c.Next()
		}
	}

	return c.Status(403).SendString("Forbidden")
}

func WebSocketUpgrade(c *fiber.Ctx) error {
	// IsWebSocketUpgrade returns true if the client
	// requested upgrade to the WebSocket protocol.
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("websocket", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}
