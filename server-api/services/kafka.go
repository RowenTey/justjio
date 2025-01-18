package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	model_kafka "github.com/RowenTey/JustJio/model/kafka"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaService struct {
	Producer *kafka.Producer
	Admin    *kafka.AdminClient
}

func NewKafkaService(bootstrapServers string) (*KafkaService, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": bootstrapServers})
	if err != nil {
		return nil, err
	}

	a, err := kafka.NewAdminClientFromProducer(p)
	if err != nil {
		return nil, err
	}

	return &KafkaService{
		Producer: p,
		Admin:    a,
	}, nil
}

func (ks *KafkaService) CreateTopic(topic string) error {
	topicSpec := kafka.TopicSpecification{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	_, err := ks.Admin.CreateTopics(context.Background(), []kafka.TopicSpecification{topicSpec})
	if err != nil {
		return err
	}

	return nil
}

func (ks *KafkaService) BroadcastMessage(userIds *[]string, message model_kafka.KafkaMessage) error {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errors := make(chan error)

	for _, userId := range *userIds {
		wg.Add(1)
		go func(userId string) {
			defer wg.Done()
			channel := fmt.Sprintf("user-%s", userId)

			if err := ks.PublishMessage(channel, string(messageJSON)); err != nil {
				errors <- err
			}
		}(userId)
	}

	// Close the errors channel after all goroutines finish
	go func() {
		wg.Wait()
		close(errors)
	}()

	// Check if any errors occurred in the goroutines
	// BLOCKING until all goroutines finish
	for err := range errors {
		if err != nil {
			log.Printf("Error publishing message: %v", err)
			return err
		}
	}

	return nil
}

func (ks *KafkaService) PublishMessage(topic string, message string) error {
	deliveryChan := make(chan kafka.Event)

	err := ks.Producer.Produce(&kafka.Message{
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

func (ks *KafkaService) Close() {
	// Flush and close the producer and the events channel
	for ks.Producer.Flush(10000) > 0 {
		fmt.Print("Still waiting to flush outstanding messages\n")
	}
	ks.Producer.Close()
}
