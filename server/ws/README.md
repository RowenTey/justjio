# JustJio-WS-Server

> WebSockets Server for JustJio

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
├── services/     # specific services
├── utils/        # utility functions
├── main.go       # driver code
├── Dockerfile
├── .env.example
├── go.mod
├── go.sum
└── README.md
```
