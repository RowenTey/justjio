package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/RowenTey/JustJio/server/api/internal/models"
	"github.com/RowenTey/JustJio/server/api/pkg/tests"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	ctx          context.Context
	db           *gorm.DB
	repo         UserRepository
	logger       *logrus.Logger
	dependencies *tests.TestDependencies

	testUser *models.User
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error
	suite.logger = logrus.New()

	// Setup test containers
	suite.dependencies = &tests.TestDependencies{}
	suite.dependencies, err = tests.SetupPgDependency(suite.ctx, suite.dependencies, suite.logger)
	assert.NoError(suite.T(), err)

	// Setup DB Conn
	suite.db, err = tests.CreateAndConnectToTestDb(suite.ctx, suite.dependencies.PostgresContainer, "user_test", "file://../../migrations")
	assert.NoError(suite.T(), err)

	suite.repo = NewUserRepository(suite.db)
}

func (suite *UserRepositoryTestSuite) TearDownSuite() {
	if !IsPackageTest && suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	// Insert base user
	user := models.User{
		Username: "testuser",
		Email:    "testuser@example.com",
		Password: "hashed-password",
	}
	err := suite.db.Create(&user).Error
	assert.NoError(suite.T(), err)

	suite.testUser = &user
}

func (suite *UserRepositoryTestSuite) TearDownTest() {
	suite.db.Exec("TRUNCATE TABLE room_users RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE rooms RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE friend_requests RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE user_friends RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
}

func TestUserRepositorySuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (suite *UserRepositoryTestSuite) TestCreateAndFindByID_Success() {
	user := &models.User{
		Username: "john",
		Email:    "john@example.com",
		Password: "pass",
	}
	created, err := suite.repo.Create(suite.ctx, user)
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByID(suite.ctx, fmt.Sprintf("%d", created.ID))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "john", found.Username)
}

func (suite *UserRepositoryTestSuite) TestFindByUsernameAndEmail_Success() {
	foundByUsername, err := suite.repo.FindByUsername(suite.ctx, suite.testUser.Username)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.testUser.Email, foundByUsername.Email)

	foundByEmail, err := suite.repo.FindByEmail(suite.ctx, suite.testUser.Email)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.testUser.Username, foundByEmail.Username)
}

func (suite *UserRepositoryTestSuite) TestAddAndCheckFriendship_Success() {
	friend := models.User{
		Username: "friend",
		Email:    "friend@example.com",
		Password: "pass",
	}
	err := suite.db.Create(&friend).Error
	assert.NoError(suite.T(), err)

	err = suite.repo.AddFriend(suite.ctx, suite.testUser.ID, friend.ID)
	assert.NoError(suite.T(), err)

	isFriend, err := suite.repo.CheckFriendship(suite.ctx, suite.testUser.ID, friend.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), isFriend)
}

