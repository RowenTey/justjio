package database

import (
	"log"

	"github.com/RowenTey/JustJio/config"
	"github.com/RowenTey/JustJio/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is a global variable that holds the connection to the database
var DB *gorm.DB

func ConnectDB() {
	// define error here to prevent overshadowing the global DB
	var err error

	dsn := config.Config("DSN")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		// SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Println("[DB] Failed to connect to database")
		log.Fatal(err)
	}
	log.Println("[DB] Connection opened to database")

	err = DB.AutoMigrate(
		&model.User{},
		&model.Room{},
		&model.RoomInvite{},
		&model.Bill{},
		&model.Consolidation{},
		&model.Transaction{},
		&model.Message{},
	)
	if err != nil {
		log.Println("[DB] Migration failed: ", err.Error())
	}
	log.Println("[DB] Database migrated")
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
