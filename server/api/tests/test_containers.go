package tests

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDependencies struct {
	PostgresContainer *postgres.PostgresContainer
	KafkaContainer    *kafka.KafkaContainer
}

func SetupTestDependencies(ctx context.Context) (*TestDependencies, error) {
	// Setup PostgreSQL
	pgContainer, err := postgres.Run(ctx,
		"postgres:15",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	// Setup Kafka
	kafkaContainer, err := kafka.Run(ctx,
		"confluentinc/cp-kafka:7.8.0",
		kafka.WithClusterID("test-cluster"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start kafka container: %w", err)
	}

	return &TestDependencies{
		PostgresContainer: pgContainer,
		KafkaContainer:    kafkaContainer,
	}, nil
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
