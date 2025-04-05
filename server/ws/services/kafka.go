package services

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/ws/utils"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaService struct {
	Consumer *kafka.Consumer
	Config   *kafka.ConfigMap
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *log.Entry
}

func GetUserChannel(userId, env string) string {
	channel := fmt.Sprintf("user-%s", userId)
	if env == "dev" || env == "staging" {
		channel = fmt.Sprintf("%s-%s", env, channel)
	}
	channel = fmt.Sprintf("%s-%s", utils.Config("KAFKA_TOPIC_PREFIX"), channel)
	return channel
}

// Creates a new KafkaService instance
func NewKafkaService(brokers, groupId string) (*KafkaService, error) {
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

	ctx, cancel := context.WithCancel(context.Background())

	return &KafkaService{
		Consumer: consumer,
		Config:   config,
		ctx:      ctx,
		cancel:   cancel,
		logger:   log.WithField("service", "Kafka"),
	}, nil
}

// Subscribes to a list of Kafka topics
func (s *KafkaService) Subscribe(topics []string) error {
	err := s.Consumer.SubscribeTopics(topics, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topics: %w", err)
	}
	s.logger.Infof("Subscribed to topics: %v\n", topics)
	return nil
}

// Unsubscribes from Kafka topics
func (s *KafkaService) Unsubscribe() error {
	// signal to stop consuming messages
	s.cancel()

	err := s.Consumer.Unsubscribe()
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from topics: %w", err)
	}
	s.logger.Info("Unsubscribed from topics")
	return nil
}

// Consumes messages from subscribed topics in a loop
func (s *KafkaService) ConsumeMessages(handler func(msg kafka.Message)) {
	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Stopping message consumption")
			return
		default:
			msg, err := s.Consumer.ReadMessage(1)
			if err != nil && err.(kafka.Error).Code() == kafka.ErrTimedOut {
				continue
			} else if err != nil {
				s.logger.Printf("Failed to consume message: %s", err.Error())
				return
			}
			s.logger.Debugf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
			handler(*msg)
		}
	}
}

func (s *KafkaService) Close() {
	s.logger.Info("Closing Kafka client")
	s.Consumer.Close()
}