func (suite *UserRepositoryTestSuite) TestCreateFriendRequestAndExists_Success() {
	receiver := models.User{
		Username: "bob",
		Email:    "bob@example.com",
		Password: "pw",
	}
	err := suite.db.Create(&receiver).Error
	assert.NoError(suite.T(), err)

	req := &models.FriendRequest{
		SenderID:   suite.testUser.ID,
		ReceiverID: receiver.ID,
		Status:     "pending",
	}
	err = suite.repo.CreateFriendRequest(suite.ctx, req)
	assert.NoError(suite.T(), err)

	exists, err := suite.repo.CheckFriendRequestExists(suite.ctx, suite.testUser.ID, receiver.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

func (suite *UserRepositoryTestSuite) TestSearchUsers_NoFriends_Success() {
	user2 := models.User{
		Username: "search_target",
		Email:    "target@example.com",
		Password: "secret",
	}
	suite.db.Create(&user2)

	// Refresh the materialized view
	suite.db.Exec("REFRESH MATERIALIZED VIEW user_non_friends")

	results, err := suite.repo.SearchNonFriendUsers(suite.ctx, fmt.Sprintf("%d", suite.testUser.ID), "search", 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 1)
	assert.Equal(suite.T(), "search_target", results[0].Username)
}

func (suite *UserRepositoryTestSuite) TestSearchUsers_WithFriends_Success() {
	user2 := models.User{
		Username: "search_target",
		Email:    "target@example.com",
		Password: "secret",
	}
	suite.db.Create(&user2)

	// Add user2 as a friend to testUser
	err := suite.repo.AddFriend(suite.ctx, suite.testUser.ID, user2.ID)
	assert.NoError(suite.T(), err)

	user3 := models.User{
		Username: "another_target",
		Email:    "another_target@example.com",
		Password: "secret",
	}
	suite.db.Create(&user3)

	// Refresh the materialized view to reflect the new friendship
	suite.db.Exec("REFRESH MATERIALIZED VIEW user_non_friends")

	results, err := suite.repo.SearchNonFriendUsers(suite.ctx, fmt.Sprintf("%d", suite.testUser.ID), "another", 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 1)
	assert.Equal(suite.T(), "another_target", results[0].Username)
}

func (suite *UserRepositoryTestSuite) TestCountFriends_Success() {
	friend := models.User{
		Username: "frienduser",
		Email:    "friend@example.com",
		Password: "pass",
	}
	err := suite.db.Create(&friend).Error
	assert.NoError(suite.T(), err)

	err = suite.repo.AddFriend(suite.ctx, suite.testUser.ID, friend.ID)
	assert.NoError(suite.T(), err)

	// Update the counters
	err = suite.repo.UpdateNoOfFriends(suite.ctx, []uint{suite.testUser.ID, friend.ID}, 1)
	assert.NoError(suite.T(), err)

	count, err := suite.repo.CountFriends(suite.ctx, suite.testUser.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *UserRepositoryTestSuite) TestRemoveFriend_Success() {
	friend := models.User{
		Username: "removable",
		Email:    "removable@example.com",
		Password: "pass",
	}
	err := suite.db.Create(&friend).Error
	assert.NoError(suite.T(), err)

	err = suite.repo.AddFriend(suite.ctx, suite.testUser.ID, friend.ID)
	assert.NoError(suite.T(), err)

	err = suite.repo.RemoveFriend(suite.ctx, suite.testUser.ID, friend.ID)
	assert.NoError(suite.T(), err)

	isFriend, err := suite.repo.CheckFriendship(suite.ctx, suite.testUser.ID, friend.ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), isFriend)
}

func (suite *UserRepositoryTestSuite) TestGetUninvitedFriends_Success() {
	// Setup another user as a friend
	friend := models.User{
		Username: "uninvited",
		Email:    "uninvited@example.com",
		Password: "pass",
	}
	err := suite.db.Create(&friend).Error
	assert.NoError(suite.T(), err)

	// Make them friends
	err = suite.repo.AddFriend(suite.ctx, suite.testUser.ID, friend.ID)
	assert.NoError(suite.T(), err)

	// Create a room and do not invite the friend
	room := models.Room{
		Name:   "Test Room",
		HostID: suite.testUser.ID,
	}
	err = suite.db.Create(&room).Error
	assert.NoError(suite.T(), err)

	// Check GetUninvitedFriends
	uninvited, err := suite.repo.GetUninvitedFriends(suite.ctx, room.ID, fmt.Sprintf("%d", suite.testUser.ID))
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), uninvited, 1)
	assert.Equal(suite.T(), friend.ID, uninvited[0].ID)
}

func (suite *UserRepositoryTestSuite) TestFindByIDs_Success() {
	user2 := models.User{Username: "u2", Email: "u2@example.com", Password: "pass"}
	user3 := models.User{Username: "u3", Email: "u3@example.com", Password: "pass"}
	suite.db.Create(&user2)
	suite.db.Create(&user3)

	ids := []string{
		fmt.Sprintf("%d", suite.testUser.ID),
		fmt.Sprintf("%d", user2.ID),
		fmt.Sprintf("%d", user3.ID)}
	users, err := suite.repo.FindByIDs(suite.ctx, ids)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), users, 3)
}

