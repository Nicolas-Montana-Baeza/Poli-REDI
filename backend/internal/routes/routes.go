package routes

import (
	"poli-redi-api/internal/handlers"
	"poli-redi-api/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Get("/health", handlers.GetHealth)

	protected := api.Group("", middleware.RequireAuth())
	protected.Get("/me", handlers.GetMe)

	protected.Post("/resources/list", handlers.GetResources)

	protected.Post("/reservations/list", handlers.GetReservations)
	protected.Post("/reservations/create", handlers.CreateReservation)
	protected.Patch("/reservations/cancel", handlers.CancelReservation)
}
