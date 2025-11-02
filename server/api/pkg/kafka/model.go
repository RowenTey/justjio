package kafka

type KafkaMessage struct {
	MsgType string `json:"type"`
	Data    any    `json:"data"`
}

type MessagePayload struct {
	RoomID     string `json:"roomId"`
	SenderID   string `json:"senderId"`
	SenderName string `json:"senderName"`
	Content    string `json:"content"`
	SentAt     string `json:"sentAt"`
}
