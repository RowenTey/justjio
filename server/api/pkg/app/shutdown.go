package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
)

// GracefulShutdown coordinates everything that must stop.
// Call it once from main().
func GracefulShutdown(
	appCtx *Context,
	workersWg *sync.WaitGroup,
	scheduler *gocron.Scheduler,
	timeout time.Duration,
) <-chan struct{} {
	logger := appCtx.Logger

	stop := make(chan struct{})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-sigs
		logger.Info("Received signal: ", sig)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := appCtx.App.ShutdownWithContext(ctx); err != nil {
			logger.Error("HTTP server forced shutdown: ", err)
		}

		close(appCtx.NotificationsChan)

		// Wait for in-flight work
		done := make(chan struct{})
		go func() {
			workersWg.Wait()
			close(done)
		}()

		// Stop scheduler
		if err := (*scheduler).Shutdown(); err != nil {
			logger.Error("Scheduler forced shutdown: ", err)
		}

		// Wait for workers with timeout
		select {
		case <-done:
			logger.Info("All workers drained")
		case <-ctx.Done():
			logger.Warn("Worker drain timeout – forcing exit")
		}

		logger.Info("Shutdown complete")
		close(stop)
	}()

	return stop
}
