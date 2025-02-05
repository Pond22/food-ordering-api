package routes

import (
	"food-ordering-api/api_handlers"
	"food-ordering-api/models"
	qr_service "food-ordering-api/services"
	utils "food-ordering-api/utility"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func SetupRoutes(app *fiber.App) {
	// API Group
	api := app.Group("/api")

	notifications := api.Group("/notifications")
	{
		notifications.Post("/call-staff", api_handlers.CallStaff)
	}
	// WebSocket สำหรับพนักงาน
	app.Get("/ws/staff", websocket.New(api_handlers.StaffWebSocket))

	pos := api.Group("/pos")
	{
		pos.Post("/sessions/start", api_handlers.StartPOSSession)
		pos.Post("sessions/:id/end", api_handlers.EndPOSSession)
		pos.Get("/sessions/:id/validate", api_handlers.ValidatePOSSession)
	}

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
		menu.Get("/ActiveMenu", api_handlers.GetActiveMenu)
		menu.Get("/", api_handlers.GetMenu)
		menu.Put("/:id", api_handlers.UpdateMenuItem)
		menu.Put("/image/:id", api_handlers.UpdateMenuImage)
		menu.Delete("/:id", api_handlers.SoftDelete_Menu)

		// Option Groups
		menu.Get("/option-groups/:id", api_handlers.GetOptionByid)
		menu.Post("/option-groups", api_handlers.AddMoreGroup)
		menu.Put("/option-groups/:id", api_handlers.UpdateOptionGroup)

		// Options
		menu.Post("/options", api_handlers.AddMoreMenuOption)
		app.Put("/api/menu/:menu_id/options/:option_id", api_handlers.UpdateOptionByMenuID) //อัปเดตผ่านไอดีัอาหาร
		menu.Put("/options/:id", api_handlers.UpdateOption)
		menu.Delete("/options/:id", api_handlers.SoftDelete_Option)

		// Deleted Items Management
		menu.Get("/deleted", api_handlers.GetDeletedMenus) //ดูเมน฿ที่ถูก softdelete
		menu.Post("/restore/:id", api_handlers.RestoreMenu)
		menu.Post("/restore-group/:id", api_handlers.RestoreOptionGroup)
		menu.Post("/restore-option/:id", api_handlers.RestoreOption)

		menu.Put("/status/:id", api_handlers.UpdateMenuStatus)
	}

	promotion := api.Group("/promotions")
	{
		promotion.Post("/", api_handlers.CreatePromotion)
		promotion.Get("/", api_handlers.GetAllPromotion)
		promotion.Get("/Active", api_handlers.GetActivePromotions)
		promotion.Patch("/status/:id", api_handlers.UpdatePromotionStatus)
		promotion.Put("/image/:id", api_handlers.UpdatePromotionImage)
		promotion.Put("/:id", api_handlers.UpdatePromotion)
		promotion.Delete("/:id", api_handlers.DeletePromotion)
		promotion.Get("/:id", api_handlers.GetPromotionByID)
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
		orders.Post("/", api_handlers.CreateOrder) //สั่งอาหารa
		orders.Post("/items/cancel", api_handlers.CancelOrderItem)
		orders.Put("/status/:id", api_handlers.UpdateOrderStatus) //สั่งอาหารa
		orders.Post("/items/serve/:id", api_handlers.ServeOrderItem)
		orders.Get("/active", api_handlers.GetActiveOrders)
		// สำหรับพนักงาน (ต้องการการยืนยันตัวตน)
		// orderStaff := orders.Group("/", utils.AuthRequired())
		// {
		// 	orderStaff.Get("/", api_handlers.GetOrders) // ต้องสร้างฟังก์ชันนี้เพิ่ม
		// 	orderStaff.Put("/:id/status", api_handlers.UpdateOrderStatus)// ต้องสร้างฟังก์ชันนี้เพิ่ม
		// }
	}

	table := api.Group("/table")
	{
		table.Get("/orders/:id", api_handlers.GetTableOrders)
		table.Post("/", api_handlers.Addtable)
		table.Delete("/:id", api_handlers.DeleteTable)
		table.Put("/:id", api_handlers.UpdateTable)
		table.Post("/mergeTable", api_handlers.MergeTables)
		table.Post("/moveTable", api_handlers.MoveTable)
		table.Post("/splitTable/:id", api_handlers.SplitTables)
		table.Post("/reservedTable/:id", api_handlers.ReservedTable)
		table.Post("/unreservedTable/:id", api_handlers.UnreservedTable)
		table.Put("/setstatus/:id", api_handlers.ToggleTableStatus)
	}

	// QR Code Management Routes
	// qr := api.Group("/qr", utils.AuthRequired(), utils.RoleRequired(models.RoleStaff, models.RoleManager))
	qr := api.Group("/qr")
	{
		qr.Get("/:id", qr_service.HandleQRCodeRequest)
		qr.Get("/tables", qr_service.Table)
	}

	printer := api.Group("/printers")
	{
		// printer.Get("/printer-categories", api_handlers.GetPrinterCategories)
		// printer.Put("/printer-categories/:id", api_handlers.UpdatePrinterCategory)
		// printer.Delete("/printer-categories/:id", api_handlers.DeletePrinterCategory)

		printer.Post("/categories/:id", api_handlers.AssignPrinterCategories)
		printer.Get("/categories/:id", api_handlers.GetPrinterCategoriesById)

		// printer.Get("/:id", api_handlers.GetPrinterByID)
		printer.Get("/", api_handlers.GetAllPrinters)

		printer.Get("/pending-jobs", api_handlers.GetPendingPrintJobs)
		printer.Put("/status/:id", api_handlers.UpdatePrintJobStatus)
	}

	payment := api.Group("/payment")
	{
		// การชำระเงินและใบเสร็จ
		payment.Post("/process", api_handlers.ProcessPayment) // ชำระเงิน
		payment.Get("/receipt/:id", api_handlers.GetReceipt)  // ดึงข้อมูลใบเสร็จ

		// จัดการประเภทส่วนลด
		discountTypes := payment.Group("/discount-types")
		{
			discountTypes.Get("/active", api_handlers.GetActiveDiscountTypes) // ดึงเฉพาะที่เปิดใช้งาน
			discountTypes.Get("/", api_handlers.GetAllDiscountTypes)          // ดึงทั้งหมด
			discountTypes.Get("/:id", api_handlers.GetDiscountType)           // ดึงตาม ID
			discountTypes.Post("/", api_handlers.CreateDiscountType)          // สร้างใหม่
			discountTypes.Put("/:id", api_handlers.UpdateDiscountType)        // แก้ไข
			discountTypes.Delete("/:id", api_handlers.DeleteDiscountType)     // ลบ/ปิดใช้งาน
		}

		// จัดการประเภทค่าใช้จ่ายเพิ่มเติม
		chargeTypes := payment.Group("/charge-types")
		{
			chargeTypes.Get("/active", api_handlers.GetActiveChargeTypes) // ดึงเฉพาะที่เปิดใช้งาน
			chargeTypes.Get("/", api_handlers.GetAllChargeTypes)          // ดึงทั้งหมด
			chargeTypes.Get("/:id", api_handlers.GetChargeType)           // ดึงตาม ID
			chargeTypes.Post("/", api_handlers.CreateChargeType)          // สร้างใหม่
			chargeTypes.Put("/:id", api_handlers.UpdateChargeType)        // แก้ไข
			chargeTypes.Delete("/:id", api_handlers.DeleteChargeType)     // ลบ/ปิดใช้งาน
		}
	}
	// payment := api.Group("/payment")
	// {
	// 	// การชำระเงินและใบเสร็จ
	// 	payment.Post("/process", api_handlers.ProcessPayment) // ชำระเงิน
	// 	payment.Get("/receipt/:id", api_handlers.GetReceipt)  // ดึงข้อมูลใบเสร็จ

	// 	// จัดการประเภทส่วนลด
	// 	discountTypes := payment.Group("/discount-types")
	// 	{
	// 		discountTypes.Get("/active", api_handlers.GetActiveDiscountTypes) // ดึงเฉพาะที่เปิดใช้งาน
	// 		discountTypes.Get("/", api_handlers.GetAllDiscountTypes)          // ดึงทั้งหมด
	// 		discountTypes.Get("/:id", api_handlers.GetDiscountType)           // ดึงตาม ID
	// 		discountTypes.Post("/", api_handlers.CreateDiscountType,
	// 			utils.RoleRequired(models.RoleManager)) // สร้างใหม่ (ผู้จัดการเท่านั้น)
	// 		discountTypes.Put("/:id", api_handlers.UpdateDiscountType,
	// 			utils.RoleRequired(models.RoleManager)) // แก้ไข (ผู้จัดการเท่านั้น)
	// 		discountTypes.Delete("/:id", api_handlers.DeleteDiscountType,
	// 			utils.RoleRequired(models.RoleManager)) // ลบ/ปิดใช้งาน (ผู้จัดการเท่านั้น)
	// 	}

	// 	// จัดการประเภทค่าใช้จ่ายเพิ่มเติม
	// 	chargeTypes := payment.Group("/charge-types")
	// 	{
	// 		chargeTypes.Get("/active", api_handlers.GetActiveChargeTypes) // ดึงเฉพาะที่เปิดใช้งาน
	// 		chargeTypes.Get("/", api_handlers.GetAllChargeTypes)          // ดึงทั้งหมด
	// 		chargeTypes.Get("/:id", api_handlers.GetChargeType)           // ดึงตาม ID
	// 		chargeTypes.Post("/", api_handlers.CreateChargeType,
	// 			utils.RoleRequired(models.RoleManager)) // สร้างใหม่ (ผู้จัดการเท่านั้น)
	// 		chargeTypes.Put("/:id", api_handlers.UpdateChargeType,
	// 			utils.RoleRequired(models.RoleManager)) // แก้ไข (ผู้จัดการเท่านั้น)
	// 		chargeTypes.Delete("/:id", api_handlers.DeleteChargeType,
	// 			utils.RoleRequired(models.RoleManager)) // ลบ/ปิดใช้งาน (ผู้จัดการเท่านั้น)
	// 	}
	// }
	SetupUserRoutes(app)
	SetupRoutesV2(app)
}
