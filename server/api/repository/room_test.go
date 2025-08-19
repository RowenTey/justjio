package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/tests"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type RoomRepositoryTestSuite struct {
	suite.Suite
	db           *gorm.DB
	ctx          context.Context
	repo         RoomRepository
	logger       *logrus.Logger
	dependencies *tests.TestDependencies

	testUser *model.User
}

func (suite *RoomRepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error
	suite.logger = logrus.New()

	// Setup test containers
	suite.dependencies = &tests.TestDependencies{}
	suite.dependencies, err = tests.SetupPgDependency(suite.ctx, suite.dependencies, suite.logger)
	assert.NoError(suite.T(), err)

	// Setup DB Conn
	suite.db, err = tests.CreateAndConnectToTestDb(suite.ctx, suite.dependencies.PostgresContainer, "room_test")
	assert.NoError(suite.T(), err)

	suite.repo = NewRoomRepository(suite.db)
}

func (suite *RoomRepositoryTestSuite) TearDownSuite() {
	if !IsPackageTest && suite.dependencies != nil {
		suite.dependencies.Teardown(suite.ctx)
	}
}

func (suite *RoomRepositoryTestSuite) SetupTest() {
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

func (suite *RoomRepositoryTestSuite) TearDownTest() {
	suite.db.Exec("TRUNCATE TABLE room_invites RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE room_users RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE rooms RESTART IDENTITY CASCADE")
	suite.db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
}

func TestRoomRepositorySuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(RoomRepositoryTestSuite))
}

func (suite *RoomRepositoryTestSuite) TestCreateAndGetByID_Success() {
	room := model.Room{
		Name:   "Test Room",
		HostID: suite.testUser.ID,
	}
	err := suite.repo.Create(&room)
	assert.NoError(suite.T(), err)

	got, err := suite.repo.GetByID(room.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), room.ID, got.ID)
	assert.Equal(suite.T(), "Test Room", got.Name)
}

func (suite *RoomRepositoryTestSuite) TestAddAndRemoveUserFromRoom_Success() {
	userIdStr := fmt.Sprintf("%d", suite.testUser.ID)
	room := model.Room{Name: "WithUser", HostID: suite.testUser.ID}
	err := suite.repo.Create(&room)
	assert.NoError(suite.T(), err)

	err = suite.repo.AddUserToRoom(room.ID, suite.testUser)
	assert.NoError(suite.T(), err)

	isIn, err := suite.repo.IsUserInRoom(room.ID, userIdStr)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), isIn)

	err = suite.repo.RemoveUserFromRoom(room.ID, userIdStr)
	assert.NoError(suite.T(), err)

	isIn, err = suite.repo.IsUserInRoom(room.ID, userIdStr)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), isIn)
}

func (suite *RoomRepositoryTestSuite) TestGetRoomAttendees_Success() {
	room := model.Room{Name: "RoomWithAttendees", HostID: suite.testUser.ID}
	err := suite.repo.Create(&room)
	assert.NoError(suite.T(), err)

	err = suite.repo.AddUserToRoom(room.ID, suite.testUser)
	assert.NoError(suite.T(), err)

	users, err := suite.repo.GetRoomAttendees(room.ID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *users, 1)
	assert.Equal(suite.T(), suite.testUser.ID, (*users)[0].ID)
}

func (suite *RoomRepositoryTestSuite) TestGetRoomAttendeeIDs_Success() {
	room := model.Room{Name: "AttendeeIDRoom", HostID: suite.testUser.ID}
	err := suite.repo.Create(&room)
	assert.NoError(suite.T(), err)

	user1 := model.User{Username: "userA", Email: "userA@example.com", Password: "pass"}
	user2 := model.User{Username: "userB", Email: "userB@example.com", Password: "pass"}
	err = suite.db.Create(&user1).Error
	assert.NoError(suite.T(), err)
	err = suite.db.Create(&user2).Error
	assert.NoError(suite.T(), err)

	err = suite.repo.AddUserToRoom(room.ID, &user1)
	assert.NoError(suite.T(), err)
	err = suite.repo.AddUserToRoom(room.ID, &user2)
	assert.NoError(suite.T(), err)

	ids, err := suite.repo.GetRoomAttendeeIDs(room.ID)
	assert.NoError(suite.T(), err)
	assert.ElementsMatch(suite.T(), []string{fmt.Sprintf("%d", user1.ID), fmt.Sprintf("%d", user2.ID)}, *ids)
}

