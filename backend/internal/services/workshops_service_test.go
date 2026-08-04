package services

import (
	"errors"
	"testing"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"

	mssql "github.com/microsoft/go-mssqldb"
)

func TestMapWorkshopErrorNeverLeaksSQLDetails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		number int32
		want   error
	}{
		{name: "unknown SQL", number: 50099, want: repositories.ErrWorkshopInternal},
		{name: "deadlock", number: 1205, want: repositories.ErrWorkshopInternal},
		{name: "invalid schedule trigger", number: 51300, want: repositories.ErrWorkshopScheduleInvalid},
		{name: "capacity trigger", number: 51301, want: repositories.ErrWorkshopCapacity},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := mapWorkshopError(mssql.Error{
				Number:  test.number,
				Message: "SECRET SQL TABLE/CONSTRAINT DETAIL",
			})
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v; want %v", err, test.want)
			}
			if err.Error() == "SECRET SQL TABLE/CONSTRAINT DETAIL" {
				t.Fatal("SQL detail leaked")
			}
		})
	}
}

func TestWithdrawFromWorkshopValidatesIdentityAndPreservesClosedError(t *testing.T) {
	t.Parallel()

	if _, err := WithdrawFromWorkshop(0, models.LocalAuthUser{ID: 1}); err == nil {
		t.Fatal("expected invalid workshop id")
	}
	if _, err := WithdrawFromWorkshop(1, models.LocalAuthUser{}); err == nil {
		t.Fatal("expected invalid user")
	}
	if err := mapWorkshopError(repositories.ErrWorkshopEnrollmentClosed); !errors.Is(err, repositories.ErrWorkshopEnrollmentClosed) {
		t.Fatalf("closed error was not preserved: %v", err)
	}
}

func TestMapWorkshopErrorHidesUnknownInternalErrors(t *testing.T) {
	t.Parallel()
	err := mapWorkshopError(errors.New("driver secret"))
	if !errors.Is(err, repositories.ErrWorkshopInternal) {
		t.Fatalf("got %v", err)
	}
}
