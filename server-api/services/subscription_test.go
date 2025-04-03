package services

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/RowenTey/JustJio/model"
	"github.com/RowenTey/JustJio/util"
)

type SubscriptionServiceTestSuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	subscriptionService *SubscriptionService
}

func TestSubscriptionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionServiceTestSuite))
}

func (s *SubscriptionServiceTestSuite) SetupTest() {
	var err error
	s.DB, s.mock, err = util.SetupTestDB()
	assert.NoError(s.T(), err)

	s.subscriptionService = NewSubscriptionService(s.DB)
}

func (s *SubscriptionServiceTestSuite) AfterTest(_, _ string) {
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *SubscriptionServiceTestSuite) TestCreateSubscription_Success() {
	// arrange
	subscription := &model.Subscription{
		UserID:   1,
		Endpoint: "https://fcm.googleapis.com/fcm/send/example-token",
		P256dh:   "BNcRdreALRFXTkOOUHK1EtK2wtaz5Ry4YfYCA_0QTpQtfZOYo0XAbn1zhRsz1You9yOFCdVV1cVniWhbGrQJq2Q",
		Auth:     "Q2QVd5bPkMEwMKKKv5gVAQ",
	}

	s.mock.ExpectBegin()
	// Use ExpectExec instead of ExpectQuery for INSERT operations that use ExecQuery
	s.mock.ExpectExec(`INSERT INTO "subscriptions"`).
		// Use sqlmock.AnyArg() for the ID since it's generated
		WithArgs(
			sqlmock.AnyArg(), // ID is auto-generated
			subscription.UserID,
			subscription.Endpoint,
			subscription.Auth,
			subscription.P256dh,
		).
		WillReturnResult(sqlmock.NewResult(1, 1)) // For Exec, use WillReturnResult
	s.mock.ExpectCommit()

	// act
	result, err := s.subscriptionService.CreateSubscription(subscription)

	// assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), subscription.UserID, result.UserID)
	assert.Equal(s.T(), subscription.Endpoint, result.Endpoint)
	assert.Equal(s.T(), subscription.P256dh, result.P256dh)
	assert.Equal(s.T(), subscription.Auth, result.Auth)
}

