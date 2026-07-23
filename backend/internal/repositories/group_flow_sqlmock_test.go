package repositories

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/joinsecret"

	"github.com/DATA-DOG/go-sqlmock"
)

func withMockDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	previous := database.DB
	database.DB = db
	return mock, func() {
		database.DB = previous
		db.Close()
	}
}

func TestExpirePendingGroupTxIsAtomicAndIdempotent(t *testing.T) {
	mock, done := withMockDB(t)
	defer done()
	mock.ExpectBegin()
	tx, _ := database.DB.Begin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM dbo.participants WITH(UPDLOCK,HOLDLOCK)")).
		WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(9))
	mock.ExpectExec("UPDATE dbo.reservations SET status='CANCELLED'").
		WithArgs(7).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO dbo.reservation_group_expirations").
		WithArgs(7, 9, 10).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO dbo.notifications").
		WithArgs(3, 7).WillReturnResult(sqlmock.NewResult(1, 1))
	changed, err := expirePendingGroupTx(t.Context(), tx, 7, 3, 10)
	if err != nil || !changed {
		t.Fatalf("first expiration = %v, %v", changed, err)
	}
	mock.ExpectCommit()
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	tx, _ = database.DB.Begin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM dbo.participants WITH(UPDLOCK,HOLDLOCK)")).
		WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(9))
	mock.ExpectExec("UPDATE dbo.reservations SET status='CANCELLED'").
		WithArgs(7).WillReturnResult(sqlmock.NewResult(0, 0))
	changed, err = expirePendingGroupTx(t.Context(), tx, 7, 3, 10)
	if err != nil || changed {
		t.Fatalf("second expiration = %v, %v", changed, err)
	}
	mock.ExpectRollback()
	_ = tx.Rollback()
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRotateLegacyJoinCodeReplacesHashAndCreatesSecretAtomically(t *testing.T) {
	mock, done := withMockDB(t)
	defer done()
	key := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("r", 32)))
	if err := joinsecret.Configure("1:"+key, 1); err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT user_id FROM dbo.reservations WITH").
		WithArgs(9).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(4))
	mock.ExpectExec("UPDATE dbo.reservations SET join_code_hash=").
		WithArgs(9, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("MERGE dbo.reservation_join_code_secrets").
		WithArgs(9, 1, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	code, err := RotateOwnerJoinCode(9, 4)
	if err != nil || code == "" {
		t.Fatalf("rotation = %q, %v", code, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreationSecretInsertIsEncryptedAndTransactionControlled(t *testing.T) {
	mock, done := withMockDB(t)
	defer done()
	key := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("c", 32)))
	if err := joinsecret.Configure("1:"+key, 1); err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	tx, _ := database.DB.Begin()
	mock.ExpectExec("INSERT INTO dbo.reservation_join_code_secrets").
		WithArgs(12, 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if err := insertJoinCodeSecretTx(t.Context(), tx, 12, "plaintext-must-not-persist"); err != nil {
		t.Fatal(err)
	}
	mock.ExpectRollback()
	_ = tx.Rollback()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRotateJoinCodeReturnsUniformNotFoundForNonOwner(t *testing.T) {
	mock, done := withMockDB(t)
	defer done()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT user_id FROM dbo.reservations WITH").
		WithArgs(9).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(4))
	mock.ExpectRollback()
	_, err := RotateOwnerJoinCode(9, 99)
	if !errors.Is(err, ErrInvalidJoinCode) {
		t.Fatalf("error = %v", err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUserReservationOverlapGuardBlocksConfirmingJoinCode(t *testing.T) {
	mock, done := withMockDB(t)
	defer done()
	mock.ExpectBegin()
	tx, err := database.DB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT TOP (1) 1 FROM dbo.reservations existing WITH(UPDLOCK,HOLDLOCK) WHERE existing.user_id=@p1 AND existing.id<>@p2 AND existing.status IN('PENDING','CONFIRMED') AND existing.start_time < DATEADD(MINUTE, @p3, @p4) AND DATEADD(MINUTE, existing.duration_minutes, existing.start_time) > @p5")).
		WithArgs(4, 7, 60, start, start).
		WillReturnRows(sqlmock.NewRows([]string{"overlap"}).AddRow(1))
	overlaps, err := userHasActiveOverlapTx(context.Background(), tx, 4, 7, start, 60)
	if err != nil {
		t.Fatal(err)
	}
	if !overlaps {
		t.Fatal("expected overlap guard to detect a conflicting existing reservation")
	}
	mock.ExpectRollback()
	_ = tx.Rollback()
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

var _ *sql.DB
