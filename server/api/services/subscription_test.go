package services

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/tests"
)

type SubscriptionServiceTestSuite struct {
	suite.Suite
	DB   *gorm.DB
	mock sqlmock.Sqlmock

	subscriptionService *SubscriptionService

	endpoint string
	p256dh   string
	auth     string
	subCols  []string
}

func TestSubscriptionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionServiceTestSuite))
}

func (s *SubscriptionServiceTestSuite) SetupTest() {
	var err error
	s.DB, s.mock, err = tests.SetupTestDB()
	assert.NoError(s.T(), err)

	s.subscriptionService = NewSubscriptionService(s.DB)

	s.endpoint = "https://fcm.googleapis.com/fcm/send/token1"
	s.p256dh = "BNcRdreALRFXTkOOUHK1EtK2wtaz5Ry4YfYCA_0QTpQtfZOYo0X1"
	s.auth = "Q2QVd5bPkMEwMKKKv5gV11"
	s.subCols = []string{
		"id", "user_id", "endpoint", "p256dh", "auth",
	}
}

func (s *SubscriptionServiceTestSuite) AfterTest(_, _ string) {
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *SubscriptionServiceTestSuite) TestCreateSubscription_Success() {
	// arrange
	subscription := tests.CreateTestSubscription(1, "1", s.endpoint, s.p256dh, s.auth)

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
	subscription := tests.CreateTestSubscription(1, "1", s.endpoint, s.p256dh, s.auth)

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
		*tests.CreateTestSubscription(userID, "1", s.endpoint, s.p256dh, s.auth),
		*tests.CreateTestSubscription(userID, "2",
			"https://fcm.googleapis.com/fcm/send/token2",
			"BNcRdreALRFXTkOOUHK1EtK2wtaz5Ry4YfYCA_0QTpQtfZOYo0X2",
			"Q2QVd5bPkMEwMKKKv5gV22",
		),
	}

	rows := sqlmock.NewRows(s.subCols)
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
	for i, sub := range expectedSubscriptions {
		assertSubscriptionEqual(s.T(), &sub, &(*subscriptions)[i])
	}
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionsByUserID_EmptyResult() {
	// arrange
	userID := uint(1)

	rows := sqlmock.NewRows(s.subCols)

	s.mock.ExpectQuery(`SELECT \* FROM "subscriptions" WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(rows)

	// act
	subscriptions, err := s.subscriptionService.GetSubscriptionsByUserID(userID)

	// assert
	tests.AssertNoErrAndNotNil(s.T(), err, subscriptions)
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
	tests.AssertErrAndNil(s.T(), err, subscriptions)
	assert.Contains(s.T(), err.Error(), "database error")
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionsByEndpoint_Success() {
	// arrange
	expectedSubscription := tests.CreateTestSubscription(1, "1", s.endpoint, s.p256dh, s.auth)

	rows := sqlmock.NewRows(s.subCols).AddRow(
		expectedSubscription.ID,
		expectedSubscription.UserID,
		expectedSubscription.Endpoint,
		expectedSubscription.P256dh,
		expectedSubscription.Auth,
	)

	s.mock.ExpectQuery(`SELECT \* FROM "subscriptions" WHERE endpoint = \$1`).
		WithArgs(s.endpoint).
		WillReturnRows(rows)

	// act
	subscription, err := s.subscriptionService.GetSubscriptionsByEndpoint(s.endpoint)

	// assert
	tests.AssertNoErrAndNotNil(s.T(), err, subscription)
	assertSubscriptionEqual(s.T(), expectedSubscription, subscription)
}

func (s *SubscriptionServiceTestSuite) TestGetSubscriptionsByEndpoint_NotFound() {
	// arrange
	endpoint := "https://fcm.googleapis.com/fcm/send/nonexistent-token"

	rows := sqlmock.NewRows(append(s.subCols, "created_at"))

	s.mock.ExpectQuery(`SELECT \* FROM "subscriptions" WHERE endpoint = \$1`).
		WithArgs(endpoint).
		WillReturnRows(rows)

	// act
	subscription, err := s.subscriptionService.GetSubscriptionsByEndpoint(endpoint)

	// assert
	tests.AssertNoErrAndNotNil(s.T(), err, subscription)
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
	tests.AssertErrAndNil(s.T(), err, subscription)
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
	subscription := tests.CreateTestSubscription(1, "1", s.endpoint, s.p256dh, s.auth)

	// act
	webpushObj := s.subscriptionService.NewWebPushSubscriptionObj(subscription)

	// assert
	assert.NotNil(s.T(), webpushObj)
	assert.Equal(s.T(), subscription.Endpoint, webpushObj.Endpoint)
	assert.Equal(s.T(), subscription.P256dh, webpushObj.Keys.P256dh)
	assert.Equal(s.T(), subscription.Auth, webpushObj.Keys.Auth)
	assert.IsType(s.T(), &webpush.Subscription{}, webpushObj)
}

func assertSubscriptionEqual(t assert.TestingT, expected, actual *model.Subscription) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.UserID, actual.UserID)
	assert.Equal(t, expected.Endpoint, actual.Endpoint)
	assert.Equal(t, expected.P256dh, actual.P256dh)
	assert.Equal(t, expected.Auth, actual.Auth)
}
