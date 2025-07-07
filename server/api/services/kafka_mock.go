package services

import (
	kafkaModel "github.com/RowenTey/JustJio/server/api/model/kafka"
	"github.com/stretchr/testify/mock"
)

type MockKafkaService struct {
	mock.Mock
}

func (m *MockKafkaService) CreateTopic(topic string) error {
	args := m.Called(topic)
	return args.Error(0)
}

func (m *MockKafkaService) BroadcastMessage(userIds *[]string, message kafkaModel.KafkaMessage) error {
	args := m.Called(userIds, message)
	return args.Error(0)
}

func (m *MockKafkaService) PublishMessage(topic string, message string) error {
	args := m.Called(topic, message)
	return args.Error(0)
}

func (m *MockKafkaService) Close() {
	m.Called()
}
