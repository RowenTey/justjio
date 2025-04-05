package database

import (
	log "github.com/sirupsen/logrus"

	config "github.com/RowenTey/JustJio/server/api/config"
	model "github.com/RowenTey/JustJio/server/api/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is a global variable that holds the connection to the database
var DB *gorm.DB

func ConnectDB() {
	// define error here to prevent overshadowing the global DB
	var err error

	logger := log.WithFields(log.Fields{"service": "Database"})

	dsn := config.Config("DSN")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		// SkipDefaultTransaction: true,
	})
	if err != nil {
		logger.Error("Failed to connect to database")
		logger.Fatal(err)
	}
	logger.Info("Connection opened to database")

	err = Migrate(DB)
	if err != nil {
		logger.Error("Migration failed: ", err.Error())
	}
	logger.Info("Database migrated")
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{},
		&model.FriendRequest{},
		&model.Room{},
		&model.RoomInvite{},
		&model.Bill{},
		&model.Consolidation{},
		&model.Transaction{},
		&model.Message{},
		&model.Notification{},
	); err != nil {
		return err
	}
	return nil
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
