package main

import (
	"log"
	"os"

	"sc2006-JustJio/config"
	"sc2006-JustJio/database"
	"sc2006-JustJio/middleware"
	"sc2006-JustJio/router"
	"sc2006-JustJio/services"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	_ "sc2006-JustJio/docs"
)

// @title           JustJio API
// @version         1.0
// @description     This is a server for NTU SC2006's project - JustJio.
// @termsOfService  http://swagger.io/terms/
// @contact.name   Kai Seong
// @contact.email  kaiseong02@gmail.com
// @license.name  MIT
// @license.url  https://opensource.org/licenses/MIT
// @host      localhost:8080
// @BasePath  /
// @securityDefinitions.basic  BasicAuth
// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        Authorization
// @description									authorise authenticated users
func main() {
	log.Println("Starting server...")

	env := ""
	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	// only load .env file if in dev environment
	if env == "dev" {
		godotenv.Load(".env")
	}

	app := fiber.New()

	database.ConnectDB()
	if env == "dev" {
		services.SeedDB(database.DB)
	}

	middleware.Fiber(app)
	router.Initalize(app)
	log.Fatal(app.Listen(":" + config.Config("PORT")))
}