func (suite *RoomRepositoryTestSuite) TestCountUserRooms_Success() {
	room := model.Room{Name: "CountRoom", HostID: suite.testUser.ID}
	err := suite.repo.Create(&room)
	assert.NoError(suite.T(), err)

	err = suite.repo.AddUserToRoom(room.ID, suite.testUser)
	assert.NoError(suite.T(), err)

	count, err := suite.repo.CountUserRooms(fmt.Sprintf("%d", suite.testUser.ID))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *RoomRepositoryTestSuite) TestCloseRoom_Success() {
	room := model.Room{Name: "Closable", HostID: suite.testUser.ID}
	err := suite.repo.Create(&room)
	assert.NoError(suite.T(), err)

	err = suite.repo.CloseRoom(room.ID)
	assert.NoError(suite.T(), err)

	updated, err := suite.repo.GetByID(room.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), updated.IsClosed)
}

func (suite *RoomRepositoryTestSuite) TestCreateInviteAndHasPendingInvites_Success() {
	room := model.Room{Name: "HasPending", HostID: suite.testUser.ID}
	err := suite.repo.Create(&room)
	assert.NoError(suite.T(), err)

	invitee := model.User{Username: "invitee1", Email: "invitee1@example.com", Password: "pass"}
	err = suite.db.Create(&invitee).Error
	assert.NoError(suite.T(), err)

	invite := model.RoomInvite{
		RoomID:    room.ID,
		UserID:    invitee.ID,
		InviterID: suite.testUser.ID,
		Status:    "pending",
	}
	err = suite.repo.CreateInvites(&[]model.RoomInvite{invite})
	assert.NoError(suite.T(), err)

	has, err := suite.repo.HasPendingInvites(room.ID, fmt.Sprintf("%d", invitee.ID))
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), has)
}

func (suite *RoomRepositoryTestSuite) TestCountPendingInvites_Success() {
	room := model.Room{Name: "CountPending", HostID: suite.testUser.ID}
	err := suite.repo.Create(&room)
	assert.NoError(suite.T(), err)

	invitee := model.User{Username: "invitee2", Email: "invitee2@example.com", Password: "pass"}
	err = suite.db.Create(&invitee).Error
	assert.NoError(suite.T(), err)

	invite := model.RoomInvite{
		RoomID:    room.ID,
		UserID:    invitee.ID,
		InviterID: suite.testUser.ID,
		Status:    "pending",
	}
	err = suite.repo.CreateInvites(&[]model.RoomInvite{invite})
	assert.NoError(suite.T(), err)

	count, err := suite.repo.CountPendingInvites(fmt.Sprintf("%d", invitee.ID))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *RoomRepositoryTestSuite) TestGetPendingInvites_Success() {
	room := model.Room{Name: "GetPending", HostID: suite.testUser.ID}
	err := suite.repo.Create(&room)
	assert.NoError(suite.T(), err)

	invitee := model.User{Username: "invitee3", Email: "invitee3@example.com", Password: "pass"}
	err = suite.db.Create(&invitee).Error
	assert.NoError(suite.T(), err)

	invite := model.RoomInvite{
		RoomID:    room.ID,
		UserID:    invitee.ID,
		InviterID: suite.testUser.ID,
		Status:    "pending",
	}
	err = suite.repo.CreateInvites(&[]model.RoomInvite{invite})
	assert.NoError(suite.T(), err)

	invites, err := suite.repo.GetPendingInvites(fmt.Sprintf("%d", invitee.ID))
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *invites, 1)
	assert.Equal(suite.T(), invitee.ID, (*invites)[0].User.ID)
	assert.Equal(suite.T(), suite.testUser.ID, (*invites)[0].Inviter.ID)
	assert.Equal(suite.T(), room.ID, (*invites)[0].Room.ID)
}

func (suite *RoomRepositoryTestSuite) TestUpdateInviteStatus_Success() {
	room := model.Room{Name: "UpdateStatus", HostID: suite.testUser.ID}
	err := suite.repo.Create(&room)
	assert.NoError(suite.T(), err)

	invitee := model.User{Username: "invitee4", Email: "invitee4@example.com", Password: "pass"}
	err = suite.db.Create(&invitee).Error
	assert.NoError(suite.T(), err)

	invite := model.RoomInvite{
		RoomID:    room.ID,
		UserID:    invitee.ID,
		InviterID: suite.testUser.ID,
		Status:    "pending",
	}
	err = suite.repo.CreateInvites(&[]model.RoomInvite{invite})
	assert.NoError(suite.T(), err)

	err = suite.repo.UpdateInviteStatus(room.ID, fmt.Sprintf("%d", invitee.ID), "accepted")
	assert.NoError(suite.T(), err)

	var updated model.RoomInvite
	err = suite.db.First(&updated, "room_id = ? AND user_id = ?", room.ID, invitee.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "accepted", updated.Status)
}

