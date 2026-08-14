package database

import (
	"net/url"
	"testing"
)

func clearDatabaseEnvironment(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"DATABASE_URL",
		"PGHOST",
		"PGPORT",
		"PGDATABASE",
		"PGUSER",
		"PGPASSWORD",
		"PGSSLMODE",
	} {
		t.Setenv(key, "")
	}
}

func TestConnectionStringPrefersDatabaseURL(t *testing.T) {
	clearDatabaseEnvironment(t)
	want := "postgres://custom:secret@db.internal:5433/custom?sslmode=require"
	t.Setenv("DATABASE_URL", want)
	t.Setenv("PGPASSWORD", "ignored")

	got, err := connectionStringFromEnv()
	if err != nil {
		t.Fatalf("connectionStringFromEnv() error = %v", err)
	}
	if got != want {
		t.Fatalf("connectionStringFromEnv() = %q, want %q", got, want)
	}
}

func TestConnectionStringBuildsEscapedPostgresURL(t *testing.T) {
	clearDatabaseEnvironment(t)
	t.Setenv("PGHOST", "localhost")
	t.Setenv("PGPORT", "5432")
	t.Setenv("PGDATABASE", "poliredi")
	t.Setenv("PGUSER", "poliredi_app")
	t.Setenv("PGPASSWORD", "local:p@ss/word")
	t.Setenv("PGSSLMODE", "disable")

	raw, err := connectionStringFromEnv()
	if err != nil {
		t.Fatalf("connectionStringFromEnv() error = %v", err)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	password, present := parsed.User.Password()
	if !present || password != "local:p@ss/word" {
		t.Fatalf("password = %q, present = %v", password, present)
	}
	if parsed.Host != "localhost:5432" || parsed.Path != "/poliredi" {
		t.Fatalf("URL PostgreSQL inesperada: %q", raw)
	}
}

func TestConnectionStringRequiresPassword(t *testing.T) {
	clearDatabaseEnvironment(t)

	if _, err := connectionStringFromEnv(); err == nil {
		t.Fatal("connectionStringFromEnv() debio rechazar PGPASSWORD vacio")
	}
}

func TestConnectionStringRejectsInvalidPort(t *testing.T) {
	clearDatabaseEnvironment(t)
	t.Setenv("PGPASSWORD", "secret")
	t.Setenv("PGPORT", "70000")

	if _, err := connectionStringFromEnv(); err == nil {
		t.Fatal("connectionStringFromEnv() debio rechazar PGPORT invalido")
	}
}

func TestConnectionStringRejectsNonPostgresURL(t *testing.T) {
	clearDatabaseEnvironment(t)
	t.Setenv("DATABASE_URL", "sqlserver://user:secret@localhost/poliredi")

	if _, err := connectionStringFromEnv(); err == nil {
		t.Fatal("connectionStringFromEnv() debio rechazar otro motor")
	}
}
