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
	protected.Patch("/me/rut", handlers.UpdateMeRUT)

	protected.Get("/resources", handlers.GetResources)
	protected.Get("/activities", handlers.GetActivities)
	protected.Get("/notifications", handlers.GetNotifications)
	protected.Get("/workshops", handlers.GetWorkshops)
	protected.Post("/workshops/:id/enroll", handlers.EnrollInWorkshop)
	protected.Get("/availability/reservations", handlers.GetAvailabilityReservations)

	protected.Get(
		"/users",
		middleware.RequireAdmin(),
		handlers.GetUsers,
	)

	protected.Get(
		"/reservations",
		middleware.RequireAdmin(),
		handlers.GetReservations,
	)

	protected.Get("/reservations/mine", handlers.GetMyReservations)
	protected.Post("/reservations", handlers.CreateReservation)
	protected.Patch("/reservations/cancel", handlers.CancelReservation)
}
