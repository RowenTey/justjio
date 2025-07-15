package utils

import (
	"os"
	// automatically load .env
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Port           string
	JwtSecret      string
	AllowedOrigins string
	Kafka          KafkaConfig
}

type KafkaConfig struct {
	Host        string
	Port        string
	TopicPrefix string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:           os.Getenv("PORT"),
		JwtSecret:      os.Getenv("JWT_SECRET"),
		AllowedOrigins: os.Getenv("ALLOWED_ORIGINS"),
		Kafka: KafkaConfig{
			Host:        os.Getenv("KAFKA_HOST"),
			Port:        os.Getenv("KAFKA_PORT"),
			TopicPrefix: os.Getenv("KAFKA_TOPIC_PREFIX"),
		},
	}
	return cfg, nil
}
