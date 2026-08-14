package routes

import (
	"poli-redi-api/internal/appscope"
	"poli-redi-api/internal/handlers"
	"poli-redi-api/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	app.Get("/", handlers.GetRoot)
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	api := app.Group("/api")

	api.Get("/health", handlers.GetHealth)

	protected := api.Group("", middleware.RequireAuth())

	protected.Get("/me", handlers.GetMe)
	protected.Patch("/me/rut", handlers.UpdateMeRUT)

	protected.Get("/resources", handlers.GetResources)
	protected.Get("/activities", handlers.GetActivities)
	protected.Get("/availability/reservations", handlers.GetAvailabilityReservations)
	protected.Get("/reservation-policy/current", handlers.GetCurrentReservationPolicy)
	protected.Get("/reservations/mine", handlers.GetMyReservations)
	protected.Get("/reservations/:id", handlers.GetReservationDetail)
	protected.Post("/reservations", handlers.CreateReservation)
	protected.Patch("/reservations/cancel", handlers.CancelReservation)
	protected.Get("/users", middleware.RequireAdmin(), handlers.GetUsers)
	protected.Get("/reservations", middleware.RequireAdmin(), handlers.GetReservations)

	if appscope.IsMVP1() {
		return
	}

	protected.Patch(
		"/resources/:id/image",
		middleware.RequireAdmin(),
		handlers.UpdateResourceImage,
	)
	protected.Get("/notifications", handlers.GetNotifications)
	protected.Get("/workshops", handlers.GetWorkshops)
	protected.Post("/workshops/:id/enroll", handlers.EnrollInWorkshop)
	protected.Get("/admin/reservation-policies", middleware.RequireAdmin(), handlers.GetReservationPolicyHistory)
	protected.Post("/admin/reservation-policies", middleware.RequireAdmin(), handlers.PublishReservationPolicy)

}