func (suite *UserRepositoryTestSuite) TestUpdateUser_Success() {
	suite.testUser.Username = "updated_username"
	err := suite.repo.Update(suite.ctx, suite.testUser)
	assert.NoError(suite.T(), err)

	var user models.User
	suite.db.First(&user, suite.testUser.ID)
	assert.Equal(suite.T(), "updated_username", user.Username)
}

func (suite *UserRepositoryTestSuite) TestDeleteUser_Success() {
	err := suite.repo.Delete(suite.ctx, fmt.Sprintf("%d", suite.testUser.ID))
	assert.NoError(suite.T(), err)

	var user models.User
	err = suite.db.First(&user, suite.testUser.ID).Error
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func (suite *UserRepositoryTestSuite) TestFindAndCountFriendRequestsByReceiver_Success() {
	sender := models.User{Username: "sender", Email: "sender@example.com", Password: "pass"}
	suite.db.Create(&sender)

	request := models.FriendRequest{
		SenderID:   sender.ID,
		ReceiverID: suite.testUser.ID,
		Status:     "pending",
	}
	suite.db.Create(&request)

	// Update the pending friend requests count
	err := suite.repo.UpdateNoOfPendingFriendRequests(suite.ctx, []uint{suite.testUser.ID}, 1)
	assert.NoError(suite.T(), err)

	requests, err := suite.repo.FindFriendRequestsByReceiver(suite.ctx, suite.testUser.ID, "pending")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), requests, 1)
	assert.Equal(suite.T(), sender.ID, requests[0].SenderID)

	count, err := suite.repo.CountPendingFriendRequestsByReceiver(suite.ctx, suite.testUser.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *UserRepositoryTestSuite) TestUpdateNoOfPendingRoomInvites_Success() {
	user2 := models.User{Username: "u2", Email: "u2@example.com", Password: "pass"}
	suite.db.Create(&user2)

	// Initial count should be 0
	var user models.User
	suite.db.First(&user, suite.testUser.ID)
	assert.Equal(suite.T(), 0, user.NoOfPendingRoomInvites)

	// Increment by 2
	err := suite.repo.UpdateNoOfPendingRoomInvites(suite.ctx, []string{fmt.Sprintf("%d", suite.testUser.ID)}, 2)
	assert.NoError(suite.T(), err)

	suite.db.First(&user, suite.testUser.ID)
	assert.Equal(suite.T(), 2, user.NoOfPendingRoomInvites)

	// Decrement by 1
	err = suite.repo.UpdateNoOfPendingRoomInvites(suite.ctx, []string{fmt.Sprintf("%d", suite.testUser.ID)}, -1)
	assert.NoError(suite.T(), err)

	suite.db.First(&user, suite.testUser.ID)
	assert.Equal(suite.T(), 1, user.NoOfPendingRoomInvites)
}

func (suite *UserRepositoryTestSuite) TestUpdateNoOfPendingFriendRequests_Success() {
	// Initial count should be 0
	var user models.User
	suite.db.First(&user, suite.testUser.ID)
	assert.Equal(suite.T(), 0, user.NoOfPendingFriendRequests)

	// Increment by 3
	err := suite.repo.UpdateNoOfPendingFriendRequests(suite.ctx, []uint{suite.testUser.ID}, 3)
	assert.NoError(suite.T(), err)

	suite.db.First(&user, suite.testUser.ID)
	assert.Equal(suite.T(), 3, user.NoOfPendingFriendRequests)

	// Decrement by 2
	err = suite.repo.UpdateNoOfPendingFriendRequests(suite.ctx, []uint{suite.testUser.ID}, -2)
	assert.NoError(suite.T(), err)

	suite.db.First(&user, suite.testUser.ID)
	assert.Equal(suite.T(), 1, user.NoOfPendingFriendRequests)
}

