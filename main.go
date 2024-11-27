package main

import (
	"food-ordering-api/db"
	_ "food-ordering-api/docs"
	"food-ordering-api/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

// @host http://localhost/8080
// @BasePath /
// @schemes http

func main() {

	db.InitDatabase()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		// AllowOrigins: "http://localhost:5173, http://127.0.0.1:8080",      // ระบุ React app ที่จะอนุญาต
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE",   // อนุญาตวิธีการ HTTP
		AllowHeaders: "Content-Type,Authorization", // อนุญาต headers
	}))

	app.Get("/swagger/*", swagger.HandlerDefault)

	routes.SetupRoutes(app)

	app.Listen(":8080")
}
