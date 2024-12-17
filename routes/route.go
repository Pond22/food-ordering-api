package routes

import (
	"food-ordering-api/api_handlers"
	"food-ordering-api/models"
	qr_service "food-ordering-api/services"
	utils "food-ordering-api/utility"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// API Group
	api := app.Group("/api")

	// Auth Routes
	auth := api.Group("/auth")
	{
		auth.Post("/login", api_handlers.Login)
		auth.Get("/verify-token", api_handlers.VerifyToken)
	}
	// User Routes
	// user := api.Group("/member", utils.AuthRequired())
	user := api.Group("/member") //เอา middleware ออก deploy อย่าลืมเอาใส่
	{
		user.Post("/", api_handlers.CreateUser, utils.RoleRequired(models.RoleManager))
		user.Get("/", api_handlers.GetUsers, utils.RoleRequired(models.RoleManager))
		user.Get("/get_member_profile", api_handlers.GetUserProfile)
		user.Put("/change-password", api_handlers.ChangePassword)
		user.Put("/:id/reset-password", api_handlers.ResetUserPassword, utils.RoleRequired(models.RoleManager)) // ผจก. เปลี่ยนรหัสผ่านพนักงาน
		user.Delete("/:id/Delete-member", api_handlers.Delete_user, utils.RoleRequired(models.RoleManager))     // ผจก. ลบพนักงาน
	}

	// Menu Management Routes - ต้องการการยืนยันตัวตนต้อง manager
	// menu := api.Group("/menu", utils.AuthRequired(), utils.RoleRequired(models.RoleManager))
	menu := api.Group("/menu")
	{
		// เมนูพื้นฐาน
		menu.Post("/", api_handlers.CreateMenuItemHandler)
		menu.Get("/", api_handlers.GetMenu)
		menu.Put("/:id", api_handlers.UpdateMenuItem)
		menu.Put("/image/:id", api_handlers.UpdateMenuImage)
		menu.Delete("/:id", api_handlers.SoftDelete_Menu)

		// Option Groups
		menu.Post("/option-groups", api_handlers.AddMoreGroup)
		menu.Put("/option-groups/:id", api_handlers.UpdateOptionGroup)

		// Options
		menu.Post("/options", api_handlers.AddMoreMenuOption)
		menu.Put("/options/:id", api_handlers.UpdateOption)
		menu.Delete("/options/:id", api_handlers.SoftDelete_Option)

		// Deleted Items Management
		menu.Get("/deleted", api_handlers.GetDeletedMenus) //ดูเมน฿ที่ถูก softdelete
		menu.Post("/restore/:id", api_handlers.RestoreMenu)
		menu.Post("/restore-group/:id", api_handlers.RestoreOptionGroup)
		menu.Post("/restore-option/:id", api_handlers.RestoreOption)
	}

	// Category Management Routes - ต้องการการยืนยันตัวตน และต้องเป็น manager
	// categories := api.Group("/categories", utils.AuthRequired(), utils.RoleRequired(models.RoleManager))
	categories := api.Group("/categories")
	{
		categories.Post("/", api_handlers.CreateCategoryHandler)
		categories.Get("/", api_handlers.GetCategoriesHandler)
		categories.Put("/:id", api_handlers.UpdateCategoryHandler)
		categories.Delete("/:id", api_handlers.Delete_categoryHandler)
		categories.Post("/restore_categories/:id", api_handlers.Restore_categoryHandler)
		categories.Get("/get_delete_categories", api_handlers.Get_Delete_Cat)
	}

	// Order Management Routes
	orders := api.Group("/orders")
	{
		// สำหรับลูกค้า (ไม่ต้องการการยืนยันตัวตน)
		orders.Post("/", api_handlers.Order_test) //สั่งอาหาร

		// สำหรับพนักงาน (ต้องการการยืนยันตัวตน)
		// orderStaff := orders.Group("/", utils.AuthRequired())
		// {
		// 	orderStaff.Get("/", api_handlers.GetOrders) // ต้องสร้างฟังก์ชันนี้เพิ่ม
		// 	orderStaff.Put("/:id/status", api_handlers.UpdateOrderStatus)// ต้องสร้างฟังก์ชันนี้เพิ่ม
		// }
	}

	table := api.Group("/table")
	{
		table.Post("/", api_handlers.Addtable)
	}

	// QR Code Management Routes
	// qr := api.Group("/qr", utils.AuthRequired(), utils.RoleRequired(models.RoleStaff, models.RoleManager))
	qr := api.Group("/qr", utils.AuthRequired())
	{
		qr.Get("/:table", qr_service.HandleQRCodeRequest)
		qr.Get("/tables", qr_service.Table)
	}
	SetupUserRoutes(app)
}