package services

import (
	"math/rand"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/utils"

	"gorm.io/gorm"
)

func SeedDB(db *gorm.DB) error {
	logger := log.WithFields(log.Fields{"service": "SeedService"})

	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		logger.Info("Database already seeded")
		return nil
	}

	userService := NewUserService(db)
	roomService := NewRoomService(db)
	billService := NewBillService(db)
	transactionService := NewTransactionService(db)

	// create users
	users := []model.User{
		{Username: "harish", Password: "Harish12345!", Email: "harish123@test.com"},
		{Username: "amabel", Password: "Amabel12345!", Email: "amabel123@test.com"},
		{Username: "zhiheng", Password: "Zh12345!", Email: "zh123@test.com"},
		{Username: "eldrick", Password: "Eldrick123!", Email: "eldrick123@test.com"},
		{Username: "kaiseong", Password: "Ks12345!", Email: "ks123@test.com"},
		{Username: "aloysius", Password: "Aloysius12345!", Email: "aloysius123@test.com"},
		{Username: "test", Password: "Test12345!", Email: "test@test.com"},
		{Username: "happy", Password: "Happy12345!", Email: "happy@test.com"},
	}

	for i, u := range users {
		hashedPassword, err := utils.HashPassword(u.Password)
		if err != nil {
			return err
		}
		u.Password = hashedPassword

		createdUser, err := userService.CreateOrUpdateUser(&u, true)
		if err != nil {
			return err
		}
		users[i] = *createdUser
		logger.Info("User created:\n", users[i])
	}

	// Remove test user from users
	users = users[:len(users)-2]

	for _, u := range users {
		// create friends
		for _, f := range users {
			if f.ID == u.ID {
				continue
			}

			err := userService.SendFriendRequest(u.ID, f.ID)
			if err != nil {
				logger.Warn("Error sending friend request: ", err)
				continue
			}
		}

		// accept friend requests
		requests, err := userService.GetFriendRequestsByStatus(u.ID, "pending")
		if err != nil {
			logger.Warn("Error getting friend requests: ", err)
			continue
		}

		for _, r := range *requests {
			err := userService.AcceptFriendRequest(r.ID)
			if err != nil {
				logger.Warn("Error accepting friend request: ", err)
				continue
			}
		}
	}

	// create rooms
	rooms := []model.Room{
		{Name: "ks birthday", Date: time.Date(2022, time.September, 4, 0, 0, 0, 0, time.UTC), Time: "5:00pm", Venue: "ntu hall 9"},
		{Name: "harish birthday", Date: time.Date(2022, time.October, 8, 0, 0, 0, 0, time.UTC), Time: "6:00pm", Venue: "clementi mall"},
		{Name: "amabel birthday", Date: time.Date(2022, time.November, 12, 0, 0, 0, 0, time.UTC), Time: "9:00am", Venue: "marina bay sand"},
		{Name: "everyone birthday", Date: time.Date(2022, time.January, 7, 0, 0, 0, 0, time.UTC), Time: "10:00am", Venue: "pulau ubin"},
		{Name: "mom birthday", Date: time.Date(2022, time.February, 28, 0, 0, 0, 0, time.UTC), Time: "11:00am", Venue: "batam"},
	}

	for i, r := range rooms {
		host := users[rand.Intn(len(users))]
		logger.Info("User selected as host:\n", host)
		createdRoom, err := roomService.CreateRoom(&r, &host)
		if err != nil {
			return err
		}
		rooms[i] = *createdRoom
		logger.Info("Room created: ", rooms[i].ID)

		// invite users to room
		var invitees = []model.User{}
		for _, u := range users {
			if u.ID == host.ID {
				continue
			}
			invitees = append(invitees, u)
		}

		_, err = roomService.InviteUserToRoom(
			rooms[i].ID, &host, &invitees, "Join my party!")
		if err != nil {
			log.Errorf("%s", err.Error())
			return err
		}

		// only accept invite for first and second room
		if i == 2 {
			continue
		}

		// accept invite
		for _, u := range invitees {
			err := roomService.UpdateRoomInviteStatus(rooms[i].ID, strconv.FormatUint(uint64(u.ID), 10), "accepted")
			if err != nil {
				log.Errorf("%s", err.Error())
				return err
			}
		}

		// only create bills for first room
		if i != 0 {
			continue
		}

		// create bill
		for j, u := range invitees {
			var payers = []model.User{}
			for _, p := range invitees {
				if p.ID == u.ID {
					continue
				}
				payers = append(payers, p)
			}

			_, err := billService.CreateBill(&rooms[i], &u, "food", float32(j+10)*10, true, &payers)
			if err != nil {
				log.Errorf("%s", err.Error())
				return err
			}

			_, err = billService.CreateBill(&rooms[i], &u, "drinks", rand.Float32()*100, false, &payers)
			if err != nil {
				log.Errorf("%s", err.Error())
				return err
			}
		}

		// consolidate bills for this room
		consolidation, err := billService.ConsolidateBills(db, rooms[i].ID)
		if err != nil {
			log.Errorf("%s", err.Error())
			return err
		}

		// generate transactions for this room
		err = transactionService.GenerateTransactions(consolidation)
		if err != nil {
			log.Errorf("%s", err.Error())
			return err
		}
	}

	return nil
}
