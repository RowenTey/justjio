package handlers

import (
	"context"
	"os"
	"testing"

	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/sirupsen/logrus"
)

var (
	IsPackageTest bool = false
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	logger := logrus.New()

	// Setup test containers
	logger.Info("Starting test dependencies...")
	dependencies := &tests.TestDependencies{}
	dependencies, err := tests.SetupTestDependencies(ctx, dependencies, logger)
	if err != nil {
		logger.Errorf("failed to start container: %v\n", err)
		os.Exit(1)
	}
	IsPackageTest = true

	// Run tests
	logger.Info("Running tests...")
	code := m.Run()

	// Cleanup
	logger.Info("Tearing down dependencies...")
	if dependencies != nil {
		dependencies.Teardown(ctx)
	}
	IsPackageTest = false

	os.Exit(code)
}
