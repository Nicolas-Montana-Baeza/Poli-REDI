package repositories

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPostgresSeedIsRepeatableAndMatchesMVP1Durations(t *testing.T) {
	seed := strings.ToUpper(readRepositoryFile(t, "database", "postgres", "seed", "PG16_seed_mvp1.sql"))
	for _, required := range []string{
		"ON CONFLICT", "NOT EXISTS", "MVP1-LOCAL-BASELINE-20260813",
		"(150)", "(180)",
	} {
		if !strings.Contains(seed, required) {
			t.Fatalf("PostgreSQL seed lacks %s", required)
		}
	}
	if strings.Contains(seed, "DBO.") || strings.Contains(seed, "DISABLE TRIGGER") {
		t.Fatal("PostgreSQL seed contains legacy or unsafe SQL")
	}
}

func TestPostgresSchemaEnforcesMVP1ReservationInvariants(t *testing.T) {
	baseline := strings.ToUpper(readRepositoryFile(t, "database", "postgres", "migrations", "PG16_0001_mvp1_baseline.sql"))
	indexes := strings.ToUpper(readRepositoryFile(t, "database", "postgres", "migrations", "PG16_0002_mvp1_indexes.sql"))
	invariants := strings.ToUpper(readRepositoryFile(t, "database", "postgres", "migrations", "PG16_0003_mvp1_invariants.sql"))

	for _, required := range []string{"END_TIME TIMESTAMPTZ NOT NULL", "RESERVATION_MODE_SNAPSHOT", "AVAILABILITY_BLOCKS"} {
		if !strings.Contains(baseline, required) {
			t.Fatalf("baseline lacks %s", required)
		}
	}
	for _, required := range []string{"EX_RESERVATIONS_RESOURCE_OVERLAP", "EX_RESERVATIONS_USER_OVERLAP", "RESERVATION_MODE_SNAPSHOT = 'RESERVABLE'"} {
		if !strings.Contains(indexes, required) {
			t.Fatalf("indexes lack %s", required)
		}
	}
	for _, required := range []string{"PG_ADVISORY_XACT_LOCK", "P1001", "P1007", "P1009", "AMERICA/SANTIAGO"} {
		if !strings.Contains(invariants, required) {
			t.Fatalf("invariants lack %s", required)
		}
	}
}

func readRepositoryFile(t *testing.T, parts ...string) string {
	t.Helper()
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test file")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(current), "..", "..", ".."))
	data, err := os.ReadFile(filepath.Join(append([]string{root}, parts...)...))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
