package services

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestSEC005DatabaseReservationErrorMapping(t *testing.T) {
	t.Run("SQLSTATE desconocido conserva error interno", func(t *testing.T) {
		original := &pgconn.PgError{
			Code:    "XX999",
			Message: "password=SECRET tabla interna",
		}

		got := mapDatabaseReservationError(original)

		if got != original {
			t.Fatal("se perdio el error PostgreSQL original")
		}
	})

	t.Run("politica ausente conserva mensaje publico", func(t *testing.T) {
		original := &pgconn.PgError{
			Code:    "P1008",
			Message: "mensaje tecnico privado",
		}

		got := mapDatabaseReservationError(original)

		if got.Error() != "no existe una politica de reservas vigente" {
			t.Fatalf("mensaje inesperado: %q", got.Error())
		}
	})

	t.Run("bloqueo de recurso conserva mensaje conocido", func(t *testing.T) {
		original := &pgconn.PgError{
			Code:    "P1007",
			Message: "informacion SQL privada",
		}

		got := mapDatabaseReservationError(original)

		if got.Error() != "el recurso esta bloqueado en ese horario" {
			t.Fatalf("mensaje inesperado: %q", got.Error())
		}
	})
}
