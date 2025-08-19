package tests

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/docker/go-connections/nat"
	"github.com/golang-jwt/jwt"
	postgresTc "github.com/testcontainers/testcontainers-go/modules/postgres"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Custom matcher for float64 values
type FloatMatcher struct {
	Expected float64
	Epsilon  float64
}

func (m FloatMatcher) Match(v driver.Value) bool {
	actual, ok := v.(float64)
	if !ok {
		return false
	}
	return math.Abs(actual-m.Expected) < m.Epsilon
}

func SetupTestDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	var (
		db  *sql.DB
		err error
	)

	// Create a mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	// Create a new logger for the gorm DB instance
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,       // Don't include params in the SQL log
			Colorful:                  false,       // Disable color
		},
	)

	// Set up the GORM connection to the mock database
	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})
	gormDB, err := gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		return nil, nil, err
	}

	return gormDB, mock, nil
}

// Helper function to generate JWT token for testing
func GenerateTestToken(userID uint, username, email, jwtSecret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"email":    email,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}
	return t, nil
}

func CreateAndConnectToTestDb(
	ctx context.Context,
	postgresContainer *postgresTc.PostgresContainer,
	dbName string,
) (*gorm.DB, error) {
	pgConnStr, err := postgresContainer.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	// Get mapped port
	mappedPort, err := postgresContainer.MappedPort(ctx, nat.Port("5432/tcp"))
	if err != nil {
		return nil, err
	}

	// Initialize database
	db, err := database.InitTestDB(pgConnStr)
	if err != nil {
		return nil, err
	}

	// Create test database
	err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName)).Error
	if err != nil {
		return nil, err
	}

	// Connect to test database
	db, err = database.InitTestDB(fmt.Sprintf(
		"postgres://postgres:postgres@localhost:%s/%s?", mappedPort.Port(), dbName))
	if err != nil {
		return nil, err
	}

	// Run migrations
	err = database.Migrate(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}
