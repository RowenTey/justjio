package kafka

type KafkaClient interface {
	CreateTopic(topic string) error
	BroadcastMessage(userIds []string, message KafkaMessage) error
	PublishMessage(topic string, message string) error
	Close()
}
