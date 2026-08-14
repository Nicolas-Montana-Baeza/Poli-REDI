package services

import (
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapPostgresReservationOverlapErrors(t *testing.T) {
	tests := []struct {
		constraint string
		want       string
	}{
		{"ex_reservations_user_overlap", "usuario ya tiene"},
		{"ex_reservations_resource_overlap", "recurso ya esta reservado"},
	}
	for _, test := range tests {
		err := mapDatabaseReservationError(&pgconn.PgError{Code: "23P01", ConstraintName: test.constraint})
		if err == nil || !strings.Contains(err.Error(), test.want) {
			t.Fatalf("constraint %s mapped to %v", test.constraint, err)
		}
	}
}

func TestMapPostgresReservationBusinessCodes(t *testing.T) {
	for code, want := range map[string]string{
		"P1006": "ventana reservable",
		"P1007": "bloqueado",
		"P1008": "politica",
	} {
		err := mapDatabaseReservationError(&pgconn.PgError{Code: code})
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("SQLSTATE %s mapped to %v", code, err)
		}
	}
}
