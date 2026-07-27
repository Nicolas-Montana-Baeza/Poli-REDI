package validators

import (
	"strings"
	"testing"
)

func TestWorkshopEnrollmentTriggerMigrationMatchesCanonicalSchema(t *testing.T) {
	t.Parallel()
	root := "../../../"
	schema := mustReadContractFile(t, root+"database/schema.sql")
	migration := mustReadContractFile(t, root+"database/migrations/006_workshop_occurrences.sql")

	schemaTrigger := extractNamedTrigger(t, schema, "dbo.trg_workshop_enrollments_validate")
	migrationTrigger := extractNamedTrigger(t, migration, "dbo.trg_workshop_enrollments_validate")
	if schemaTrigger != migrationTrigger {
		t.Fatal("006 must contain the complete canonical workshop enrollment trigger")
	}
	for _, obligation := range []string{
		"THROW 51300", "THROW 51301", "UPDLOCK", "HOLDLOCK",
		"existing.id<>i.id", "existing_o.start_minute<target_o.end_minute",
		"target_o.start_minute<existing_o.end_minute",
		"counts.confirmed_count>w.capacity",
	} {
		if !strings.Contains(migrationTrigger, obligation) {
			t.Fatalf("workshop trigger lost obligation %q", obligation)
		}
	}
	for _, structure := range []string{
		"pk_workshop_occurrences", "fk_workshop_occurrences_workshop",
		"ck_workshop_occurrences_weekday", "ck_workshop_occurrences_minutes",
		"uq_workshop_occurrences_slot", "idx_workshop_occurrences_overlap",
		"THROW 56002", "THROW 56003", "THROW 56004",
	} {
		if !strings.Contains(migration, structure) {
			t.Fatalf("006 lacks structural obligation %q", structure)
		}
	}
}

func extractNamedTrigger(t *testing.T, content, name string) string {
	t.Helper()
	start := "CREATE OR ALTER TRIGGER " + name
	startIndex := strings.Index(content, start)
	if startIndex < 0 {
		t.Fatalf("missing %s", start)
	}
	remaining := content[startIndex:]
	endIndex := strings.Index(remaining, "\nGO\n")
	if endIndex < 0 {
		t.Fatalf("%s has no GO terminator", name)
	}
	return strings.TrimSpace(remaining[:endIndex])
}
