package utils

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func HandleError(c *fiber.Ctx, statusCode int, message string, err error) error {
	if err == nil {
		return c.Status(statusCode).JSON(fiber.Map{
			"status":  "error",
			"message": message,
			"data":    nil,
		})
	}
	return c.Status(statusCode).JSON(fiber.Map{
		"status":  "error",
		"message": message,
		"data":    err.Error(),
	})
}

func HandleInvalidInputError(c *fiber.Ctx, err error) error {
	return HandleError(c, fiber.StatusBadRequest, "Review your input", err)
}

func HandleInternalServerError(c *fiber.Ctx, err error) error {
	log.Println("Error occurred in server:", err)
	return HandleError(c, fiber.StatusInternalServerError, "Error occured in server", err)
}

func HandleNotFoundOrInternalError(c *fiber.Ctx, err error, notFoundMsg string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return HandleError(c, fiber.StatusNotFound, notFoundMsg, nil)
	}
	return HandleInternalServerError(c, nil)
}

func HandleSuccess(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": message,
		"data":    data,
	})
}

func HandleLoginSuccess(c *fiber.Ctx, message string, token string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": message,
		"token":   token,
		"data":    data,
	})
}
