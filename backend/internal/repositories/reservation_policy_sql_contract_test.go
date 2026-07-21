package repositories

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSeedBootstrapsInitialPolicyOnceWithoutDisablingTriggers(t *testing.T) {
	seed := readRepositoryFile(t, "database", "seed.sql")
	upper := strings.ToUpper(seed)
	for _, forbidden := range []string{"DISABLE TRIGGER", "ENABLE TRIGGER"} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("seed.sql changes trigger state: %s", forbidden)
		}
	}
	if strings.Count(upper, "INSERT INTO DBO.RESERVATION_POLICY_RESOURCES") != 1 {
		t.Fatal("seed must contain exactly one guarded bootstrap association")
	}
	resourceSeed := strings.Index(upper, "INSERT INTO DBO.RESOURCES")
	bootstrap := strings.Index(upper, "INSERT INTO DBO.RESERVATION_POLICY_RESOURCES")
	marker := strings.Index(upper, "INSERT INTO DBO.RESERVATION_POLICY_SCOPE_MIGRATIONS")
	if resourceSeed < 0 || bootstrap <= resourceSeed || marker <= bootstrap {
		t.Fatal("bootstrap must run after resources and mark the policy after association")
	}
	for _, required := range []string{"BEGIN TRANSACTION", "COMMIT TRANSACTION", "ROLLBACK TRANSACTION", "LEGACY_POLICY_SCOPE_BOOTSTRAP", "IDEMPOTENCY_KEY IS NULL", "NOT EXISTS"} {
		if !strings.Contains(upper, required) {
			t.Fatalf("seed bootstrap lacks guard: %s", required)
		}
	}
}

func TestSchemaKeepsBootstrapExceptionNarrow(t *testing.T) {
	schema := strings.ToUpper(readRepositoryFile(t, "database", "schema.sql"))
	for _, required := range []string{
		"RESERVATION_POLICY_SCOPE_MIGRATIONS", "LEGACY_POLICY_SCOPE_BOOTSTRAP",
		"IDEMPOTENCY_KEY IS NULL", "ORDER BY EFFECTIVE_FROM, ID", "NOT EXISTS",
	} {
		if !strings.Contains(schema, required) {
			t.Fatalf("schema.sql lacks guarded migration element: %s", required)
		}
	}
	if strings.Contains(schema, "DISABLE TRIGGER") {
		t.Fatal("schema.sql disables an immutability trigger")
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
