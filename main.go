package main

import (
	"food-ordering-api/db"
	_ "food-ordering-api/docs" // สำคัญ: ต้องมี import docs
	"food-ordering-api/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// @title Food Ordering API
// @version 1.0
// @description This is a simple API for ordering food.
// @termsOfService https://example.com/terms

// @contact.name API Support
// @contact.url https://www.example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
func main() {

	db.InitDatabase()

	app := fiber.New()

	app.Get("/swagger/*", swagger.HandlerDefault)

	routes.SetupRoutes(app)

	app.Listen(":8080")
}
