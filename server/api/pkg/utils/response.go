package utils

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// EmptyApiResponse is used for Swagger documentation when no data is returned
type EmptyApiResponse struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message" example:"Operation completed successfully"`
	Data    any    `json:"data" swaggertype:"object"`
}

type ApiResponse[T any] struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func HandleError(c *fiber.Ctx, statusCode int, message string, err error) error {
	var errorData any
	if err != nil {
		errorData = err.Error()
	}

	return c.Status(statusCode).JSON(ApiResponse[any]{
		Status:  "error",
		Message: message,
		Data:    errorData,
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
	return HandleInternalServerError(c, err)
}

func HandleSuccess[T any](c *fiber.Ctx, message string, data T) error {
	return c.Status(fiber.StatusOK).JSON(ApiResponse[T]{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func HandleLoginSuccess(c *fiber.Ctx, message string, token string, data any) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": message,
		"token":   token,
		"data":    data,
	})
}
