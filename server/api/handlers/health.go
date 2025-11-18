package handlers

import (
	"time"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/gofiber/fiber/v2"
)

type HealthCheckResponse struct {
	Status   string            `json:"status"`
	Uptime   float64           `json:"uptime_seconds"`
	Services map[string]string `json:"services"`
}

var startTime = time.Now()

// HealthCheck performs a comprehensive health check of the service and its dependencies
func HealthCheck(c *fiber.Ctx) error {
	response := HealthCheckResponse{
		Status:   "healthy",
		Uptime:   time.Since(startTime).Seconds(),
		Services: make(map[string]string),
	}

	// Check database connectivity
	sqlDB, err := database.DB.DB()
	if err != nil {
		response.Status = "unhealthy"
		response.Services["database"] = "unavailable"
	} else {
		if err := sqlDB.Ping(); err != nil {
			response.Status = "degraded"
			response.Services["database"] = "unreachable"
		} else {
			response.Services["database"] = "healthy"
		}
	}

	statusCode := fiber.StatusOK
	if response.Status == "unhealthy" {
		statusCode = fiber.StatusServiceUnavailable
	} else if response.Status == "degraded" {
		statusCode = fiber.StatusOK // Still return 200 for degraded state
	}

	return c.Status(statusCode).JSON(response)
}

// LivenessProbe is a simple endpoint to check if the service is running
func LivenessProbe(c *fiber.Ctx) error {
	return c.SendString("OK")
}

// ReadinessProbe checks if the service is ready to accept traffic
func ReadinessProbe(c *fiber.Ctx) error {
	// Check database connectivity
	sqlDB, err := database.DB.DB()
	if err != nil || sqlDB.Ping() != nil {
		return c.Status(fiber.StatusServiceUnavailable).SendString("NOT READY")
	}

	return c.SendString("READY")
}