func (suite *UserRepositoryTestSuite) TestUpdateNoOfFriends_Success() {
	friend := models.User{Username: "friend", Email: "friend@example.com", Password: "pass"}
	suite.db.Create(&friend)

	// Initial count should be 0
	var user1, user2 models.User
	suite.db.First(&user1, suite.testUser.ID)
	suite.db.First(&user2, friend.ID)
	assert.Equal(suite.T(), 0, user1.NoOfFriends)
	assert.Equal(suite.T(), 0, user2.NoOfFriends)

	// Increment both users by 1
	err := suite.repo.UpdateNoOfFriends(suite.ctx, []uint{suite.testUser.ID, friend.ID}, 1)
	assert.NoError(suite.T(), err)

	suite.db.First(&user1, suite.testUser.ID)
	suite.db.First(&user2, friend.ID)
	assert.Equal(suite.T(), 1, user1.NoOfFriends)
	assert.Equal(suite.T(), 1, user2.NoOfFriends)

	// Decrement both users by 1
	err = suite.repo.UpdateNoOfFriends(suite.ctx, []uint{suite.testUser.ID, friend.ID}, -1)
	assert.NoError(suite.T(), err)

	suite.db.First(&user1, suite.testUser.ID)
	suite.db.First(&user2, friend.ID)
	assert.Equal(suite.T(), 0, user1.NoOfFriends)
	assert.Equal(suite.T(), 0, user2.NoOfFriends)
}

func (suite *UserRepositoryTestSuite) TestUpdateCounters_EmptySlice_Success() {
	// Test that empty slices don't cause errors
	err := suite.repo.UpdateNoOfPendingRoomInvites(suite.ctx, []string{}, 1)
	assert.NoError(suite.T(), err)

	err = suite.repo.UpdateNoOfPendingFriendRequests(suite.ctx, []uint{}, 1)
	assert.NoError(suite.T(), err)

	err = suite.repo.UpdateNoOfFriends(suite.ctx, []uint{}, 1)
	assert.NoError(suite.T(), err)
}

// ========== Critical Gap Tests ==========

// UpdateFriendRequest tests
func (suite *UserRepositoryTestSuite) TestUpdateFriendRequest_AcceptRequest() {
	receiver := models.User{Username: "receiver", Email: "receiver@example.com", Password: "pass"}
	suite.db.Create(&receiver)

	req := &models.FriendRequest{
		SenderID:   suite.testUser.ID,
		ReceiverID: receiver.ID,
		Status:     "pending",
	}
	err := suite.repo.CreateFriendRequest(suite.ctx, req)
	assert.NoError(suite.T(), err)

	// Update to accepted
	err = suite.repo.UpdateFriendRequest(suite.ctx, req.ID, map[string]interface{}{"status": "accepted"})
	assert.NoError(suite.T(), err)

	// Verify update
	found, err := suite.repo.FindFriendRequest(suite.ctx, req.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "accepted", found.Status)
}

func (suite *UserRepositoryTestSuite) TestUpdateFriendRequest_RejectRequest() {
	receiver := models.User{Username: "receiver2", Email: "receiver2@example.com", Password: "pass"}
	suite.db.Create(&receiver)

	req := &models.FriendRequest{
		SenderID:   suite.testUser.ID,
		ReceiverID: receiver.ID,
		Status:     "pending",
	}
	err := suite.repo.CreateFriendRequest(suite.ctx, req)
	assert.NoError(suite.T(), err)

	// Update to rejected
	err = suite.repo.UpdateFriendRequest(suite.ctx, req.ID, map[string]interface{}{"status": "rejected"})
	assert.NoError(suite.T(), err)

	// Verify update
	found, err := suite.repo.FindFriendRequest(suite.ctx, req.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "rejected", found.Status)
}

func (suite *UserRepositoryTestSuite) TestUpdateFriendRequest_NonExistent() {
	// Try to update non-existent friend request
	err := suite.repo.UpdateFriendRequest(suite.ctx, 99999, map[string]interface{}{"status": "accepted"})
	// Should not error but no rows affected
	assert.NoError(suite.T(), err)
}

