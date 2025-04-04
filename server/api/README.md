# JustJio-API-Server

> REST API Server for JustJio

![server-landing](../client/assets/gifs/JustJio-Server.gif)

## 🛠 Getting Started

> Make sure you're at the `server-api` directory and run the following scripts in the terminal.

1\. Install dependencies

```terminal
go mod tidy
```

2\. Make a copy of `.env` and populate the environment variables inside

```terminal
copy .env.example .env
```

3\. Run the code

```terminal
air dev
```

or if you don't have `air` installed

```terminal
go run main.go dev
```

## 📂 Project Folder Structure

#### Top Level Directory Layout

```terminal
.
├── config/            # configurations
├── database/          # global DB object
├── handlers/          # API handlers
├── middleware/        # middleware logic
├── model/             # model & DTOs
├── router/            # API routing
├── main.go            # driver code
├── Dockerfile
├── .env.example
├── go.mod
├── go.sum
└── README.md
```

### API Testing Tracker

```
auth := v1.Group("/auth")
auth.Post("/", handlers.Login) /
auth.Post("/signup", handlers.SignUp) x
auth.Post("/verify", handlers.VerifyOTP) x

// private routes
users := v1.Group("/users")
users.Get("/:id", handlers.GetUser) /
users.Patch("/:id", handlers.UpdateUser) x
users.Delete("/:id", handlers.DeleteUser) /

rooms := v1.Group("/rooms")
rooms.Get("/", handlers.GetRooms) /
rooms.Get("/:id", handlers.GetRoom) /
rooms.Get("/invites", handlers.GetRoomInvitations) /
rooms.Get("/attendees/:id", handlers.GetRoomAttendees) /
rooms.Post("/", handlers.CreateRoom)
rooms.Post("/:id", handlers.InviteUser)
rooms.Patch("/:id", handlers.RespondToRoomInvite)
rooms.Patch("/close/:id", handlers.CloseRoom)
rooms.Patch("/leave/:id", handlers.LeaveRoom)

messages := rooms.Group("/:roomId/messages")
messages.Use(middleware.IsUserInRoom)
messages.Get("/", handlers.GetMessages)
messages.Get("/:id", handlers.GetMessage)
messages.Post("/", handlers.CreateMessage)
```
