package kafka

import "context"

type KafkaClient interface {
	CreateTopic(topic string) error
	BroadcastMessage(ctx context.Context, userIds []string, message KafkaMessage) error
	PublishMessage(ctx context.Context, topic string, messageJson []byte) error
	Close()
}
