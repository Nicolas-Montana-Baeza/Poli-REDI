package validators

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReservationTriggerMigrationMatchesCanonicalSchema(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "..")
	schema := mustReadContractFile(t, filepath.Join(root, "database", "schema.sql"))
	migration := mustReadContractFile(t, filepath.Join(
		root, "database", "migrations", "005_rut_integrity_and_admin_exemption.sql",
	))

	schemaTrigger := extractReservationTrigger(t, schema)
	migrationTrigger := extractReservationTrigger(t, migration)
	if schemaTrigger != migrationTrigger {
		t.Fatal("005 must contain the complete canonical reservation validation trigger from schema.sql")
	}

	obligations := []string{
		"THROW 51000", "THROW 51001", "THROW 51002", "THROW 51003",
		"THROW 51004", "THROW 51005", "THROW 51006", "THROW 51007",
		"THROW 51009", "THROW 51010", "THROW 51015", "THROW 51016",
		"THROW 51017",
		"inserted_resource.reservation_mode <> 'OPEN_USE'",
		"u.is_admin=0 AND NULLIF(LTRIM(RTRIM(u.rut)),'') IS NULL",
	}
	for _, obligation := range obligations {
		if !strings.Contains(migrationTrigger, obligation) {
			t.Fatalf("canonical reservation trigger lost obligation %q", obligation)
		}
	}
}

func mustReadContractFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return strings.ReplaceAll(string(content), "\r\n", "\n")
}

func extractReservationTrigger(t *testing.T, content string) string {
	t.Helper()
	const start = "CREATE OR ALTER TRIGGER dbo.trg_reservations_validate_conflicts"
	startIndex := strings.Index(content, start)
	if startIndex < 0 {
		t.Fatalf("missing %s", start)
	}
	remaining := content[startIndex:]
	endIndex := strings.Index(remaining, "\nGO\n")
	if endIndex < 0 {
		t.Fatal("reservation trigger has no GO terminator")
	}
	return strings.TrimSpace(remaining[:endIndex])
}
