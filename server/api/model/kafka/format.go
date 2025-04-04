package model_kafka

type KafkaMessage struct {
	MsgType string      `json:"type"`
	Data    interface{} `json:"data"`
}
