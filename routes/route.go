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
		notifications.Post("/call", api_handlers.HandleCall) // สำหรับเรียกพนักงาน
		notifications.Post("/:id/read", api_handlers.MarkAsRead)
		notifications.Get("/unread", api_handlers.GetUnreadNotifications)
	}
	// WebSocket สำหรับพนักงานรับแจ้งเตือน
	app.Get("/ws/staff", utils.WebSocketAPIKeyMiddleware("websocket_staff"), websocket.New(api_handlers.HandleWebSocket))

	pos := api.Group("/pos")
	{
		// Public endpoints (ไม่ต้องการ authentication)
		pos.Post("/verify-code", api_handlers.VerifyPOSAccessCode)
		pos.Get("/session-status", api_handlers.GetPOSSessionStatus)
		pos.Post("/generate-code", utils.AuthRequired(), api_handlers.GeneratePOSVerificationCode)
		// Protected endpoints (ต้องการ authentication)
		posAuth := pos.Group("", utils.POSAuthRequired())
		{
			// สำหรับการจัดการ POS session
			posAuth.Post("/logout", api_handlers.LogoutPOS)
			posAuth.Get("/session-status", api_handlers.GetPOSSessionStatus)
		}
	}
	// Auth Routes
	auth := api.Group("/auth")
	{
		auth.Post("/login", api_handlers.Login)
		auth.Get("/verify-token", api_handlers.VerifyToken)
	}

	// User Routes
	// user := api.Group("/member", utils.AuthRequired())
	user := api.Group("/member", utils.AuthRequired()) //เอา middleware ออก deploy อย่าลืมเอาใส่
	{
		user.Post("/", utils.RoleRequired(models.RoleManager), api_handlers.CreateUser)
		user.Get("/", api_handlers.GetUsers)
		user.Get("/get_member_profile", api_handlers.GetUserProfile)
		user.Put("/change-password", api_handlers.ChangePassword)
		user.Put("/:id/reset-password", utils.RoleRequired(models.RoleManager), api_handlers.ResetUserPassword) // ผจก. เปลี่ยนรหัสผ่านพนักงาน
		user.Delete("/:id/Delete-member", utils.RoleRequired(models.RoleManager), api_handlers.Delete_user)     // ผจก. ลบพนักงาน
	}

	// Menu Management Routes - ต้องการการยืนยันตัวตนต้อง manager
	// menu := api.Group("/menu", utils.AuthRequired(), utils.RoleRequired(models.RoleManager))
	menu := api.Group("/menu")
	{
		// เมนูพื้นฐาน
		menu.Post("/import", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.ImportMenuFromExcel)
		menu.Post("/", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.CreateMenuItemHandler)
		menu.Get("/ActiveMenu", api_handlers.GetActiveMenu) // สำหรับดึงเมนูที่เปิดใช้งาน
		menu.Get("/", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.GetMenu)
		menu.Put("/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.UpdateMenuItem)
		menu.Put("/image/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.UpdateMenuImage)
		menu.Delete("/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.SoftDelete_Menu)

		menu.Get("/recommended", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.GetRecommendedMenuItems)
		menu.Put("/:id/recommend", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.ToggleMenuItemRecommendation)

		// Option Groups
		menu.Get("/option-groups/:id", api_handlers.GetOptionByid)
		menu.Post("/option-groups", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.AddMoreGroup)
		menu.Put("/option-groups/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.UpdateOptionGroup)
		menu.Delete("/option-groups/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.SoftDelete_OptionGroup)

		// Options
		menu.Post("/options", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.AddMoreMenuOption)
		app.Put("/api/menu/:menu_id/options/:option_id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.UpdateOptionByMenuID) //อัปเดตผ่านไอดีัอาหารไม่ได้ใช้
		menu.Put("/options/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.UpdateOption)
		menu.Delete("/options/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.SoftDelete_Option)

		// Deleted Items Management
		menu.Get("/deleted", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.GetDeletedMenus) //ดูเมน฿ที่ถูก softdelete
		menu.Post("/restore/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.RestoreMenu)
		menu.Post("/restore-group/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.RestoreOptionGroup)
		menu.Post("/restore-option/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.RestoreOption)

		menu.Put("/status/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.UpdateMenuStatus)
	}

	promotion := api.Group("/promotions")
	{
		promotion.Post("/", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.CreatePromotion)
		promotion.Get("/", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.GetAllPromotion)
		promotion.Get("/Active", api_handlers.GetActivePromotions) // สำหรับดึงโปรโมชั่นที่เปิดใช้งาน
		promotion.Patch("/status/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.UpdatePromotionStatus)
		promotion.Put("/image/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.UpdatePromotionImage)
		promotion.Put("/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.UpdatePromotion)
		promotion.Delete("/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.DeletePromotion)
		promotion.Get("/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.GetPromotionByID)
		promotion.Delete("/:id/items/:item_id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.DeletePromotionItem)
		promotion.Post("/:id/items", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.AddPromotionItems)
	}

	// Category Management Routes - ต้องการการยืนยันตัวตน และต้องเป็น manager
	// categories := api.Group("/categories", utils.AuthRequired(), utils.RoleRequired(models.RoleManager))
	categories := api.Group("/categories")
	{
		categories.Post("/", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.CreateCategoryHandler)
		categories.Get("/", api_handlers.GetCategoriesHandler) // สำหรับดึงหมวดหมู่อาหาร
		categories.Put("/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.UpdateCategoryHandler)
		categories.Delete("/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.Delete_categoryHandler)
		categories.Post("/restore_categories/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.Restore_categoryHandler)
		categories.Get("/get_delete_categories", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.Get_Delete_Cat)
	}

	// Order Management Routes
	orders := api.Group("/orders")
	{
		orders.Post("/", api_handlers.CreateOrder) // สำหรับสั่งอาหาร
		orders.Post("/items/cancel", utils.AuthRequired(), api_handlers.CancelOrderItem)
		orders.Put("/status/:id", utils.AuthRequired(), api_handlers.UpdateOrderStatus) //สั่งอาหารa
		orders.Post("/items/serve/:id", utils.AuthRequired(), api_handlers.ServeOrderItem)
		orders.Get("/active", utils.POSAuthRequired(), api_handlers.GetActiveOrders)
		orders.Get("/table/:uuid", api_handlers.GetOrdersByid) // สำหรับดูรายการอาหารที่สั่งของโต๊ะ
		orders.Post("/finalize", utils.POSAuthRequired(), api_handlers.FinalizeOrderItems)
		orders.Get("/cancellation-logs", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.GetCancellationLogs)
		// สำหรับพนักงาน (ต้องการการยืนยันตัวตน)
		// orderStaff := orders.Group("/", utils.AuthRequired())
		// {
		// 	orderStaff.Get("/", api_handlers.GetOrders) // ต้องสร้างฟังก์ชันนี้เพิ่ม
		// 	orderStaff.Put("/:id/status", api_handlers.UpdateOrderStatus)// ต้องสร้างฟังก์ชันนี้เพิ่ม
		// }
	}

	table := api.Group("/table")
	{
		table.Get("/reservations", utils.POSAuthRequired(), api_handlers.GetAllReservations)
		table.Get("/billable/:uuid", utils.POSAuthRequired(), api_handlers.GetBillableItems)
		table.Post("/", utils.POSAuthRequired(), api_handlers.Addtable)
		table.Delete("/:id", utils.POSAuthRequired(), api_handlers.DeleteTable)
		table.Put("/:id", utils.POSAuthRequired(), api_handlers.UpdateTable)
		table.Post("/mergeTable", utils.POSAuthRequired(), api_handlers.MergeTables)
		table.Post("/moveTable", utils.POSAuthRequired(), api_handlers.MoveTable)
		table.Post("/splitTable/:id", utils.POSAuthRequired(), api_handlers.SplitTables)
		table.Post("/reservedTable/:id", utils.POSAuthRequired(), api_handlers.ReservedTable)
		table.Post("/unreservedTable/:id", utils.POSAuthRequired(), api_handlers.UnreservedTable)
		table.Put("/setstatus/:id", utils.POSAuthRequired(), api_handlers.ToggleTableStatus)
		table.Post("/close/:id", utils.POSAuthRequired(), api_handlers.CloseTable)
	}

	// QR Code Management Routes
	// qr := api.Group("/qr", utils.AuthRequired(), utils.RoleRequired(models.RoleStaff, models.RoleManager))
	qr := api.Group("/qr", utils.POSAuthRequired())
	{
		qr.Post("/reprint/:id", qr_service.HandleQRCodeReprint)
		qr.Get("/:id", qr_service.HandleQRCodeRequest)
		qr.Get("/tables", qr_service.Table)
	}

	printer := api.Group("/printers")
	{
		// printer.Get("/printer-categories", api_handlers.GetPrinterCategories)
		// printer.Put("/printer-categories/:id", api_handlers.UpdatePrinterCategory)
		// printer.Delete("/printer-categories/:id", api_handlers.DeletePrinterCategory)

		printer.Post("/categories/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.AssignPrinterCategories)
		printer.Get("/categories/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.GetPrinterCategoriesById)

		printer.Post("reprint/:id", utils.POSAuthRequired(), api_handlers.ReprintDocument)
		printer.Get("failed-jobs", utils.POSAuthRequired(), api_handlers.GetFailedPrintJobs)
		printer.Get("/reprintable-jobs", utils.POSAuthRequired(), api_handlers.GetReprintableJobs)
		printer.Get("/", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.GetAllPrinters)
		printer.Post("/bill-check", utils.POSAuthRequired(), api_handlers.PrintBillCheck)

		// เฉพาะ endpoints ที่ต้องการใช้ API key
		printer.Get("/pending-jobs", utils.PrinterAPIKeyMiddleware(), api_handlers.GetPendingPrintJobs)
		printer.Put("/status/:id", utils.PrinterAPIKeyMiddleware(), api_handlers.UpdatePrintJobStatus)
	}

	payment := api.Group("/payment")
	{
		// การชำระเงินและใบเสร็จ
		payment.Post("/process", utils.POSAuthRequired(), api_handlers.ProcessPayment) // ชำระเงิน
		payment.Get("/receipt/:id", utils.POSAuthRequired(), api_handlers.GetReceipt)  // ดึงข้อมูลใบเสร็จ

		// จัดการประเภทส่วนลด
		discountTypes := payment.Group("/discount-types")
		{
			discountTypes.Get("/active", utils.POSAuthRequired(), api_handlers.GetActiveDiscountTypes)                                  // ดึงเฉพาะที่เปิดใช้งาน
			discountTypes.Get("/", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.GetAllDiscountTypes)      // ดึงทั้งหมด
			discountTypes.Get("/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.GetDiscountType)       // ดึงตาม ID
			discountTypes.Post("/", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.CreateDiscountType)      // สร้างใหม่
			discountTypes.Put("/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.UpdateDiscountType)    // แก้ไข
			discountTypes.Delete("/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.DeleteDiscountType) // ลบ/ปิดใช้งาน
		}

		// จัดการประเภทค่าใช้จ่ายเพิ่มเติม
		chargeTypes := payment.Group("/charge-types")
		{
			chargeTypes.Get("/active", utils.POSAuthRequired(), api_handlers.GetActiveChargeTypes)                                  // ดึงเฉพาะที่เปิดใช้งาน
			chargeTypes.Get("/", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.GetAllChargeTypes)      // ดึงทั้งหมด
			chargeTypes.Get("/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.GetChargeType)       // ดึงตาม ID
			chargeTypes.Post("/", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.CreateChargeType)      // สร้างใหม่
			chargeTypes.Put("/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.UpdateChargeType)    // แก้ไข
			chargeTypes.Delete("/:id", utils.AuthRequired(), utils.RoleRequired(models.RoleManager), api_handlers.DeleteChargeType) // ลบ/ปิดใช้งาน
		}

		reservation := api.Group("/reservation", utils.AuthRequired(), utils.RoleRequired(models.RoleManager))
		{
			reservation.Get("/rules/active", api_handlers.GetActiveReservationRule)
			reservation.Post("/rules", api_handlers.SetReservationRules)
			reservation.Get("/rules/history", api_handlers.GetReservationRulesHistory)
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
