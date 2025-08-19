package tests

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDependencies struct {
	PostgresContainer *postgres.PostgresContainer
	KafkaContainer    *kafka.KafkaContainer
}

func SetupTestDependencies(
	ctx context.Context,
	testDep *TestDependencies,
	logger *logrus.Logger,
) (*TestDependencies, error) {
	// Setup PostgreSQL
	testDep, err := SetupPgDependency(ctx, testDep, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to setup PostgreSQL dependency: %w", err)
	}

	// Setup Kafka
	testDep, err = SetupKafkaDependency(ctx, testDep, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to setup Kafka dependency: %w", err)
	}

	return testDep, nil
}

func SetupPgDependency(
	ctx context.Context,
	testDep *TestDependencies,
	logger *logrus.Logger,
) (*TestDependencies, error) {
	// Setup PostgreSQL
	pgContainer, err := postgres.Run(
		ctx,
		"postgres:15",
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Name: "test-postgres",
			},
			Reuse: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}
	logger.Info("PostgreSQL container started successfully")

	testDep.PostgresContainer = pgContainer
	return testDep, nil
}

func SetupKafkaDependency(ctx context.Context, testDep *TestDependencies, logger *logrus.Logger) (*TestDependencies, error) {
	// Setup Kafka
	kafkaContainer, err := kafka.Run(
		ctx,
		"confluentinc/cp-kafka:7.8.0",
		kafka.WithClusterID("test-cluster"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Kafka Server started").
				WithOccurrence(1).
				WithStartupTimeout(5*time.Second)),
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Name: "test-kafka",
			},
			Reuse: true,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start kafka container: %w", err)
	}
	logger.Info("Kafka container started successfully")

	testDep.KafkaContainer = kafkaContainer
	return testDep, nil
}

func (td *TestDependencies) Teardown(ctx context.Context) {
	if td.PostgresContainer != nil {
		if err := td.PostgresContainer.Terminate(ctx); err != nil {
			log.Printf("failed to terminate postgres container: %v", err)
		}
	}
	if td.KafkaContainer != nil {
		if err := td.KafkaContainer.Terminate(ctx); err != nil {
			log.Printf("failed to terminate kafka container: %v", err)
		}
	}
}
