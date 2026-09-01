package repositories

import (
	"strings"
	"testing"
)

func TestRuntimeRepositoriesContainNoExecutableSQLServerSyntax(
	t *testing.T,
) {
	files := []string{
		readRepositoryFile(
			t,
			"backend",
			"internal",
			"repositories",
			"notifications_repository.go",
		),
		readRepositoryFile(
			t,
			"backend",
			"internal",
			"repositories",
			"reservation_policies_repository.go",
		),
	}

	for _, raw := range files {
		sqlText := strings.ToUpper(raw)

		for _, forbidden := range []string{
			"DBO.",
			"UPDLOCK",
			"HOLDLOCK",
			"OUTPUT INSERTED",
			"SYSUTCDATETIME",
			"SELECT TOP (",
			"@P1",
			"@P2",
		} {
			if strings.Contains(
				sqlText,
				forbidden,
			) {
				t.Fatalf(
					"runtime repository contains SQL Server syntax %q",
					forbidden,
				)
			}
		}
	}
}

func TestPostgresPolicyPublicationPreservesRequiredContracts(
	t *testing.T,
) {
	raw := strings.ToUpper(
		readRepositoryFile(
			t,
			"backend",
			"internal",
			"repositories",
			"reservation_policies_repository.go",
		),
	)

	for _, required := range []string{
		"LOCK TABLE RESERVATION_POLICIES",
		"SHARE ROW EXCLUSIVE MODE",
		"CURRENT_TIMESTAMP",
		"IDEMPOTENCY_PAYLOAD_HASH",
		"TO_REGCLASS",
		"RESERVATION_POLICY_GROUP_RESOURCES",
		"LATE_WITHDRAWAL_MINUTES",
		"GROUP_RECOVERY_DEADLINE_MINUTES",
		"ON CONFLICT DO NOTHING",
		"RETURNING ID",
	} {
		if !strings.Contains(
			raw,
			required,
		) {
			t.Fatalf(
				"PostgreSQL policy publication lacks %q",
				required,
			)
		}
	}
}

func TestPostgresNotificationsMigrationAndRepository(
	t *testing.T,
) {
	migration := strings.ToUpper(
		readRepositoryFile(
			t,
			"database",
			"postgres",
			"migrations",
			"PG16_0009_full_notifications.sql",
		),
	)

	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS NOTIFICATIONS",
		"USER_ID INTEGER NOT NULL",
		"RESERVATION_ID INTEGER",
		"TIMESTAMPTZ",
		"ON DELETE CASCADE",
		"ON DELETE SET NULL",
		"IX_NOTIFICATIONS_USER_READ_CREATED",
	} {
		if !strings.Contains(
			migration,
			required,
		) {
			t.Fatalf(
				"notifications migration lacks %q",
				required,
			)
		}
	}

	repository := strings.ToUpper(
		readRepositoryFile(
			t,
			"backend",
			"internal",
			"repositories",
			"notifications_repository.go",
		),
	)

	for _, required := range []string{
		"FROM NOTIFICATIONS",
		"WHERE USER_ID = $1",
		"LIMIT 20",
	} {
		if !strings.Contains(
			repository,
			required,
		) {
			t.Fatalf(
				"notifications repository lacks %q",
				required,
			)
		}
	}
}

func TestPostgresGroupResourceRulesMigrationAndRepositories(
	t *testing.T,
) {
	migration := strings.ToUpper(
		readRepositoryFile(
			t,
			"database",
			"postgres",
			"migrations",
			"PG16_0010_mvp2_group_resource_rules.sql",
		),
	)

	for _, required := range []string{
		"RESERVATION_POLICY_GROUP_RESOURCES",
		"MINIMUM_PARTICIPANTS INTEGER",
		"GROUP_MINIMUM_PARTICIPANTS_SNAPSHOT",
		"SALA MULTIUSO, CENTRO DEPORTIVO",
		"MVP2-GROUP-RESOURCE-RULES-V1",
		"CONFIRMATION_DEADLINE_MINUTES",
		"RESERVATION_MODE = 'RESERVABLE'",
	} {
		if !strings.Contains(migration, required) {
			t.Fatalf(
				"group resource rules migration lacks %q",
				required,
			)
		}
	}

	reservationRepository := strings.ToUpper(
		readRepositoryFile(
			t,
			"backend",
			"internal",
			"repositories",
			"reservations_repository.go",
		),
	)

	for _, required := range []string{
		"GROUP_MINIMUM_PARTICIPANTS_SNAPSHOT",
		"SELECT MINIMUM_PARTICIPANTS",
		"FROM RESERVATION_POLICY_GROUP_RESOURCES",
	} {
		if !strings.Contains(reservationRepository, required) {
			t.Fatalf(
				"reservation repository lacks %q",
				required,
			)
		}
	}

	participantRepository := strings.ToUpper(
		readRepositoryFile(
			t,
			"backend",
			"internal",
			"repositories",
			"participants_repository.go",
		),
	)

	if !strings.Contains(
		participantRepository,
		"R.GROUP_MINIMUM_PARTICIPANTS_SNAPSHOT",
	) {
		t.Fatal(
			"participant progress does not use the minimum snapshot",
		)
	}

	for _, required := range []string{
		"PG_ADVISORY_XACT_LOCK(73002, $1)",
		"OTHER.START_TIME < $4",
		"OTHER.END_TIME > $3",
		"OTHER.USER_ID = $1",
		"MEMBERSHIP.STATUS = 'CONFIRMED'",
	} {
		if !strings.Contains(participantRepository, required) {
			t.Fatalf(
				"participant overlap protection lacks %q",
				required,
			)
		}
	}

	if strings.Contains(
		participantRepository,
		"P.GROUP_RECOVERY_DEADLINE_MINUTES",
	) {
		t.Fatal(
			"participant join still depends on recovery deadline",
		)
	}

	policyRepository := strings.ToUpper(
		readRepositoryFile(
			t,
			"backend",
			"internal",
			"repositories",
			"reservation_policies_repository.go",
		),
	)

	if !strings.Contains(
		policyRepository,
		"OLD_GROUP.MINIMUM_PARTICIPANTS",
	) {
		t.Fatal(
			"policy publication does not preserve per-resource minimum",
		)
	}

	installer := strings.ToUpper(
		readRepositoryFile(
			t,
			"infra",
			"local",
			"quadlet",
			"install.sh",
		),
	)

	if !strings.Contains(
		installer,
		"PG16_0010_MVP2_GROUP_RESOURCE_RULES.SQL",
	) {
		t.Fatal("Quadlet bootstrap does not include PG16_0010")
	}
}

