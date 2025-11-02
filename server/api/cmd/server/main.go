package main

import (
	"context"
	"os"
	"time"

	"github.com/RowenTey/JustJio/server/api/internal/middlewares"
	"github.com/RowenTey/JustJio/server/api/internal/router"
	"github.com/RowenTey/JustJio/server/api/pkg/app"
	"github.com/RowenTey/JustJio/server/api/pkg/config"
	"github.com/RowenTey/JustJio/server/api/pkg/database"
	"github.com/RowenTey/JustJio/server/api/pkg/kafka"
	"github.com/RowenTey/JustJio/server/api/pkg/logger"
	"github.com/RowenTey/JustJio/server/api/pkg/otel"
	"github.com/RowenTey/JustJio/server/api/pkg/workers"

	"github.com/gofiber/fiber/v2"

	// Swagger docs
	_ "github.com/RowenTey/JustJio/server/api/pkg/docs"
)

// @title JustJio API
// @version 1.0
// @description API server for JustJio.
// @host localhost:8080
// @schemes http
// @BasePath /v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	env := ""
	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	var err error
	appCtx := &app.Context{
		Ctx: context.Background(),
	}

	appCtx.Config, err = config.LoadConfig(env)
	if err != nil {
		panic("Failed to load configuration!")
	}

	appCtx.Logger = logger.InitLogger(appCtx)
	appCtx.Logger.Info("Starting API server...")

	tp, err := otel.InitTracer(appCtx.Config)
	if err != nil {
		appCtx.Logger.Warn("Failed to initialize tracer: ", err)
	}
	defer func() {
		if err := otel.ShutdownTracer(context.Background(), tp); err != nil {
			appCtx.Logger.Error("Failed to shutdown tracer: ", err)
		}
	}()
	appCtx.Logger.Info("OpenTelemetry tracer initialized")

	appCtx.DB = database.ConnectDB(appCtx)

	workerDeps := workers.StartWorkers(appCtx)
	appCtx.NotificationsChan = workerDeps.NotificationsChan

	appCtx.Kafka, err = kafka.NewKafkaClient(
		appCtx.Config,
		appCtx.Logger,
		appCtx.Config.Environment,
	)
	if err != nil {
		appCtx.Logger.Fatal(err)
	}
	defer appCtx.Kafka.Close()

	appCtx.App = fiber.New()

	middlewares.Fiber(appCtx)
	router.Initalize(appCtx)

	// Start HTTP server in background
	go func() {
		appCtx.Logger.Info("Server starting on :", appCtx.Config.Port)
		if err := appCtx.App.Listen(":" + appCtx.Config.Port); err != nil {
			appCtx.Logger.Fatal("Listen failed: ", err)
		}
	}()

	stopCh := app.GracefulShutdown(
		appCtx,
		workerDeps.WorkersWg,
		workerDeps.Scheduler,
		15*time.Second,
	)

	<-stopCh
	appCtx.Logger.Info("Server exited gracefully")
}
