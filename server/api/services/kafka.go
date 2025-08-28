package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/config"
	modelKafka "github.com/RowenTey/JustJio/server/api/dto/kafka"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaService interface {
	CreateTopic(topic string) error
	BroadcastMessage(userIds *[]string, message modelKafka.KafkaMessage) error
	PublishMessage(topic string, message string) error
	Close()
}

type kafkaService struct {
	producer    *kafka.Producer
	admin       *kafka.AdminClient
	env         string
	topicPrefix string
	logger      *logrus.Entry
}

func NewKafkaService(conf *config.Config, logger *logrus.Logger, env string) (KafkaService, error) {
	bootstrapServers := fmt.Sprintf("%s:%s", conf.Kafka.Host, conf.Kafka.Port)
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
	})
	if err != nil {
		return nil, err
	}

	a, err := kafka.NewAdminClientFromProducer(p)
	if err != nil {
		return nil, err
	}

	return &kafkaService{
		producer:    p,
		admin:       a,
		env:         env,
		topicPrefix: conf.Kafka.TopicPrefix,
		logger:      logger.WithFields(logrus.Fields{"service": "KafkaService"}),
	}, nil
}

func (ks *kafkaService) CreateTopic(topic string) error {
	topic = ks.getFormattedTopic(topic)
	topicSpec := kafka.TopicSpecification{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	_, err := ks.admin.CreateTopics(context.Background(), []kafka.TopicSpecification{topicSpec})
	if err != nil {
		return err
	}

	return nil
}

func (ks *kafkaService) BroadcastMessage(userIds *[]string, message modelKafka.KafkaMessage) error {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errorChan := make(chan error)

	// TODO: Need to think more about this
	for _, userId := range *userIds {
		wg.Add(1)
		go func(userId string) {
			defer wg.Done()

			channel := fmt.Sprintf("user-%s", userId)
			channel = ks.getFormattedTopic(channel)

			if err := ks.PublishMessage(channel, string(messageJSON)); err != nil {
				errorChan <- err
			}
		}(userId)
	}

	// Close the errors channel after all goroutines finish
	go func() {
		wg.Wait()
		close(errorChan)
	}()

	// Check if any errors occurred in the goroutines
	// BLOCKING until all goroutines finish
	var allErrors error = nil
	for err := range errorChan {
		if err != nil {
			ks.logger.Errorf("Error publishing message: %v", err)
			allErrors = errors.Join(allErrors, err)
		}
	}

	return allErrors
}

func (ks *kafkaService) PublishMessage(topic string, message string) error {
	deliveryChan := make(chan kafka.Event)

	err := ks.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, deliveryChan)
	if err != nil {
		return err
	}

	// Block until the message is delivered
	e := <-deliveryChan
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		return m.TopicPartition.Error
	}

	close(deliveryChan)
	return nil
}

func (ks *kafkaService) Close() {
	// Flush and close the producer and the events channel
	unflushed := ks.producer.Flush(10000)
	ks.logger.Warnf("Unflushed messages: %d\n", unflushed)

	ks.producer.Close()
}

func (ks *kafkaService) getFormattedTopic(topic string) string {
	if ks.env == "dev" || ks.env == "staging" {
		topic = fmt.Sprintf("%s-%s", ks.env, topic)
	}
	topic = fmt.Sprintf("%s-%s", ks.topicPrefix, topic)
	return topic
}
