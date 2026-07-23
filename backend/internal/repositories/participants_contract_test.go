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

func TestOpenUseDoesNotConsumeFrequencyButKeepsUserOverlapGuard(t *testing.T) {
	repository := strings.ToUpper(readRepositoryFile(t, "backend", "internal", "repositories", "reservations_repository.go"))
	schema := strings.ToUpper(readRepositoryFile(t, "database", "schema.sql"))
	migration := strings.ToUpper(readRepositoryFile(t, "database", "migrations", "003_open_use_frequency_scope.sql"))

	if !strings.Contains(repository, "RESOURCE.RESERVATION_MODE <> 'OPEN_USE'") {
		t.Fatal("frequency lookup must exclude OPEN_USE reservations")
	}
	for _, required := range []string{
		"INSERTED_RESOURCE.RESERVATION_MODE <> 'OPEN_USE'",
		"PREVIOUS_RESOURCE.RESERVATION_MODE <> 'OPEN_USE'",
		"EXISTING.USER_ID = I.USER_ID",
	} {
		if !strings.Contains(schema, required) {
			t.Fatalf("schema lacks OPEN_USE frequency/overlap rule: %s", required)
		}
	}
	for _, required := range []string{
		"CREATE OR ALTER TRIGGER DBO.TRG_RESERVATIONS_VALIDATE_CONFLICTS",
		"INSERTED_RESOURCE.RESERVATION_MODE <> 'OPEN_USE'",
		"PREVIOUS_RESOURCE.RESERVATION_MODE <> 'OPEN_USE'",
		"EXISTING.USER_ID = I.USER_ID",
	} {
		if !strings.Contains(migration, required) {
			t.Fatalf("migration lacks OPEN_USE frequency/overlap rule: %s", required)
		}
	}
	for _, fragile := range []string{"CHARINDEX(", "STUFF(", "SP_EXECUTESQL", "THROW 53001", "THROW 53002"} {
		if strings.Contains(migration, fragile) {
			t.Fatalf("migration must not reconstruct the stored trigger header: %s", fragile)
		}
	}
}

