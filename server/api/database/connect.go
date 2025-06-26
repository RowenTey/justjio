package database

import (
	"fmt"

	"github.com/sirupsen/logrus"

	config "github.com/RowenTey/JustJio/server/api/config"
	model "github.com/RowenTey/JustJio/server/api/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB(conf *config.Config, logger *logrus.Logger) *gorm.DB {
	dbLogger := logger.WithFields(logrus.Fields{"service": "Database"})

	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		conf.DB.Username,
		conf.DB.Password,
		conf.DB.Host,
		conf.DB.Port,
		conf.DB.Database,
	)
	dbConn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		dbLogger.Error("Failed to connect to database!")
		dbLogger.Fatal(err)
	}
	dbLogger.Info("Connection opened to database")

	if err := Migrate(dbConn); err != nil {
		dbLogger.Error("Migration failed: ", err.Error())
	}
	dbLogger.Info("Database migrated")

	return dbConn
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.FriendRequest{},
		&model.Room{},
		&model.RoomInvite{},
		&model.Bill{},
		&model.Consolidation{},
		&model.Transaction{},
		&model.Message{},
		&model.Notification{},
		&model.Subscription{},
	)
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func InitTestDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}
