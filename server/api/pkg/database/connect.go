package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/pkg/app"

	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
)

func ConnectDB(ctx *app.Context) *gorm.DB {
	logger := ctx.Logger.WithFields(logrus.Fields{"service": "Database"})
	environment := ctx.Config.Environment
	config := ctx.Config.DB

	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	gormConfig := &gorm.Config{
		TranslateError: true,
		PrepareStmt:    true,
		Logger:         nil,
	}
	if environment != "production" {
		gormConfig.Logger = gormLogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			gormLogger.Config{
				SlowThreshold: time.Second,     // slow SQL threshold
				LogLevel:      gormLogger.Info, // show all SQL
				Colorful:      true,
			},
		)
	}

	conn, err := gorm.Open(gormPostgres.Open(dsn), gormConfig)
	if err != nil {
		logger.Error("Failed to connect to database!")
		logger.Fatal(err)
	}
	logger.Info("Connection opened to database")

	// Set connection pool settings
	sqlDb, err := conn.DB()
	if err != nil {
		logger.Error("Failed to get database instance from GORM!")
		logger.Fatal(err)
	}
	sqlDb.SetMaxIdleConns(10)
	sqlDb.SetMaxOpenConns(50)
	sqlDb.SetConnMaxLifetime(10 * time.Minute)
	sqlDb.SetConnMaxIdleTime(5 * time.Minute)

	if err := conn.Use(otelgorm.NewPlugin(otelgorm.WithDBName(config.Database))); err != nil {
		logger.Warn("Failed to add OpenTelemetry plugin to GORM: ", err.Error())
	}
	logger.Info("OpenTelemetry tracing enabled for database")

	if environment != "production" {
		if err := Migrate(conn, "file://migrations"); err != nil {
			logger.Error("Migration failed: ", err.Error())
		}
		logger.Debug("Database migrated")
	}

	return conn
}

func Migrate(db *gorm.DB, migrationsUrl string) error {
	conn, err := db.DB()
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(conn, &postgres.Config{})
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

	// Run down migrations first to avoid issues with dirty database state
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
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