func TestSchemaAndMigrationInstallEquivalentMVP2TriggerObligations(t *testing.T) {
	schema := strings.ToUpper(readRepositoryFile(t, "database", "schema.sql"))
	migration := strings.ToUpper(readRepositoryFile(t, "database", "migrations", "001_mvp2_group_participants.sql"))
	required := []string{
		"TRG_RESERVATION_POLICIES_IMMUTABLE", "THROW 51011", "THROW 51012",
		"TRG_RESERVATION_POLICY_RESOURCES_IMMUTABLE", "LEGACY_POLICY_SCOPE_BOOTSTRAP", "THROW 51013",
		"TRG_RESERVATION_POLICY_DURATIONS_IMMUTABLE", "THROW 51014",
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

func TestMigrationNormalizesOldMVP1BeforeReferencingNewColumns(t *testing.T) {
	migration := strings.ToUpper(readRepositoryFile(t, "database", "migrations", "001_mvp2_group_participants.sql"))
	for _, column := range []string{
		"OPENING_MINUTE", "CLOSING_MINUTE", "SLOT_INTERVAL_MINUTES",
		"IDEMPOTENCY_KEY", "IDEMPOTENCY_PAYLOAD_HASH", "IS_PUBLISHED",
		"POLICY_ID", "JOIN_CODE_HASH", "GROUP_CAPACITY_SNAPSHOT",
	} {
		if !strings.Contains(migration, "COL_LENGTH(") || !strings.Contains(migration, column) {
			t.Fatalf("old MVP1 prerequisite is not normalized: %s", column)
		}
	}
	addColumns := strings.Index(migration, "-- FASE 1:")
	firstBatchEnd := strings.Index(migration[addColumns:], "\nGO")
	publication := strings.Index(migration, "-- FASE 6:")
	if addColumns < 0 || firstBatchEnd < 0 || publication < 0 || publication <= addColumns+firstBatchEnd {
		t.Fatal("new columns must be separated by GO and publication must be last")
	}
	for _, required := range []string{
		"EXEC(N'ALTER TABLE", "WITH VALUES", "-- PREFLIGHT:", "-- POSTCHECK",
		"ELSE IF NOT EXISTS", "NO SE MODIFICO LA VIGENCIA",
	} {
		if !strings.Contains(migration, required) {
			t.Fatalf("migration lacks recoverable old/partial-state behavior: %s", required)
		}
	}
}

func TestMigrationDoesNotPublishBeforeCanonicalTriggers(t *testing.T) {
	migration := strings.ToUpper(readRepositoryFile(t, "database", "migrations", "001_mvp2_group_participants.sql"))
	triggers := strings.Index(migration, "-- FASE 5:")
	publication := strings.Index(migration, "-- FASE 6:")
	if triggers < 0 || publication < 0 || triggers >= publication {
		t.Fatal("canonical triggers must be installed before prospective publication")
	}
	if strings.Contains(migration[:publication], "SET IS_PUBLISHED=1") {
		t.Fatal("migration publishes a policy before all prerequisites and triggers")
	}
}

func TestMigrationRecoveryGuidesAvoidOptionalColumnCompileErrors(t *testing.T) {
	guide := strings.ToUpper(readRepositoryFile(t, "database", "migrations", "README.md"))
	installation := strings.ToUpper(readRepositoryFile(t, "docs", "01-instalacion-y-ejecucion.md"))
	for _, unsafe := range []string{
		"SELECT ID,EFFECTIVE_FROM,EFFECTIVE_TO,IS_PUBLISHED",
		"SELECT ID,EFFECTIVE_FROM,EFFECTIVE_TO,IDEMPOTENCY_KEY",
	} {
		if strings.Contains(guide, unsafe) {
			t.Fatalf("recovery guide references optional columns directly: %s", unsafe)
		}
	}
	for _, safe := range []string{
		"COL_LENGTH('DBO.RESERVATION_POLICIES','IS_PUBLISHED')",
		"COL_LENGTH('DBO.RESERVATION_POLICIES','IDEMPOTENCY_KEY')",
		"FROM SYS.COLUMNS",
		"SELECT ID,EFFECTIVE_FROM,EFFECTIVE_TO",
	} {
		if !strings.Contains(guide, safe) {
			t.Fatalf("recovery guide lacks safe inspection: %s", safe)
		}
	}
	for _, required := range []string{
		"DATABASE/MIGRATIONS/README.MD", "BACKUP", "SESION NUEVA",
		"COMPATIBLE CON `GO`", "POSTCHECK", "UNICA FUENTE OPERATIVA",
	} {
		if !strings.Contains(installation, required) {
			t.Fatalf("installation guide lacks safe migration referral: %s", required)
		}
	}
}

func TestTargetParticipantsMigrationIsCumulativeAndDoesNotBackfill(t *testing.T) {
	migration := strings.ToUpper(readRepositoryFile(t, "database", "migrations", "002_mvp2_target_participants.sql"))
	for _, required := range []string{"COL_LENGTH('DBO.RESERVATIONS','TARGET_PARTICIPANTS')", "EXEC(N'ALTER TABLE", "RESERVATION_TARGET_AUDIT", "CREATE OR ALTER TRIGGER DBO.TRG_RESERVATIONS_TARGET_VALIDATE", "TRG_RESERVATION_TARGET_AUDIT_APPEND_ONLY", "INSTEAD OF UPDATE,DELETE", "TARGET_CONSTRAINT_OK", "TARGET_AUDIT_APPEND_ONLY_OK", "GROUP_CAPACITY_SNAPSHOT", "GO"} {
		if !strings.Contains(migration, required) {
			t.Fatalf("migration 002 lacks %s", required)
		}
	}
	for _, forbidden := range []string{"UPDATE DBO.RESERVATIONS SET TARGET_PARTICIPANTS", "DROP TABLE", "DELETE FROM"} {
		if strings.Contains(migration, forbidden) {
			t.Fatalf("migration 002 is not safe for legacy/rerun: %s", forbidden)
		}
	}
}

func TestGroupCompletionMigrationAndRoutesContract(t *testing.T) {
	migration := strings.ToUpper(readRepositoryFile(t, "database", "migrations", "004_group_flow_completion.sql"))
	routes := strings.ToUpper(readRepositoryFile(t, "backend", "internal", "routes", "routes.go"))
	repository := strings.ToUpper(readRepositoryFile(t, "backend", "internal", "repositories", "participants_repository.go"))
	for _, required := range []string{"RESERVATION_JOIN_CODE_SECRETS", "KEY_VERSION", "NONCE VARBINARY", "CIPHERTEXT VARBINARY", "RESERVATION_GROUP_EXPIRATIONS", "IF OBJECT_ID", "JOIN_CODE_SECRETS_TABLE_OK", "JOIN_CODE_SECRETS_COLUMNS_OK", "GROUP_EXPIRATIONS_TABLE_OK", "GROUP_EXPIRATIONS_COLUMNS_OK", "SYS.COLUMNS", "IS_NULLABLE", "GROUP_EXPIRATIONS_UNIQUE_OK"} {
		if !strings.Contains(migration, required) {
			t.Fatalf("migration 004 lacks %s", required)
		}
	}
	for _, forbidden := range []string{"UPDATE DBO.RESERVATIONS SET", "DELETE FROM", "JOIN_CODE NVARCHAR"} {
		if strings.Contains(migration, forbidden) {
			t.Fatalf("migration 004 reinterprets historical data: %s", forbidden)
		}
	}
	for _, required := range []string{"/RESERVATIONS/:ID/JOIN-CODE", "/JOIN-CODE/ROTATE", "/GROUP-RESERVATIONS/:CODE/CONFIRMATION"} {
		if !strings.Contains(routes, required) {
			t.Fatalf("routes lack %s", required)
		}
	}
	for _, required := range []string{"WITH(UPDLOCK,HOLDLOCK)", "CONFIRMATION_DEADLINE", "RESERVATION_GROUP_EXPIRATIONS", "RESERVATION_JOIN_CODE_SECRETS"} {
		if !strings.Contains(repository, required) {
			t.Fatalf("repository lacks %s", required)
		}
	}
}

func TestGroupCompletionMigrationUsesRealSQLServerDateTime2Metadata(t *testing.T) {
	migration := strings.ToUpper(readRepositoryFile(t, "database", "migrations", "004_group_flow_completion.sql"))
	for _, required := range []string{
		"('ROTATED_AT','DATETIME2',6,19,0,0)",
		"('EXPIRED_AT','DATETIME2',6,19,0,0)",
		"TYPE_NAME(C.USER_TYPE_ID)='DATETIME2'",
		"C.MAX_LENGTH=6",
		"C.PRECISION=19",
		"C.SCALE=0",
	} {
		if !strings.Contains(migration, required) {
			t.Fatalf("migration 004 lacks real DATETIME2(0) metadata: %s", required)
		}
	}
	for _, impossible := range []string{
		"'DATETIME2',8,27,0,0",
		"TYPE_NAME(C.USER_TYPE_ID)='DATETIME2' AND C.MAX_LENGTH=8",
		"TYPE_NAME(C.USER_TYPE_ID)='DATETIME2' AND C.PRECISION=27",
	} {
		if strings.Contains(migration, impossible) {
			t.Fatalf("migration 004 regressed to invalid DATETIME2 metadata: %s", impossible)
		}
	}
}
