package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/RowenTey/JustJio/server/api/database"
	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/tests"
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

	testUser *model.User
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error
	suite.logger = logrus.New()

	// Setup test containers
	suite.dependencies = &tests.TestDependencies{}
	suite.dependencies, err = tests.SetupPgDependency(suite.ctx, suite.dependencies, suite.logger)
	assert.NoError(suite.T(), err)

	// Get PostgreSQL connection string
	pgConnStr, err := suite.dependencies.PostgresContainer.ConnectionString(suite.ctx)
	assert.NoError(suite.T(), err)
	fmt.Println("Test DB Connection String:", pgConnStr)

	// Initialize database
	suite.db, err = database.InitTestDB(pgConnStr)
	assert.NoError(suite.T(), err)

	// Run migrations
	err = database.Migrate(suite.db)
	assert.NoError(suite.T(), err)

	suite.repo = NewUserRepository(suite.db)
}

func (suite *UserRepositoryTestSuite) TearDownSuite() {
	if suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	// Insert base user
	user := model.User{
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
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (suite *UserRepositoryTestSuite) TestCreateAndFindByID_Success() {
	user := &model.User{
		Username: "john",
		Email:    "john@example.com",
		Password: "pass",
	}
	created, err := suite.repo.Create(user)
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByID(fmt.Sprintf("%d", created.ID))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "john", found.Username)
}

func (suite *UserRepositoryTestSuite) TestFindByUsernameAndEmail_Success() {
	foundByUsername, err := suite.repo.FindByUsername(suite.testUser.Username)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.testUser.Email, foundByUsername.Email)

	foundByEmail, err := suite.repo.FindByEmail(suite.testUser.Email)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.testUser.Username, foundByEmail.Username)
}

func (suite *UserRepositoryTestSuite) TestAddAndCheckFriendship_Success() {
	friend := model.User{
		Username: "friend",
		Email:    "friend@example.com",
		Password: "pass",
	}
	err := suite.db.Create(&friend).Error
	assert.NoError(suite.T(), err)

	err = suite.repo.AddFriend(suite.testUser.ID, friend.ID)
	assert.NoError(suite.T(), err)

	isFriend, err := suite.repo.CheckFriendship(suite.testUser.ID, friend.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), isFriend)
}

