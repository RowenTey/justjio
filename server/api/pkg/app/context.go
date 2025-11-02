package app

import (
	"context"

	"github.com/RowenTey/JustJio/server/api/pkg/config"
	"github.com/RowenTey/JustJio/server/api/pkg/dto/notifications"
	"github.com/RowenTey/JustJio/server/api/pkg/kafka"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Context struct {
	Ctx               context.Context
	Config            *config.Config
	Logger            *logrus.Logger
	DB                *gorm.DB
	Kafka             kafka.KafkaClient
	NotificationsChan chan<- notifications.NotificationData
	App               *fiber.App
}
