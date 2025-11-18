package database

import (
	"time"

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

	// Configure connection pool for better performance
	sqlDB, err := DB.DB()
	if err != nil {
		logger.Error("Failed to get database instance")
		logger.Fatal(err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)                           // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(100)                          // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(time.Hour * 2)             // Maximum lifetime of a connection (2 hours)
	sqlDB.SetConnMaxIdleTime(time.Minute * 10)          // Maximum idle time for a connection

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
		&model.Subscription{},
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