func (suite *UserRepositoryTestSuite) TestCreateFriendRequestAndExists_Success() {
	receiver := model.User{
		Username: "bob",
		Email:    "bob@example.com",
		Password: "pw",
	}
	err := suite.db.Create(&receiver).Error
	assert.NoError(suite.T(), err)

	req := &model.FriendRequest{
		SenderID:   suite.testUser.ID,
		ReceiverID: receiver.ID,
		Status:     "pending",
	}
	err = suite.repo.CreateFriendRequest(req)
	assert.NoError(suite.T(), err)

	exists, err := suite.repo.CheckFriendRequestExists(suite.testUser.ID, receiver.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

func (suite *UserRepositoryTestSuite) TestSearchUsers_NoFriends_Success() {
	user2 := model.User{
		Username: "search_target",
		Email:    "target@example.com",
		Password: "secret",
	}
	suite.db.Create(&user2)

	results, err := suite.repo.SearchUsers(fmt.Sprintf("%d", suite.testUser.ID), "search", 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *results, 1)
	assert.Equal(suite.T(), "search_target", (*results)[0].Username)
}

func (suite *UserRepositoryTestSuite) TestSearchUsers_WithFriends_Success() {
	user2 := model.User{
		Username: "search_target",
		Email:    "target@example.com",
		Password: "secret",
	}
	suite.db.Create(&user2)

	// Add user2 as a friend to testUser
	err := suite.repo.AddFriend(suite.testUser.ID, user2.ID)
	assert.NoError(suite.T(), err)

	user3 := model.User{
		Username: "another_target",
		Email:    "another_target@example.com",
		Password: "secret",
	}
	suite.db.Create(&user3)

	results, err := suite.repo.SearchUsers(fmt.Sprintf("%d", suite.testUser.ID), "another", 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *results, 1)
	assert.Equal(suite.T(), "another_target", (*results)[0].Username)
}

func (suite *UserRepositoryTestSuite) TestCountFriends_Success() {
	friend := model.User{
		Username: "frienduser",
		Email:    "friend@example.com",
		Password: "pass",
	}
	err := suite.db.Create(&friend).Error
	assert.NoError(suite.T(), err)

	err = suite.repo.AddFriend(suite.testUser.ID, friend.ID)
	assert.NoError(suite.T(), err)

	count, err := suite.repo.CountFriends(suite.testUser.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *UserRepositoryTestSuite) TestRemoveFriend_Success() {
	friend := model.User{
		Username: "removable",
		Email:    "removable@example.com",
		Password: "pass",
	}
	err := suite.db.Create(&friend).Error
	assert.NoError(suite.T(), err)

	err = suite.repo.AddFriend(suite.testUser.ID, friend.ID)
	assert.NoError(suite.T(), err)

	err = suite.repo.RemoveFriend(suite.testUser.ID, friend.ID)
	assert.NoError(suite.T(), err)

	isFriend, err := suite.repo.CheckFriendship(suite.testUser.ID, friend.ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), isFriend)
}

func (suite *UserRepositoryTestSuite) TestGetUninvitedFriends_Success() {
	// Setup another user as a friend
	friend := model.User{
		Username: "uninvited",
		Email:    "uninvited@example.com",
		Password: "pass",
	}
	err := suite.db.Create(&friend).Error
	assert.NoError(suite.T(), err)

	// Make them friends
	err = suite.repo.AddFriend(suite.testUser.ID, friend.ID)
	assert.NoError(suite.T(), err)

	// Create a room and do not invite the friend
	room := model.Room{
		Name:   "Test Room",
		HostID: suite.testUser.ID,
	}
	err = suite.db.Create(&room).Error
	assert.NoError(suite.T(), err)

	// Check GetUninvitedFriends
	uninvited, err := suite.repo.GetUninvitedFriends(room.ID, fmt.Sprintf("%d", suite.testUser.ID))
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *uninvited, 1)
	assert.Equal(suite.T(), friend.ID, (*uninvited)[0].ID)
}

func (suite *UserRepositoryTestSuite) TestFindByIDs_Success() {
	user2 := model.User{Username: "u2", Email: "u2@example.com", Password: "pass"}
	user3 := model.User{Username: "u3", Email: "u3@example.com", Password: "pass"}
	suite.db.Create(&user2)
	suite.db.Create(&user3)

	ids := []uint{suite.testUser.ID, user2.ID, user3.ID}
	users, err := suite.repo.FindByIDs(&ids)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *users, 3)
}

func (suite *UserRepositoryTestSuite) TestUpdateUser_Success() {
	suite.testUser.Username = "updated_username"
	err := suite.repo.Update(suite.testUser)
	assert.NoError(suite.T(), err)

	var user model.User
	suite.db.First(&user, suite.testUser.ID)
	assert.Equal(suite.T(), "updated_username", user.Username)
}

func (suite *UserRepositoryTestSuite) TestDeleteUser_Success() {
	err := suite.repo.Delete(fmt.Sprintf("%d", suite.testUser.ID))
	assert.NoError(suite.T(), err)

	var user model.User
	err = suite.db.First(&user, suite.testUser.ID).Error
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

func (suite *UserRepositoryTestSuite) TestFindAndCountFriendRequestsByReceiver_Success() {
	sender := model.User{Username: "sender", Email: "sender@example.com", Password: "pass"}
	suite.db.Create(&sender)

	request := model.FriendRequest{
		SenderID:   sender.ID,
		ReceiverID: suite.testUser.ID,
		Status:     "pending",
	}
	suite.db.Create(&request)

	requests, err := suite.repo.FindFriendRequestsByReceiver(suite.testUser.ID, "pending")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *requests, 1)
	assert.Equal(suite.T(), sender.ID, (*requests)[0].SenderID)

	count, err := suite.repo.CountFriendRequestsByReceiver(suite.testUser.ID, "pending")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *UserRepositoryTestSuite) TestFindAndUpdateFriendRequest_Success() {
	sender := model.User{Username: "sender2", Email: "sender2@example.com", Password: "pass"}
	suite.db.Create(&sender)

	request := model.FriendRequest{
		SenderID:   sender.ID,
		ReceiverID: suite.testUser.ID,
		Status:     "pending",
	}
	suite.db.Create(&request)

	// Find
	found, err := suite.repo.FindFriendRequest(request.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "pending", found.Status)

	// Update
	found.Status = "accepted"
	err = suite.repo.UpdateFriendRequest(found)
	assert.NoError(suite.T(), err)

	// Verify update
	var updated model.FriendRequest
	suite.db.First(&updated, found.ID)
	assert.Equal(suite.T(), "accepted", updated.Status)
}
