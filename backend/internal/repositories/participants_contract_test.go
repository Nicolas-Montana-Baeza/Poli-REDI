package repositories

import (
	"strings"
	"testing"
)

func TestJoinCodeHashDoesNotExposeCode(t *testing.T) {
	code := "codigo-compartible-secreto"
	hash := codeHash(code)
	if hash == code || strings.Contains(hash, code) {
		t.Fatal("join code was not irreversibly represented")
	}
	if len(hash) != 64 {
		t.Fatalf("hash length=%d", len(hash))
	}
}

func TestParticipantSQLContractIsTransactionalAndServerControlled(t *testing.T) {
	schema := strings.ToUpper(readRepositoryFile(t, "database", "schema.sql"))
	for _, required := range []string{"RESERVATION_POLICY_GROUP_RESOURCES", "JOIN_CODE_HASH", "UQ_PARTICIPANTS_RESERVATION_USER", "UQ_PARTICIPANTS_OWNER", "RESERVATION_PARTICIPANT_AUDIT", "STATUS IN ('PENDING','CONFIRMED')"} {
		if !strings.Contains(schema, required) {
			t.Fatalf("schema lacks %s", required)
		}
	}
}

func TestMVP2MigrationIsSeparateProspectiveAndIdempotent(t *testing.T) {
	migration := strings.ToUpper(readRepositoryFile(t, "database", "migrations", "001_mvp2_group_participants.sql"))
	for _, required := range []string{"SET XACT_ABORT ON", "BEGIN TRANSACTION", "ROLLBACK TRANSACTION", "IF NOT EXISTS", "MIGRATION-MVP2-GROUP-V1", "GROUP_CAPACITY_SNAPSHOT", "VALUES(@NEW,1),(@NEW,2),(@NEW,7)"} {
		if !strings.Contains(migration, required) {
			t.Fatalf("migration lacks %s", required)
		}
	}
	for _, forbidden := range []string{"UPDATE DBO.RESERVATIONS SET GROUP_CAPACITY_SNAPSHOT", "INSERT INTO DBO.PARTICIPANTS SELECT"} {
		if strings.Contains(migration, forbidden) {
			t.Fatalf("migration reinterprets historical reservations: %s", forbidden)
		}
	}
}

func TestParticipantQueriesUseCapacitySnapshotTrimmedRUTAndInitialAudit(t *testing.T) {
	participants := strings.ToUpper(readRepositoryFile(t, "backend", "internal", "repositories", "participants_repository.go"))
	reservations := strings.ToUpper(readRepositoryFile(t, "backend", "internal", "repositories", "reservations_repository.go"))
	if strings.Contains(participants, "RES.CAPACITY") || !strings.Contains(participants, "GROUP_CAPACITY_SNAPSHOT") {
		t.Fatal("participant flow must use frozen capacity")
	}
	if !strings.Contains(participants, "STRINGSTRIMSPACE") && !strings.Contains(participants, "STRINGS.TRIMSPACE") {
		t.Fatal("participant RUT must reject whitespace")
	}
	for _, required := range []string{"GROUP_CAPACITY_SNAPSHOT", "REQUESTER_ADDED", "RESERVATION_PARTICIPANT_AUDIT", "WITH(UPDLOCK,HOLDLOCK)"} {
		if !strings.Contains(reservations, required) {
			t.Fatalf("creation lacks %s", required)
		}
	}
}

func TestSchemaAndMigrationInstallEquivalentMVP2TriggerObligations(t *testing.T) {
	schema := strings.ToUpper(readRepositoryFile(t, "database", "schema.sql"))
	migration := strings.ToUpper(readRepositoryFile(t, "database", "migrations", "001_mvp2_group_participants.sql"))
	required := []string{
		"TRG_RESERVATIONS_PENDING_CONFLICTS", "R.USER_ID=I.USER_ID",
		"R.STATUS IN('PENDING','CONFIRMED')", "TRG_BLOCKS_PENDING_CONFLICTS",
		"TRG_SCHEDULED_ACTIVITIES_PENDING_CONFLICTS",
		"TRG_RESERVATION_POLICY_GROUP_RESOURCES_IMMUTABLE",
		"R.CAPACITY IS NULL OR R.CAPACITY<P.MINIMUM_PARTICIPANTS",
		"R.RESERVATION_MODE='OPEN_USE'", "FK_GROUP_RESOURCES_ALLOWED",
	}
	for _, obligation := range required {
		if !strings.Contains(schema, obligation) {
			t.Fatalf("schema lacks obligation %s", obligation)
		}
		if !strings.Contains(migration, obligation) {
			t.Fatalf("migration lacks obligation %s", obligation)
		}
	}
	for _, divergent := range []string{"TRG_MVP2_PENDING", "TRG_MVP2_BLOCKS", "TRG_MVP2_ACTIVITIES"} {
		if strings.Contains(schema, divergent) || strings.Contains(migration, divergent) {
			t.Fatalf("divergent trigger remains: %s", divergent)
		}
	}
}
