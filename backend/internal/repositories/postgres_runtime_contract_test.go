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
