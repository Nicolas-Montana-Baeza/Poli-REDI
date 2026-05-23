package routes

import (
	"poli-redi-api/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Get("/health", handlers.GetHealth)

	api.Get("/resources", handlers.GetResources)

	api.Get("/reservations", handlers.GetReservations)

	api.Post("/reservations", handlers.CreateReservation)
}