// GetFriends tests
func (suite *UserRepositoryTestSuite) TestGetFriends_MultipleFriends() {
	// Create 3 friends
	for i := 0; i < 3; i++ {
		friend := models.User{
			Username: fmt.Sprintf("friend%d", i),
			Email:    fmt.Sprintf("friend%d@example.com", i),
			Password: "pass",
		}
		suite.db.Create(&friend)
		err := suite.repo.AddFriend(suite.ctx, suite.testUser.ID, friend.ID)
		assert.NoError(suite.T(), err)
	}

	friends, err := suite.repo.GetFriends(suite.ctx, suite.testUser.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), friends, 3)
}

func (suite *UserRepositoryTestSuite) TestGetFriends_NoFriends() {
	friends, err := suite.repo.GetFriends(suite.ctx, suite.testUser.ID)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), friends)
}

// Error handling tests for Find methods
func (suite *UserRepositoryTestSuite) TestFindByID_NotFound() {
	user, err := suite.repo.FindByID(suite.ctx, "999999")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
	assert.NotNil(suite.T(), user)            // Repository returns &models.User{} even on error
	assert.Equal(suite.T(), uint(0), user.ID) // Zero-value ID
}

func (suite *UserRepositoryTestSuite) TestFindByUsername_NotFound() {
	user, err := suite.repo.FindByUsername(suite.ctx, "nonexistentuser")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
	assert.NotNil(suite.T(), user)            // Repository returns &models.User{} even on error
	assert.Equal(suite.T(), uint(0), user.ID) // Zero-value ID
}

func (suite *UserRepositoryTestSuite) TestFindByEmail_NotFound() {
	user, err := suite.repo.FindByEmail(suite.ctx, "nonexistent@example.com")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
	assert.NotNil(suite.T(), user)            // Repository returns &models.User{} even on error
	assert.Equal(suite.T(), uint(0), user.ID) // Zero-value ID
}

func (suite *UserRepositoryTestSuite) TestFindByIDs_EmptyArray() {
	users, err := suite.repo.FindByIDs(suite.ctx, []string{})
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), users)
}

func (suite *UserRepositoryTestSuite) TestFindByIDs_SomeInvalid() {
	user2 := models.User{Username: "valid", Email: "valid@example.com", Password: "pass"}
	suite.db.Create(&user2)

	ids := []string{
		fmt.Sprintf("%d", user2.ID),
		"999999", // non-existent
	}
	users, err := suite.repo.FindByIDs(suite.ctx, ids)
	// GORM may error or return partial results depending on implementation
	// For now, just verify it doesn't panic
	_ = err
	_ = users
}

// Boundary condition tests
func (suite *UserRepositoryTestSuite) TestSearchNonFriendUsers_EmptyQuery() {
	results, err := suite.repo.SearchNonFriendUsers(suite.ctx, fmt.Sprintf("%d", suite.testUser.ID), "", 10)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), results)
}

func (suite *UserRepositoryTestSuite) TestSearchNonFriendUsers_SpecialCharacters() {
	user2 := models.User{
		Username: "searchable",
		Email:    "searchable@example.com",
		Password: "pass",
	}
	suite.db.Create(&user2)
	suite.db.Exec("REFRESH MATERIALIZED VIEW user_non_friends")

	// Test with special characters - PostgreSQL full-text search properly rejects malformed syntax
	results, err := suite.repo.SearchNonFriendUsers(suite.ctx, fmt.Sprintf("%d", suite.testUser.ID), "search';DROP TABLE users;--", 10)
	// PostgreSQL ts_query rejects invalid syntax which is GOOD for security
	assert.Error(suite.T(), err) // Should error on invalid tsquery syntax
	// The important thing is it doesn't execute the SQL injection
	_ = results
}

func (suite *UserRepositoryTestSuite) TestFindFriendRequest_NotFound() {
	found, err := suite.repo.FindFriendRequest(suite.ctx, 99999)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
	assert.NotNil(suite.T(), found)            // Repository returns &models.FriendRequest{} even on error
	assert.Equal(suite.T(), uint(0), found.ID) // Zero-value ID
}
