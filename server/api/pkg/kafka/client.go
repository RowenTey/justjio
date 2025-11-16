package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/pkg/config"

	confluentKafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type kafkaClient struct {
	producer    *confluentKafka.Producer
	admin       *confluentKafka.AdminClient
	env         string
	topicPrefix string
	logger      *logrus.Entry
}

func NewKafkaClient(conf *config.Config, logger *logrus.Logger, env string) (KafkaClient, error) {
	bootstrapServers := fmt.Sprintf("%s:%s", conf.Kafka.Host, conf.Kafka.Port)
	p, err := confluentKafka.NewProducer(&confluentKafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
	})
	if err != nil {
		return nil, err
	}

	a, err := confluentKafka.NewAdminClientFromProducer(p)
	if err != nil {
		return nil, err
	}

	return &kafkaClient{
		producer:    p,
		admin:       a,
		env:         env,
		topicPrefix: conf.Kafka.TopicPrefix,
		logger:      logger.WithFields(logrus.Fields{"component": "KafkaClient"}),
	}, nil
}

func (kc *kafkaClient) CreateTopic(topic string) error {
	topic = kc.getFormattedTopic(topic)
	topicSpec := confluentKafka.TopicSpecification{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	if _, err := kc.admin.CreateTopics(
		context.Background(),
		[]confluentKafka.TopicSpecification{topicSpec},
	); err != nil {
		return err
	}

	return nil
}

func (kc *kafkaClient) BroadcastMessage(userIds []string, message KafkaMessage) error {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errorChan := make(chan error)

	// TODO: Need to think more about this
	for _, userId := range userIds {
		wg.Add(1)
		go func(userId string) {
			defer wg.Done()

			channel := fmt.Sprintf("user-%s", userId)
			channel = kc.getFormattedTopic(channel)

			if err := kc.PublishMessage(channel, string(messageJSON)); err != nil {
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
			kc.logger.Errorf("Error publishing message: %v", err)
			allErrors = errors.Join(allErrors, err)
		}
	}

	return allErrors
}

func (kc *kafkaClient) PublishMessage(topic string, message string) error {
	deliveryChan := make(chan confluentKafka.Event)

	if err := kc.producer.Produce(&confluentKafka.Message{
		TopicPartition: confluentKafka.TopicPartition{
			Topic:     &topic,
			Partition: confluentKafka.PartitionAny,
		},
		Value: []byte(message),
	}, deliveryChan); err != nil {
		return err
	}

	// Block until the message is delivered
	e := <-deliveryChan
	m := e.(*confluentKafka.Message)
	if m.TopicPartition.Error != nil {
		return m.TopicPartition.Error
	}

	close(deliveryChan)
	return nil
}

func (kc *kafkaClient) Close() {
	unflushed := kc.producer.Flush(10000)
	if unflushed > 0 {
		kc.logger.Warnf("Unflushed messages: %d\n", unflushed)
	}
	kc.producer.Close()
}

func (kc *kafkaClient) getFormattedTopic(topic string) string {
	if kc.env != "production" {
		topic = fmt.Sprintf("%s-%s", kc.env, topic)
	}
	topic = fmt.Sprintf("%s-%s", kc.topicPrefix, topic)
	return topic
}
