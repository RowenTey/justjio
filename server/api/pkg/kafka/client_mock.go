package kafka

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockKafkaClient struct {
	mock.Mock
}

func (m *MockKafkaClient) CreateTopic(topic string) error {
	args := m.Called(topic)
	return args.Error(0)
}

func (m *MockKafkaClient) BroadcastMessage(ctx context.Context, userIds []string, message KafkaMessage) error {
	args := m.Called(ctx, userIds, message)
	return args.Error(0)
}

func (m *MockKafkaClient) PublishMessage(ctx context.Context, topic string, messageJson []byte) error {
	args := m.Called(ctx, topic, messageJson)
	return args.Error(0)
}

func (m *MockKafkaClient) Close() {
	m.Called()
}