func (suite *RoomRepositoryTestSuite) TestDeletePendingInvites_Success() {
	room := model.Room{Name: "DeletePending", HostID: suite.testUser.ID}
	err := suite.repo.Create(&room)
	assert.NoError(suite.T(), err)

	invitee1 := model.User{Username: "invitee5", Email: "invitee5@example.com", Password: "pass"}
	invitee2 := model.User{Username: "invitee6", Email: "invitee6@example.com", Password: "pass"}
	err = suite.db.Create(&invitee1).Error
	assert.NoError(suite.T(), err)
	err = suite.db.Create(&invitee2).Error
	assert.NoError(suite.T(), err)

	invites := []model.RoomInvite{
		{RoomID: room.ID, UserID: invitee1.ID, InviterID: suite.testUser.ID, Status: "pending"},
		{RoomID: room.ID, UserID: invitee2.ID, InviterID: suite.testUser.ID, Status: "accepted"},
	}
	err = suite.repo.CreateInvites(&invites)
	assert.NoError(suite.T(), err)

	err = suite.repo.DeletePendingInvites(room.ID)
	assert.NoError(suite.T(), err)

	var remaining []model.RoomInvite
	err = suite.db.Where("room_id = ?", room.ID).Find(&remaining).Error
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), remaining, 1)
	assert.Equal(suite.T(), "accepted", remaining[0].Status)
}

func (suite *RoomRepositoryTestSuite) TestGetUnjoinedRoomsByIsPrivate_Success() {
	otherUser := model.User{Username: "host1", Email: "host1@example.com", Password: "pass"}
	err := suite.db.Create(&otherUser).Error
	assert.NoError(suite.T(), err)

	unjoinedRoom := model.Room{
		Name:      "PrivateUnjoined",
		IsPrivate: true,
		HostID:    otherUser.ID,
	}
	err = suite.repo.Create(&unjoinedRoom)
	assert.NoError(suite.T(), err)

	rooms, err := suite.repo.GetUnjoinedRoomsByIsPrivate(fmt.Sprintf("%d", suite.testUser.ID), true)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *rooms, 1)
	assert.Equal(suite.T(), "PrivateUnjoined", (*rooms)[0].Name)
}

func (suite *RoomRepositoryTestSuite) TestGetUnjoinedRoomsByIsPrivate_ExcludesJoinedRooms() {
	otherUser := model.User{Username: "host2", Email: "host2@example.com", Password: "pass"}
	err := suite.db.Create(&otherUser).Error
	assert.NoError(suite.T(), err)

	room := model.Room{
		Name:      "PrivateJoined",
		IsPrivate: true,
		HostID:    otherUser.ID,
	}
	err = suite.repo.Create(&room)
	assert.NoError(suite.T(), err)

	err = suite.repo.AddUserToRoom(room.ID, suite.testUser)
	assert.NoError(suite.T(), err)

	rooms, err := suite.repo.GetUnjoinedRoomsByIsPrivate(fmt.Sprintf("%d", suite.testUser.ID), true)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), *rooms)
}

func (suite *RoomRepositoryTestSuite) TestGetUnjoinedRoomsByIsPrivate_ExcludesInvitedRooms() {
	otherUser := model.User{Username: "host3", Email: "host3@example.com", Password: "pass"}
	err := suite.db.Create(&otherUser).Error
	assert.NoError(suite.T(), err)

	room := model.Room{
		Name:      "PrivateInvited",
		IsPrivate: true,
		HostID:    otherUser.ID,
	}
	err = suite.repo.Create(&room)
	assert.NoError(suite.T(), err)

	invite := model.RoomInvite{
		RoomID:    room.ID,
		UserID:    suite.testUser.ID,
		InviterID: otherUser.ID,
		Status:    "pending",
	}
	err = suite.repo.CreateInvites(&[]model.RoomInvite{invite})
	assert.NoError(suite.T(), err)

	rooms, err := suite.repo.GetUnjoinedRoomsByIsPrivate(fmt.Sprintf("%d", suite.testUser.ID), true)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), *rooms)
}
