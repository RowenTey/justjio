package config

import (
	"os"
	// automatically load .env
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Port             string
	JwtSecret        string
	AdminEmail       string
	Smtp2goApiKey    string
	GoogleMapsApiKey string
	AllowedOrigins   string
	DB               PostgresConfig
	Kafka            KafkaConfig
	Vapid            VapidConfig
	GoogleOauth      GoogleOauthConfig
}

type PostgresConfig struct {
	Username string
	Password string
	Host     string
	Port     string
	Database string
}

type KafkaConfig struct {
	Host        string
	Port        string
	TopicPrefix string
}

type VapidConfig struct {
	Email      string
	PublicKey  string
	PrivateKey string
}

type GoogleOauthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:             os.Getenv("PORT"),
		JwtSecret:        os.Getenv("JWT_SECRET"),
		AdminEmail:       os.Getenv("ADMIN_EMAIL"),
		Smtp2goApiKey:    os.Getenv("SMTP2GO_API_KEY"),
		GoogleMapsApiKey: os.Getenv("GOOGLE_MAPS_API_KEY"),
		AllowedOrigins:   os.Getenv("ALLOWED_ORIGINS"),
		DB: PostgresConfig{
			Username: os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			Host:     os.Getenv("POSTGRES_HOST"),
			Port:     os.Getenv("POSTGRES_PORT"),
			Database: os.Getenv("POSTGRES_DB"),
		},
		Kafka: KafkaConfig{
			Host:        os.Getenv("KAFKA_HOST"),
			Port:        os.Getenv("KAFKA_PORT"),
			TopicPrefix: os.Getenv("KAFKA_TOPIC_PREFIX"),
		},
		Vapid: VapidConfig{
			Email:      os.Getenv("VAPID_EMAIL"),
			PublicKey:  os.Getenv("VAPID_PUBLIC_KEY"),
			PrivateKey: os.Getenv("VAPID_PRIVATE_KEY"),
		},
		GoogleOauth: GoogleOauthConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		},
	}
	return cfg, nil
}
