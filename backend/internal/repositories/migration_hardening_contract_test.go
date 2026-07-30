package repositories

import (
	"os"
	"strings"
	"testing"
)

func TestBootstrapRepairMigrationIsFailClosedAndTraceable(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../../database/migrations/007_repair_bootstrap_group_policy.sql")
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)
	for _, expected := range []string{
		"SET QUOTED_IDENTIFIER ON;",
		"repair-bootstrap-group-policy-v2|source=bootstrap-19000101",
		"8a33d6dbedf56c20dbb857c3579ed986ccfce686afb60278dff06190aafe6ed3",
		"effective_from = CONVERT(DATETIME2(0), '19000101', 112)",
		"(SELECT COUNT(*) FROM dbo.reservation_policies WITH (HOLDLOCK)) <> 1",
		"reservation_policy_scope_migrations",
		"resource_id NOT IN (1, 2, 3, 4, 5, 6, 7, 8)",
		"resource_id NOT IN (1, 2, 7)",
		"@replayed AS replayed",
		"AS already_correct",
	} {
		if !strings.Contains(source, expected) {
			t.Fatalf("007 lost fail-closed obligation %q", expected)
		}
	}
}

func TestPersonalOverlapMigrationProtectsCreateAndConfirm(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../../database/migrations/008_personal_overlap_includes_participations.sql")
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)
	for _, expected := range []string{
		"SET QUOTED_IDENTIFIER ON;",
		"trg_reservations_validate_participant_overlap",
		"trg_participants_validate_personal_overlap",
		"membership.status = 'CONFIRMED'",
		"other_membership.status = 'CONFIRMED'",
		"joined_reservation.start_time < DATEADD(MINUTE, existing.duration_minutes, existing.start_time)",
		"DATEADD(MINUTE, joined_reservation.duration_minutes, joined_reservation.start_time) > existing.start_time",
		"THROW 51023",
	} {
		if !strings.Contains(source, expected) {
			t.Fatalf("008 lost personal overlap obligation %q", expected)
		}
	}
}
