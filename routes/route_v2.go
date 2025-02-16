package routes

import (
	"food-ordering-api/api_v2"
	utils "food-ordering-api/utility"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutesV2(app *fiber.App) {
	api := app.Group("/api/v2")

	payment := api.Group("/payment", utils.POSAuthRequired())
	{
		payment.Post("/merge", api_v2.CreateMergedReceipt)
	}

	reservation := api.Group("/reservation", utils.POSAuthRequired())
	{
		reservation.Post("/cancel/:id", api_v2.CancelReservation)
		reservation.Post("/checkin/:id", api_v2.CheckinReservedTable)
	}

	printer := api.Group("/printers")
	{
		// งานพิมพ์ที่รอดำเนินการ
		printer.Get("/pending-jobs", utils.PrinterAPIKeyMiddleware(), api_v2.GetPendingPrintJobs)

		// รีพริ้นเอกสาร
		printer.Post("/reprint/:id", utils.POSAuthRequired(), api_v2.ReprintDocument)

		// ดึงรายการงานพิมพ์ที่สามารถรีพริ้นได้
		printer.Get("/reprintable-jobs", utils.POSAuthRequired(), api_v2.GetReprintableJobs)

		// พิมพ์ใบรายการอาหารก่อนชำระเงิน
		// printer.Post("/bill-check", api_v2.PrintBillCheck)
	}
}
