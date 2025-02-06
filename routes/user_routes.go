package routes

import (
	"food-ordering-api/api_handlers"
	"food-ordering-api/models"
	utils "food-ordering-api/utility"

	"github.com/gofiber/fiber/v2"
)

// SetupUserRoutes กำหนด routes สำหรับจัดการผู้ใช้
func SetupUserRoutes(app *fiber.App) {
	api := app.Group("/api")

	// สร้าง group สำหรับ user routes ที่ต้องการ authentication
	users := api.Group("/users", utils.AuthRequired())

	// routes สำหรับ manager เท่านั้น
	managerOnly := users.Group("/", utils.RoleRequired(models.RoleManager))
	{
		// ดูรายชื่อผู้ใช้ทั้งหมด
		managerOnly.Get("/", api_handlers.GetUsers)

		// สร้างผู้ใช้ใหม่
		managerOnly.Post("/", api_handlers.CreateUser)

		// รีเซ็ตรหัสผ่านให้ผู้ใช้คนอื่น
		managerOnly.Put("/:id/reset-password", api_handlers.ResetUserPassword)
	}

	// routes สำหรับผู้ใช้ทุกคนที่ login แล้ว
	{
		// users.Get("/profile", api_handlers.GetUserProfile)
		// users.Put("/change-password", api_handlers.ChangePassword)
	}
}
