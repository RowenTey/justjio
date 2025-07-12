package tests

import (
	"database/sql"
	"database/sql/driver"
	"log"
	"math"
	"os"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
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

func AssertErrAndNil(t assert.TestingT, err error, obj any) {
	assert.Error(t, err)
	assert.Nil(t, obj)
}

func AssertNoErrAndNotNil(t assert.TestingT, err error, obj any) {
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func CreateTestUser(id uint, username, email string) *model.User {
	now := time.Now()
	return &model.User{
		ID:           id,
		Username:     username,
		Email:        email,
		Password:     "hashedpassword",
		PictureUrl:   "https://default-image.jpg",
		IsEmailValid: true,
		IsOnline:     false,
		LastSeen:     now,
		RegisteredAt: now,
		UpdatedAt:    now,
	}
}

func CreateTestRoom(id, name string, hostId uint) *model.Room {
	now := time.Now()
	return &model.Room{
		ID:             id,
		Name:           name,
		CreatedAt:      now,
		HostID:         hostId,
		AttendeesCount: 1,
		IsClosed:       false,
	}
}

func CreateTestRoomInvite(id uint, roomId string, userId, inviterId uint) *model.RoomInvite {
	now := time.Now()
	return &model.RoomInvite{
		ID:        id,
		RoomID:    roomId,
		UserID:    userId,
		InviterID: inviterId,
		Status:    "pending",
		CreatedAt: now,
	}
}

func CreateTestSubscription(userID uint, id, endpoint, p256dh, auth string) *model.Subscription {
	return &model.Subscription{
		ID:       id,
		UserID:   userID,
		Endpoint: endpoint,
		P256dh:   p256dh,
		Auth:     auth,
	}
}

func CreateTestTransaction(id, consolidationID, payerID, payeeID uint, amount float32) *model.Transaction {
	return &model.Transaction{
		ID:              id,
		ConsolidationID: consolidationID,
		PayerID:         payerID,
		PayeeID:         payeeID,
		Amount:          amount,
		IsPaid:          false,
	}
}
