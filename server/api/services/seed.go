package services

import (
	"math/rand"
	"time"

	"github.com/RowenTey/JustJio/server/api/model"
	"github.com/RowenTey/JustJio/server/api/utils"
	"github.com/sirupsen/logrus"

	"gorm.io/gorm"
)

func SeedDB(
	db *gorm.DB,
	userService *UserService,
	roomService *RoomService,
	billService *BillService,
	logger *logrus.Logger,
) error {
	log := logger.WithFields(logrus.Fields{"service": "SeedService"})

	// Check if database is already seeded
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		log.Info("Database already seeded")
		return nil
	}

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
		log.Info("User created: ", users[i])
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
				log.Warn("Error sending friend request: ", err)
				continue
			}
		}

		// accept friend requests
		requests, err := userService.GetFriendRequestsByStatus(u.ID, "pending")
		if err != nil {
			log.Warn("Error getting friend requests: ", err)
			continue
		}

		for _, r := range *requests {
			err := userService.AcceptFriendRequest(r.ID)
			if err != nil {
				log.Warn("Error accepting friend request: ", err)
				continue
			}
		}
	}

	// create rooms
	rooms := []model.Room{
		{
			Name:     "ks birthday",
			Date:     time.Date(2022, time.September, 4, 0, 0, 0, 0, time.UTC),
			Time:     "5:00pm",
			Venue:    "ntu hall 9",
			ImageUrl: "/imgs/birthday.png",
		},
		{
			Name:     "harish birthday",
			Date:     time.Date(2022, time.October, 8, 0, 0, 0, 0, time.UTC),
			Time:     "6:00pm",
			Venue:    "clementi mall",
			ImageUrl: "/imgs/birthday.png",
		},
		{
			Name:     "amabel birthday",
			Date:     time.Date(2022, time.November, 12, 0, 0, 0, 0, time.UTC),
			Time:     "9:00am",
			Venue:    "marina bay sand",
			ImageUrl: "/imgs/birthday.png",
		},
		{
			Name:     "everyone birthday",
			Date:     time.Date(2022, time.January, 7, 0, 0, 0, 0, time.UTC),
			Time:     "10:00am",
			Venue:    "pulau ubin",
			ImageUrl: "/imgs/birthday.png",
		},
		{
			Name:     "mom birthday",
			Date:     time.Date(2022, time.February, 28, 0, 0, 0, 0, time.UTC),
			Time:     "11:00am",
			Venue:    "lot one",
			ImageUrl: "/imgs/birthday.png",
		},
	}

	placeIds := []string{
		"ChIJxz9oLqcP2jER_geaxEqstWg", // ntu hall 9
		"ChIJq8UGJo4a2jER0ypiQDLiXpg", // clementi mall
		"ChIJA5LATO4Z2jER111V-v6abAI", // marina bay sand
		"ChIJYcYVP2I-2jERQTiBy40oT6w", // pulau ubin
		"ChIJs5CBlukR2jER-uyFhQ2NFbU", // lot one
	}

	for i, r := range rooms {
		host := users[rand.Intn(len(users))]
		log.Info("User selected as host: ", host)

		// invite users to room
		var invitees []uint
		for _, u := range users {
			if u.ID == host.ID {
				continue
			}
			invitees = append(invitees, u.ID)
		}

		createdRoom, _, err := roomService.CreateRoomWithInvites(
			&r,
			utils.UIntToString(host.ID),
			placeIds[i],
			&invitees,
		)
		if err != nil {
			return err
		}
		rooms[i] = *createdRoom
		log.Info("Room created: ", rooms[i].ID)

		// only accept invite for first and second room
		if i == 2 {
			continue
		}

		// accept invite
		for _, userid := range invitees {
			err := roomService.UpdateRoomInviteStatus(rooms[i].ID, utils.UIntToString(userid), "accepted")
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
			var payers []uint
			for _, p := range invitees {
				if p == u {
					continue
				}
				payers = append(payers, p)
			}

			if _, err := billService.CreateBill(
				rooms[i].ID,
				utils.UIntToString(u),
				&payers,
				"food",
				float32(j+10)*10,
				true,
			); err != nil {
				log.Errorf("%s", err.Error())
				return err
			}

			if _, err = billService.CreateBill(
				rooms[i].ID,
				utils.UIntToString(u),
				&payers,
				"drinks",
				rand.Float32()*100,
				false,
			); err != nil {
				log.Errorf("%s", err.Error())
				return err
			}
		}

		log.Info("Consolidating bills for room: ", rooms[i].ID)
		// consolidate bills and generate transactions for this room
		if err := billService.
			ConsolidateBills(rooms[i].ID, utils.UIntToString(host.ID)); err != nil {
			log.Errorf("%s", err.Error())
			return err
		}
	}

	return nil
}
