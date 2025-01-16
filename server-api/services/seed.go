package services

import (
	"log"
	"math/rand"
	"sc2006-JustJio/model"
	"sc2006-JustJio/util"
	"strconv"
	"time"

	"gorm.io/gorm"
)

func SeedDB(db *gorm.DB) error {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		log.Println("[SEED] Database already seeded")
		return nil
	}

	userService := UserService{DB: db}
	roomService := RoomService{DB: db}

	// create users
	users := []model.User{
		{Username: "harish", Password: "Harish12345!", Email: "harish123@test.com"},
		{Username: "amabel", Password: "Amabel12345!", Email: "amabel123@test.com"},
		{Username: "zhiheng", Password: "Zh12345!", Email: "zh123@test.com"},
		{Username: "eldrick", Password: "Eldrick123!", Email: "eldrick123@test.com"},
		{Username: "kaiseong", Password: "Ks12345!", Email: "ks123@test.com"},
		{Username: "aloysius", Password: "Aloysius12345!", Email: "aloysius123@test.com"},
	}
	for i, u := range users {
		hashedPassword, err := util.HashPassword(u.Password)
		if err != nil {
			return err
		}
		u.Password = hashedPassword

		createdUser, err := userService.CreateOrUpdateUser(&u, true)
		if err != nil {
			return err
		}
		users[i] = *createdUser
		log.Println("[SEED] User created:\n", users[i])
	}

	for _, u := range users {
		// create friends
		for _, f := range users {
			if f.ID == u.ID {
				continue
			}

			userIdStr := strconv.FormatUint(uint64(u.ID), 10)
			friendIdStr := strconv.FormatUint(uint64(f.ID), 10)

			err := userService.AddFriend(userIdStr, friendIdStr)
			if err != nil {
				return err
			}
		}
	}

	// create rooms
	rooms := []model.Room{
		{Name: "ks birthday", Date: time.Date(2022, time.September, 4, 0, 0, 0, 0, time.UTC), Time: "5:00pm", Venue: "ntu hall 9"},
		{Name: "harish birthday", Date: time.Date(2022, time.October, 4, 0, 0, 0, 0, time.UTC), Time: "6:00pm", Venue: "clementi mall"},
		{Name: "amabel birthday", Date: time.Date(2022, time.November, 4, 0, 0, 0, 0, time.UTC), Time: "9:00am", Venue: "marina bay sand"},
	}
	for i, r := range rooms {
		host := users[rand.Intn(len(users))]
		log.Println("[SEED] User selected as host:\n", host)
		createdRoom, err := roomService.CreateRoom(&r, &host)
		if err != nil {
			return err
		}
		rooms[i] = *createdRoom
		log.Println("[SEED] Room created: ", rooms[i].ID)

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
			log.Fatalf("%s", err.Error())
			return err
		}
	}

	return nil
}
