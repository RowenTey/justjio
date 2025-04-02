package middleware

import (
	"strings"

	"github.com/RowenTey/JustJio/config"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
)

func jwtError(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Missing or malformed JWT",
			"data":    nil,
		})
	}

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"message": "Unauthorized. Invalid or expired JWT",
		"data":    nil,
	})
}

func whitelist(c *fiber.Ctx) bool {
	whitelistPaths := []string{"/v1/auth", "/docs"}
	whitelistEndpoints := []string{"/", "/openapi.yaml"}

	for _, url := range whitelistPaths {
		if strings.HasPrefix(c.Path(), url) {
			return true
		}
	}

	for _, url := range whitelistEndpoints {
		if c.Path() == url {
			return true
		}
	}

	return false
}

func Authenticated() fiber.Handler {
	return jwtware.New(jwtware.Config{
		Filter:       whitelist,
		SigningKey:   []byte(config.Config("JWT_SECRET")),
		ErrorHandler: jwtError,
	})
}
