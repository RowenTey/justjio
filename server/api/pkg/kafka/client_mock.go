package kafka

import (
	"github.com/stretchr/testify/mock"
)

type MockKafkaClient struct {
	mock.Mock
}

func (m *MockKafkaClient) CreateTopic(topic string) error {
	args := m.Called(topic)
	return args.Error(0)
}

func (m *MockKafkaClient) BroadcastMessage(userIds []string, message KafkaMessage) error {
	args := m.Called(userIds, message)
	return args.Error(0)
}

func (m *MockKafkaClient) PublishMessage(topic string, message string) error {
	args := m.Called(topic, message)
	return args.Error(0)
}

func (m *MockKafkaClient) Close() {
	m.Called()
}
