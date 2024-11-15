package routes

import (
	"fmt"
	"food-ordering-api/api_handlers"
	"food-ordering-api/db"
	"food-ordering-api/models"
	qr_service "food-ordering-api/services"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {

	app.Post("/add_menu", api_handlers.CreateMenuItemHandler)

	app.Post("/add_category", api_handlers.CreateCategoryHandler)

	app.Get("/getCategory", api_handlers.GetCategoriesHandler)

	app.Get("/api/action", api_handlers.GetMenu)

	app.Get("/qr_code", qr_service.HandleQRCodeRequest)

	app.Get("/order", qr_service.Table)

	app.Post("/order", qr_service.Order)
	//-----

	app.Get("/ping", pingHandler)

	app.Post("/migrate", migrateHandler)

	app.Get("/hello/:name/:num", helloHandler)

}

func pingHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "pong"})
}

func migrateHandler(c *fiber.Ctx) error {
	err := migrate()
	if err != nil {
		return c.Status(500).SendString(fmt.Sprintf("Migration failed: %v", err))
	}
	return c.SendString("Database migration successful")
}

func helloHandler(c *fiber.Ctx) error {
	num, err := strconv.Atoi(c.Params("num"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid number:%s , you are not a human bruh", c.Params("num")),
		})
	}
	return c.JSON(fiber.Map{"message": hello(c.Params("name"), num)})
}

// @Summary อัปเดต orm
// @Description api เส้นนี้ใช้เพื่อ migrate ฐานข้อมูลของ orm
// @Produce json
// @Router /migrate [post]
func migrate() error {
	if err := db.DB.AutoMigrate(&models.MenuItem{}); err != nil {
		return err
	}
	return nil
}

func hello(name string, round int) string {
	return "bello " + name + " " + strconv.Itoa(round) + " times,i'm not human"
}
