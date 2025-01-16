package services

import (
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaService struct {
	Consumer *kafka.Consumer
	Config   *kafka.ConfigMap
}

func GetChannel(userId string) string {
	return fmt.Sprintf("user-%s", userId)
}

// Creates a new KafkaService instance
func NewKafkaService(brokers string, groupId string) (*KafkaService, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers":       brokers,
		"group.id":                groupId,
		"auto.offset.reset":       "earliest",
		"enable.auto.commit":      true,
		"session.timeout.ms":      6000,
		"auto.commit.interval.ms": 5000,
	}

	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, err
	}

	return &KafkaService{Consumer: consumer, Config: config}, nil
}

// Subscribes to a list of Kafka topics
func (s *KafkaService) Subscribe(topics []string) error {
	err := s.Consumer.SubscribeTopics(topics, nil)
	if err != nil {
		return fmt.Errorf("Failed to subscribe to topics: %w", err)
	}
	log.Printf("Subscribed to topics: %v\n", topics)
	return nil
}

// Unsubscribes from Kafka topics
func (s *KafkaService) Unsubscribe() error {
	err := s.Consumer.Unsubscribe()
	if err != nil {
		return fmt.Errorf("Failed to unsubscribe from topics: %w", err)
	}
	log.Println("Unsubscribed from topics")
	return nil
}

// Consumes messages from subscribed topics in a loop
func (s *KafkaService) ConsumeMessages(handler func(msg kafka.Message)) error {
	for {
		msg, err := s.Consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("Error consuming messages: %v\n", err)
			return nil
		}
		handler(*msg)
	}
}

func (s *KafkaService) Close() {
	log.Println("Closing Kafka client")
	s.Consumer.Close()
}