func TestPostgresPendingGroupExpirationContracts(t *testing.T) {
	repository := strings.ToUpper(
		readRepositoryFile(
			t,
			"backend",
			"internal",
			"repositories",
			"reservations_repository.go",
		),
	)

	for _, required := range []string{
		"STATUS = 'CANCELLED'",
		"CANCELLATION_REASON = $2",
		"RESERVATION.STATUS = 'PENDING'",
		"RESERVATION.GROUP_CAPACITY_SNAPSHOT IS NOT NULL",
		"POLICY.CONFIRMATION_DEADLINE_MINUTES",
		"PARTICIPANT.STATUS = 'CONFIRMED'",
		"RESERVATION.GROUP_MINIMUM_PARTICIPANTS_SNAPSHOT",
		"CANCELLATIONREASONMINIMUMNOTMET",
	} {
		if !strings.Contains(repository, required) {
			t.Fatalf(
				"pending group expiration lacks %q",
				required,
			)
		}
	}

	if strings.Contains(
		repository,
		"GROUP_RECOVERY_DEADLINE_MINUTES",
	) {
		t.Fatal(
			"pending group expiration depends on recovery deadline",
		)
	}

	services := strings.ToUpper(
		readRepositoryFile(
			t,
			"backend",
			"internal",
			"services",
			"reservations_service.go",
		),
	)

	for _, required := range []string{
		"RUNRESERVATIONHOUSEKEEPING()",
		"EXPIREPENDINGGROUPRESERVATIONS(NOW)",
	} {
		if !strings.Contains(services, required) {
			t.Fatalf(
				"reservation housekeeping lacks %q",
				required,
			)
		}
	}
}

func TestMVP2EphemeralVerificationContracts(t *testing.T) {
	script := strings.ToUpper(
		readRepositoryFile(
			t,
			"infra",
			"local",
			"quadlet",
			"verify-mvp2-ephemeral.sh",
		),
	)

	for _, required := range []string{
		"POSTGRES:16-ALPINE",
		"PG16_0001_MVP1_BASELINE.SQL",
		"PG16_0002_MVP1_INDEXES.SQL",
		"PG16_0003_MVP1_INVARIANTS.SQL",
		"PG16_0004_MVP2_GROUP_PARTICIPANTS.SQL",
		"PG16_0005_MVP2_INSTITUTIONAL_SCHEDULING.SQL",
		"PG16_0006_MVP2_INSTITUTIONAL_AVAILABILITY.SQL",
		"PG16_0007_MVP2_SCHEDULE_EXCEPTIONS.SQL",
		"PG16_0008_MVP2_SCHEDULE_EXCEPTION_AVAILABILITY.SQL",
		"PG16_0009_FULL_NOTIFICATIONS.SQL",
		"PG16_0010_MVP2_GROUP_RESOURCE_RULES.SQL",
		"POLIREDI_INTEGRATION='1'",
		"GO TEST ./... -P 1 -COUNT=1",
		"VITE_MVP_SCOPE='MVP2'",
		"NPM RUN BUILD",
		"PODMAN VOLUME RM",
		"CHMOD 0755 \"${RUNTIME_DIR}\"",
		"WAIT_UNTIL_INITIALIZED",
		"POSTGRESQL INIT PROCESS COMPLETE; READY FOR START UP.",
		"FRONTEND_RUNTIME_DIR",
		"--EXCLUDE='./NODE_MODULES'",
		"NPM CI",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf(
				"ephemeral MVP2 verification lacks %q",
				required,
			)
		}
	}

	verification := strings.ToUpper(
		readRepositoryFile(
			t,
			"database",
			"postgres",
			"check",
			"PG16_verify_mvp2.sql",
		),
	)

	for _, required := range []string{
		"MVP2-GROUP-RESOURCE-RULES-V1",
		"CONFIRMATION_DEADLINE_MINUTES = 60",
		"GROUP_RECOVERY_DEADLINE_MINUTES = 0",
		"SALA MULTIUSO, CENTRO DEPORTIVO",
		"MINIMUM_PARTICIPANTS = 10",
		"UQ_RESERVATIONS_JOIN_CODE_HASH",
		"UQ_PARTICIPANTS_RESERVATION_USER",
		"ROLLBACK",
	} {
		if !strings.Contains(verification, required) {
			t.Fatalf(
				"MVP2 SQL verification lacks %q",
				required,
			)
		}
	}
}