func (s *SubscriptionServiceTestSuite) TestCreateSubscription_Error() {
	// arrange
	subscription := &model.Subscription{
		UserID:   1,
		Endpoint: "https://fcm.googleapis.com/fcm/send/example-token",
		P256dh:   "BNcRdreALRFXTkOOUHK1EtK2wtaz5Ry4YfYCA_0QTpQtfZOYo0XAbn1zhRsz1You9yOFCdVV1cVniWh",
		Auth:     "Q2QVd5bPkMEwMKKKv5gVAQ",
	}

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`INSERT INTO "subscriptions"`).
		WithArgs(
			sqlmock.AnyArg(), // ID is auto-generated
			subscription.UserID,
			subscription.Endpoint,
			subscription.Auth,
			subscription.P256dh,
		).
		WillReturnError(errors.New("database error"))
	s.mock.ExpectRollback()

	// act
	result, err := s.subscriptionService.CreateSubscription(subscription)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionsByUserID_Success() {
	// arrange
	userID := uint(1)

	expectedSubscriptions := []model.Subscription{
		{
			ID:       "1",
			UserID:   userID,
			Endpoint: "https://fcm.googleapis.com/fcm/send/token1",
			P256dh:   "BNcRdreALRFXTkOOUHK1EtK2wtaz5Ry4YfYCA_0QTpQtfZOYo0X1",
			Auth:     "Q2QVd5bPkMEwMKKKv5gV11",
		},
		{
			ID:       "2",
			UserID:   userID,
			Endpoint: "https://fcm.googleapis.com/fcm/send/token2",
			P256dh:   "BNcRdreALRFXTkOOUHK1EtK2wtaz5Ry4YfYCA_0QTpQtfZOYo0X2",
			Auth:     "Q2QVd5bPkMEwMKKKv5gV22",
		},
	}

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "endpoint", "p256dh", "auth",
	})

	for _, sub := range expectedSubscriptions {
		rows.AddRow(
			sub.ID,
			sub.UserID,
			sub.Endpoint,
			sub.P256dh,
			sub.Auth,
		)
	}

	s.mock.ExpectQuery(`SELECT \* FROM "subscriptions" WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	// act
	subscriptions, err := s.subscriptionService.GetSubscriptionsByUserID(userID)

	// assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), subscriptions)
	assert.Equal(s.T(), 2, len(*subscriptions))
	assert.Equal(s.T(), expectedSubscriptions[0].ID, (*subscriptions)[0].ID)
	assert.Equal(s.T(), expectedSubscriptions[0].Endpoint, (*subscriptions)[0].Endpoint)
	assert.Equal(s.T(), expectedSubscriptions[1].ID, (*subscriptions)[1].ID)
	assert.Equal(s.T(), expectedSubscriptions[1].Endpoint, (*subscriptions)[1].Endpoint)
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionsByUserID_EmptyResult() {
	// arrange
	userID := uint(1)

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "endpoint", "p256dh", "auth",
	})

	s.mock.ExpectQuery(`SELECT \* FROM "subscriptions" WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	// act
	subscriptions, err := s.subscriptionService.GetSubscriptionsByUserID(userID)

	// assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), subscriptions)
	assert.Equal(s.T(), 0, len(*subscriptions))
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionsByUserID_DatabaseError() {
	// arrange
	userID := uint(1)

	s.mock.ExpectQuery(`SELECT \* FROM "subscriptions" WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnError(errors.New("database error"))

	// act
	subscriptions, err := s.subscriptionService.GetSubscriptionsByUserID(userID)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), subscriptions)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionsByEndpoint_Success() {
	// arrange
	endpoint := "https://fcm.googleapis.com/fcm/send/example-token"

	expectedSubscription := model.Subscription{
		ID:       "1",
		UserID:   1,
		Endpoint: endpoint,
		P256dh:   "BNcRdreALRFXTkOOUHK1EtK2wtaz5Ry4YfYCA_0QTpQtfZOYo0X1",
		Auth:     "Q2QVd5bPkMEwMKKKv5gV11",
	}

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "endpoint", "p256dh", "auth",
	}).AddRow(
		expectedSubscription.ID,
		expectedSubscription.UserID,
		expectedSubscription.Endpoint,
		expectedSubscription.P256dh,
		expectedSubscription.Auth,
	)

	s.mock.ExpectQuery(`SELECT \* FROM "subscriptions" WHERE endpoint = \$1`).
		WithArgs(endpoint).
		WillReturnRows(rows)

	// act
	subscription, err := s.subscriptionService.GetSubscriptionsByEndpoint(endpoint)

	// assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), subscription)
	assert.Equal(s.T(), expectedSubscription.ID, subscription.ID)
	assert.Equal(s.T(), expectedSubscription.UserID, subscription.UserID)
	assert.Equal(s.T(), expectedSubscription.Endpoint, subscription.Endpoint)
	assert.Equal(s.T(), expectedSubscription.P256dh, subscription.P256dh)
	assert.Equal(s.T(), expectedSubscription.Auth, subscription.Auth)
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionsByEndpoint_NotFound() {
	// arrange
	endpoint := "https://fcm.googleapis.com/fcm/send/nonexistent-token"

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "endpoint", "p256dh", "auth", "created_at",
	})

	s.mock.ExpectQuery(`SELECT \* FROM "subscriptions" WHERE endpoint = \$1`).
		WithArgs(endpoint).
		WillReturnRows(rows)

	// act
	subscription, err := s.subscriptionService.GetSubscriptionsByEndpoint(endpoint)

	// assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), subscription)
	assert.Equal(s.T(), "", subscription.ID) // Empty ID indicates no subscription found
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionsByEndpoint_DatabaseError() {
	// arrange
	endpoint := "https://fcm.googleapis.com/fcm/send/example-token"

	s.mock.ExpectQuery(`SELECT \* FROM "subscriptions" WHERE endpoint = \$1`).
		WithArgs(endpoint).
		WillReturnError(errors.New("database error"))

	// act
	subscription, err := s.subscriptionService.GetSubscriptionsByEndpoint(endpoint)

	// assert
	assert.Error(s.T(), err)
	assert.Nil(s.T(), subscription)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *SubscriptionServiceTestSuite) TestDeleteSubscription_Success() {
	// arrange
	subID := "1"

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`DELETE FROM "subscriptions" WHERE id = \$1`).
		WithArgs(subID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	s.mock.ExpectCommit()

	// act
	err := s.subscriptionService.DeleteSubscription(subID)

	// assert
	assert.NoError(s.T(), err)
}

func (s *SubscriptionServiceTestSuite) TestDeleteSubscription_NotFound() {
	// arrange
	subID := "999"

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`DELETE FROM "subscriptions" WHERE id = \$1`).
		WithArgs(subID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	s.mock.ExpectCommit()

	// act
	err := s.subscriptionService.DeleteSubscription(subID)

	// assert
	// No error should be returned even if no record was deleted
	assert.NoError(s.T(), err)
}

func (s *SubscriptionServiceTestSuite) TestDeleteSubscription_DatabaseError() {
	// arrange
	subID := "1"

	s.mock.ExpectBegin()
	s.mock.ExpectExec(`DELETE FROM "subscriptions" WHERE id = \$1`).
		WithArgs(subID).
		WillReturnError(errors.New("database error"))
	s.mock.ExpectRollback()

	// act
	err := s.subscriptionService.DeleteSubscription(subID)

	// assert
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *SubscriptionServiceTestSuite) TestNewWebPushSubscriptionObj() {
	// arrange
	subscription := &model.Subscription{
		UserID:   1,
		Endpoint: "https://fcm.googleapis.com/fcm/send/example-token",
		P256dh:   "BNcRdreALRFXTkOOUHK1EtK2wtaz5Ry4YfYCA_0QTpQtfZOYo0XAbn1zhRsz1You9yOFCdVV1cVniWh",
		Auth:     "Q2QVd5bPkMEwMKKKv5gVAQ",
	}

	// act
	webpushObj := s.subscriptionService.NewWebPushSubscriptionObj(subscription)

	// assert
	assert.NotNil(s.T(), webpushObj)
	assert.Equal(s.T(), subscription.Endpoint, webpushObj.Endpoint)
	assert.Equal(s.T(), subscription.P256dh, webpushObj.Keys.P256dh)
	assert.Equal(s.T(), subscription.Auth, webpushObj.Keys.Auth)
	assert.IsType(s.T(), &webpush.Subscription{}, webpushObj)
}
