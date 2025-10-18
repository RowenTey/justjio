package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	config "github.com/RowenTey/JustJio/server/api/config"

	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
)

func ConnectDB(conf *config.Config, environment string, logger *logrus.Logger) *gorm.DB {
	dbLogger := logger.WithFields(logrus.Fields{"service": "Database"})

	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		conf.DB.Username,
		conf.DB.Password,
		conf.DB.Host,
		conf.DB.Port,
		conf.DB.Database,
	)

	gormConfig := &gorm.Config{
		TranslateError: true,
		Logger:         nil,
	}
	if environment == "dev" || environment == "staging" {
		gormConfig.Logger = gormLogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			gormLogger.Config{
				SlowThreshold: time.Second,     // slow SQL threshold
				LogLevel:      gormLogger.Info, // show all SQL
				Colorful:      true,
			},
		)
	}

	dbConn, err := gorm.Open(gormPostgres.Open(dsn), gormConfig)
	if err != nil {
		dbLogger.Error("Failed to connect to database!")
		dbLogger.Fatal(err)
	}
	dbLogger.Info("Connection opened to database")

	if err := dbConn.Use(otelgorm.NewPlugin(otelgorm.WithDBName(conf.DB.Database))); err != nil {
		dbLogger.Warn("Failed to add OpenTelemetry plugin to GORM: ", err.Error())
	}
	dbLogger.Info("OpenTelemetry tracing enabled for database")

	if err := Migrate(dbConn, "file://migrations"); err != nil {
		dbLogger.Error("Migration failed: ", err.Error())
	}
	dbLogger.Info("Database migrated")

	return dbConn
}

func Migrate(db *gorm.DB, migrationsUrl string) error {
	sqlConn, err := db.DB()
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(sqlConn, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsUrl,
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
