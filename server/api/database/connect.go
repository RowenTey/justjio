package database

import (
	"fmt"

	"github.com/sirupsen/logrus"

	config "github.com/RowenTey/JustJio/server/api/config"

	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
	dbConn, err := gorm.Open(gormPostgres.Open(dsn), &gorm.Config{
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
	sqlConn, err := db.DB()
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(sqlConn, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver)
	if err != nil {
		return err
	}

	return m.Up()
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
	return gorm.Open(gormPostgres.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
}
