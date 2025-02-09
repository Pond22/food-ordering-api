package routes

import (
	"food-ordering-api/api_v2"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutesV2(app *fiber.App) {
	api := app.Group("/api/v2")

	payment := api.Group("/payment")
	{
		payment.Post("/merge", api_v2.CreateMergedReceipt)
	}

	reservation := api.Group("/reservation")
	{
		reservation.Post("/cancel/:id", api_v2.CancelReservation)
	}
}
