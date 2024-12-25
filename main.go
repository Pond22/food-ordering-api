package main

import (
	"food-ordering-api/api_handlers"
	"food-ordering-api/db"
	_ "food-ordering-api/docs"
	"food-ordering-api/routes"
	service "food-ordering-api/services"
	"log"

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

// @host localhost:8080
// @BasePath /
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @tags categories - การจัดการหมวดหมู่อาหาร
// @tags menu - การจัดการเมนูอาหาร
// @tags orders - การจัดการคำสั่งซื้อ
// @tags tables - การจัดกโต๊ะ
func main() {
	db.InitDatabase()

	app := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: true,

		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		ExposeHeaders:    "Content-Length, Access-Control-Allow-Origin",
		AllowCredentials: false,
		MaxAge:           300,
	}))

	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Get("/ws/tables", api_handlers.TableWebSocketHandler(db.DB))

	// app.Use("/ws/printer", api_handlers.HandlePrinterWebSocket)
	app.Use("/ws/printer", service.HandlePrinterWebSocket)

	routes.SetupRoutes(app)

	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}