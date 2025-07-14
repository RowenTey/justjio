package tests

import (
	"database/sql"
	"database/sql/driver"
	"log"
	"math"
	"os"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt"
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
