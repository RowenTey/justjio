package middlewares

import (
	"github.com/RowenTey/JustJio/server/api/pkg/utils"
	"github.com/RowenTey/JustJio/server/api/pkg/validator"
	"github.com/gofiber/fiber/v2"
)

// ParseAndValidate is a middleware factory that parses and validates request bodies
func ParseAndValidate[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req T

		if err := c.BodyParser(&req); err != nil {
			return utils.HandleInvalidInputError(c, err)
		}

		if err := validator.ValidateStruct(&req); err != nil {
			return utils.HandleError(c, fiber.StatusBadRequest, err.Error(), err)
		}

		// Store the validated request in locals for the handler to use
		c.Locals("validatedRequest", &req)

		return c.Next()
	}
}

// GetValidatedRequest retrieves the validated request from context
func GetValidatedRequest[T any](c *fiber.Ctx) *T {
	if req := c.Locals("validatedRequest"); req != nil {
		if typedReq, ok := req.(*T); ok {
			return typedReq
		}
	}
	return nil
}
