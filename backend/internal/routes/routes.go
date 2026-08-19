package routes

import (
	"poli-redi-api/internal/appscope"
	"poli-redi-api/internal/handlers"
	"poli-redi-api/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registra la superficie HTTP disponible según el scope
// funcional configurado para la aplicación.
//
// La estrategia es incremental:
//
//   - MVP1 mantiene únicamente la superficie estable original;
//   - MVP2 agrega las funcionalidades ya migradas y validadas en PostgreSQL;
//   - FULL agrega además los módulos legacy que todavía están siendo
//     adaptados progresivamente.
//
// Esto evita que habilitar MVP2 exponga accidentalmente funcionalidades
// antiguas que aún no forman parte del bloque validado.
func RegisterRoutes(app *fiber.App) {
	app.Get("/", handlers.GetRoot)

	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	api := app.Group("/api")

	// Health permanece público para permitir monitoreo del servicio sin
	// depender de Microsoft Entra ID ni del modo de autenticación local.
	api.Get("/health", handlers.GetHealth)

	// Toda la funcionalidad de negocio requiere una identidad autenticada.
	protected := api.Group("", middleware.RequireAuth())

	// ------------------------------------------------------------
	// Superficie común MVP1+.
	// ------------------------------------------------------------

	protected.Get("/me", handlers.GetMe)
	protected.Patch("/me/rut", handlers.UpdateMeRUT)

	protected.Get("/resources", handlers.GetResources)
	protected.Get("/activities", handlers.GetActivities)

	protected.Get(
		"/availability/reservations",
		handlers.GetAvailabilityReservations,
	)

	protected.Get(
		"/reservation-policy/current",
		handlers.GetCurrentReservationPolicy,
	)

	protected.Get(
		"/reservations/mine",
		handlers.GetMyReservations,
	)

	protected.Get(
		"/reservations/:id",
		handlers.GetReservationDetail,
	)

	protected.Post(
		"/reservations",
		handlers.CreateReservation,
	)

	protected.Patch(
		"/reservations/cancel",
		handlers.CancelReservation,
	)

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

	// ------------------------------------------------------------
	// Superficie MVP2 validada.
	// ------------------------------------------------------------
	//
	// Estas rutas pertenecen al nuevo flujo de reservas grupales y solo
	// se exponen cuando el runtime declara explícitamente soporte MVP2.
	//
	// FULL también las incluye porque representa un superset funcional.

	if appscope.HasMVP2() {
		// Consulta el estado del grupo antes o después de incorporarse.
		protected.Get(
			"/reservations/join/:code",
			handlers.GetGroupReservationProgress,
		)
		// Permite al owner o administrador reemplazar el código de invitación.
		//
		// La autorización fina se realiza en la capa de servicio; conocer el ID
		// de una reserva no basta para rotar su código.
		protected.Post(
			"/reservations/:id/join-code",
			handlers.RotateGroupReservationJoinCode,
		)
		// Incorpora al usuario autenticado al grupo identificado por el
		// join code. El userID nunca proviene del cliente.
		protected.Post(
			"/reservations/join/:code",
			handlers.JoinGroupReservation,
		)

		// Retira al usuario autenticado del grupo.
		//
		// El owner no puede utilizar esta operación para abandonar la
		// reserva; debe cancelar la reserva completa.
		protected.Delete(
			"/reservations/join/:code",
			handlers.LeaveGroupReservation,
		)

		// La lista completa de participantes queda restringida al owner
		// de la reserva y a administradores.
		protected.Get(
			"/reservations/:id/participants",
			handlers.GetGroupReservationParticipants,
		)

		// ------------------------------------------------------------
		// Workshops institucionales
		// ------------------------------------------------------------
		//
		// Los talleres utilizan institutional_activities y comparten la
		// infraestructura de programación, disponibilidad y conflictos.

		protected.Get(
			"/workshops",
			handlers.GetInstitutionalWorkshops,
		)

		protected.Get(
			"/workshops/:id",
			handlers.GetInstitutionalWorkshop,
		)

		protected.Post(
			"/workshops/:id/enroll",
			handlers.EnrollInInstitutionalWorkshop,
		)

		protected.Delete(
			"/workshops/:id/enroll",
			handlers.LeaveInstitutionalWorkshop,
		)

		// ------------------------------------------------------------
		// Programación Institucional - Unidades y gestores
		// ------------------------------------------------------------
		//
		// La autorización fina permanece en la capa de servicio:
		//
		//   - administradores globales pueden crear unidades y membresías;
		//   - managers pueden consultar únicamente las unidades que gestionan
		//     y sus membresías;
		//   - MEMBER por sí solo no entrega permisos administrativos.

		protected.Get(
			"/institutional-units",
			handlers.GetInstitutionalUnits,
		)

		protected.Post(
			"/admin/institutional-units",
			middleware.RequireAdmin(),
			handlers.CreateInstitutionalUnit,
		)

		protected.Get(
			"/institutional-units/:id/memberships",
			handlers.GetInstitutionalUnitMemberships,
		)

		protected.Post(
			"/admin/institutional-units/:id/memberships",
			middleware.RequireAdmin(),
			handlers.AddInstitutionalUnitMembership,
		)

		// ------------------------------------------------------------
		// Programación Institucional - Actividades
		// ------------------------------------------------------------
		//
		// Estas rutas permiten al administrador global o MANAGER de una
		// unidad consultar y crear programación institucional.
		//
		// No se usa RequireAdmin porque un MANAGER legítimo también puede
		// programar su propia unidad. La autorización fina se resuelve en
		// services.EnsureInstitutionalUnitManager.

		protected.Get(
			"/institutional-units/:id/activities",
			handlers.GetInstitutionalActivitiesForUnit,
		)

		protected.Post(
			"/institutional-activities",
			handlers.CreateInstitutionalActivity,
		)

		// ------------------------------------------------------------
		// Programación Institucional - Resolución Administrativa
		// ------------------------------------------------------------
		//
		// Los conflictos pueden contener 2..N ocupaciones.
		//
		// Toda esta superficie queda limitada a administradores porque las
		// decisiones pueden cancelar reservas, reprogramar actividades o
		// autorizar coexistencias institucionales.

		protected.Get(
			"/admin/scheduling-conflicts",
			middleware.RequireAdmin(),
			handlers.GetSchedulingConflicts,
		)

		protected.Get(
			"/admin/scheduling-conflicts/:id",
			middleware.RequireAdmin(),
			handlers.GetSchedulingConflict,
		)

		protected.Patch(
			"/admin/scheduling-conflicts/:id/items/:itemId",
			middleware.RequireAdmin(),
			handlers.PatchSchedulingConflictItem,
		)

	}

	// MVP1 y MVP2 terminan aquí.
	//
	// Los módulos que siguen pertenecen todavía a la superficie legacy
	// completa y no deben exponerse simplemente por activar MVP2.
	if !appscope.IsFull() {
		return
	}

	// ------------------------------------------------------------
	// Superficie FULL / legacy.
	// ------------------------------------------------------------

	protected.Patch(
		"/resources/:id/image",
		middleware.RequireAdmin(),
		handlers.UpdateResourceImage,
	)

	protected.Get(
		"/notifications",
		handlers.GetNotifications,
	)

	protected.Get(
		"/admin/reservation-policies",
		middleware.RequireAdmin(),
		handlers.GetReservationPolicyHistory,
	)

	protected.Post(
		"/admin/reservation-policies",
		middleware.RequireAdmin(),
		handlers.PublishReservationPolicy,
	)
}
