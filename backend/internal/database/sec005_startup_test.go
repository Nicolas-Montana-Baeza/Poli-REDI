package database

import (
	"strings"
	"testing"
)

func TestSEC005MalformedDatabaseURLHidesPassword(t *testing.T) {
	clearDatabaseEnvironment(t)

	t.Setenv(
		"DATABASE_URL",
		"postgres://app:SEC005_PASSWORD%ZZ@db.invalid/poliredi",
	)

	_, err := connectionStringFromEnv()

	if err == nil {
		t.Fatal("se esperaba un error de URL")
	}

	if err.Error() != "DATABASE_URL invalida" {
		t.Fatalf("mensaje inesperado: %q", err.Error())
	}

	if strings.Contains(err.Error(), "SEC005_PASSWORD") {
		t.Fatal("la contrasena aparecio en el error")
	}
}
