package schedulingconflicts

import (
	"errors"
	"testing"
	"time"
)

func TestDetectConnectedComponentsPairwiseN(t *testing.T) {
	resourceID := 1

	activityA := 10
	activityB := 20
	activityC := 30

	scheduleA := 101
	scheduleB := 102
	scheduleC := 103

	occupancies := []Occupancy{
		{
			Key:                     "A",
			ResourceID:              resourceID,
			Kind:                    OccupancyKindInstitutionalActivity,
			InstitutionalActivityID: &activityA,
			ScheduleID:              &scheduleA,
			Start:                   testTime(t, "2026-09-01 10:00"),
			End:                     testTime(t, "2026-09-01 11:30"),
		},
		{
			Key:                     "B",
			ResourceID:              resourceID,
			Kind:                    OccupancyKindInstitutionalActivity,
			InstitutionalActivityID: &activityB,
			ScheduleID:              &scheduleB,
			Start:                   testTime(t, "2026-09-01 10:30"),
			End:                     testTime(t, "2026-09-01 12:00"),
		},
		{
			Key:                     "C",
			ResourceID:              resourceID,
			Kind:                    OccupancyKindInstitutionalActivity,
			InstitutionalActivityID: &activityC,
			ScheduleID:              &scheduleC,
			Start:                   testTime(t, "2026-09-01 11:00"),
			End:                     testTime(t, "2026-09-01 13:00"),
		},
	}

	components, err :=
		DetectConnectedComponents(occupancies)

	if err != nil {
		t.Fatalf("detect components: %v", err)
	}

	if len(components) != 1 {
		t.Fatalf(
			"expected 1 component, got %d",
			len(components),
		)
	}

	if len(components[0].Items) != 3 {
		t.Fatalf(
			"expected 3 items, got %d",
			len(components[0].Items),
		)
	}

	start, end, err :=
		ProtectedInterval(components[0])

	if err != nil {
		t.Fatalf("protected interval: %v", err)
	}

	if !start.Equal(
		testTime(t, "2026-09-01 10:00"),
	) {
		t.Fatalf(
			"expected protected start 10:00, got %v",
			start,
		)
	}

	if !end.Equal(
		testTime(t, "2026-09-01 13:00"),
	) {
		t.Fatalf(
			"expected protected end 13:00, got %v",
			end,
		)
	}
}

// TestDetectConnectedComponentsChainedConflict demuestra el caso importante:
//
//	A 10:00-11:00
//	B 10:30-11:30
//	C 11:15-12:00
//
// A y C NO se intersectan directamente.
//
// Sin embargo:
//
//	A ↔ B ↔ C
//
// constituye un único componente conectado.
func TestDetectConnectedComponentsChainedConflict(
	t *testing.T,
) {
	resourceID := 1

	activityA := 10
	activityB := 20
	activityC := 30

	scheduleA := 101
	scheduleB := 102
	scheduleC := 103

	occupancies := []Occupancy{
		newActivityOccupancy(
			"A",
			resourceID,
			activityA,
			scheduleA,
			testTime(t, "2026-09-01 10:00"),
			testTime(t, "2026-09-01 11:00"),
		),

		newActivityOccupancy(
			"B",
			resourceID,
			activityB,
			scheduleB,
			testTime(t, "2026-09-01 10:30"),
			testTime(t, "2026-09-01 11:30"),
		),

		newActivityOccupancy(
			"C",
			resourceID,
			activityC,
			scheduleC,
			testTime(t, "2026-09-01 11:15"),
			testTime(t, "2026-09-01 12:00"),
		),
	}

	components, err :=
		DetectConnectedComponents(occupancies)

	if err != nil {
		t.Fatalf("detect components: %v", err)
	}

	if len(components) != 1 {
		t.Fatalf(
			"expected 1 chained component, got %d",
			len(components),
		)
	}

	if len(components[0].Items) != 3 {
		t.Fatalf(
			"expected 3 chained items, got %d",
			len(components[0].Items),
		)
	}

	start, end, err :=
		ProtectedInterval(components[0])

	if err != nil {
		t.Fatalf("protected interval: %v", err)
	}

	if !start.Equal(
		testTime(t, "2026-09-01 10:00"),
	) ||
		!end.Equal(
			testTime(t, "2026-09-01 12:00"),
		) {
		t.Fatalf(
			"expected protected interval 10:00-12:00, got %v-%v",
			start,
			end,
		)
	}
}

// Los intervalos consecutivos no deben producir conflictos:
//
//	A 10:00-11:00
//	B 11:00-12:00
//
// Usamos semántica [start,end).
func TestDetectConnectedComponentsTouchingIntervals(
	t *testing.T,
) {
	resourceID := 1

	activityA := 10
	activityB := 20

	scheduleA := 101
	scheduleB := 102

	occupancies := []Occupancy{
		newActivityOccupancy(
			"A",
			resourceID,
			activityA,
			scheduleA,
			testTime(t, "2026-09-01 10:00"),
			testTime(t, "2026-09-01 11:00"),
		),

		newActivityOccupancy(
			"B",
			resourceID,
			activityB,
			scheduleB,
			testTime(t, "2026-09-01 11:00"),
			testTime(t, "2026-09-01 12:00"),
		),
	}

	components, err :=
		DetectConnectedComponents(occupancies)

	if err != nil {
		t.Fatalf("detect components: %v", err)
	}

	if len(components) != 0 {
		t.Fatalf(
			"expected no conflict, got %d components",
			len(components),
		)
	}
}

// Distintos conflictos pueden coexistir y los recursos nunca deben mezclarse
// dentro del mismo componente.
func TestDetectConnectedComponentsMultipleGroups(
	t *testing.T,
) {
	activity1 := 1
	activity2 := 2
	activity3 := 3
	activity4 := 4

	schedule1 := 101
	schedule2 := 102
	schedule3 := 103
	schedule4 := 104

	reservation1 := 500
	reservation2 := 501

	occupancies := []Occupancy{
		newActivityOccupancy(
			"A",
			1,
			activity1,
			schedule1,
			testTime(t, "2026-09-01 10:00"),
			testTime(t, "2026-09-01 11:00"),
		),

		newActivityOccupancy(
			"B",
			1,
			activity2,
			schedule2,
			testTime(t, "2026-09-01 10:30"),
			testTime(t, "2026-09-01 11:30"),
		),

		newActivityOccupancy(
			"C",
			1,
			activity3,
			schedule3,
			testTime(t, "2026-09-01 15:00"),
			testTime(t, "2026-09-01 16:00"),
		),

		newActivityOccupancy(
			"D",
			1,
			activity4,
			schedule4,
			testTime(t, "2026-09-01 15:30"),
			testTime(t, "2026-09-01 17:00"),
		),

		newReservationOccupancy(
			"R1",
			2,
			reservation1,
			testTime(t, "2026-09-01 09:00"),
			testTime(t, "2026-09-01 10:00"),
		),

		newReservationOccupancy(
			"R2",
			2,
			reservation2,
			testTime(t, "2026-09-01 09:30"),
			testTime(t, "2026-09-01 10:30"),
		),
	}

	components, err :=
		DetectConnectedComponents(occupancies)

	if err != nil {
		t.Fatalf("detect components: %v", err)
	}

	if len(components) != 3 {
		t.Fatalf(
			"expected 3 conflict groups, got %d",
			len(components),
		)
	}

	if components[0].ResourceID != 1 ||
		len(components[0].Items) != 2 {
		t.Fatalf("unexpected first component")
	}

	if components[1].ResourceID != 1 ||
		len(components[1].Items) != 2 {
		t.Fatalf("unexpected second component")
	}

	if components[2].ResourceID != 2 ||
		len(components[2].Items) != 2 {
		t.Fatalf("unexpected third component")
	}
}

func TestDetectConnectedComponentsRejectsDuplicateKey(
	t *testing.T,
) {
	activityID := 1
	scheduleID := 1

	occupancy := newActivityOccupancy(
		"duplicate",
		1,
		activityID,
		scheduleID,
		testTime(t, "2026-09-01 10:00"),
		testTime(t, "2026-09-01 11:00"),
	)

	_, err := DetectConnectedComponents(
		[]Occupancy{
			occupancy,
			occupancy,
		},
	)

	if !errors.Is(
		err,
		ErrDuplicateOccupancy,
	) {
		t.Fatalf(
			"expected duplicate occupancy error, got %v",
			err,
		)
	}
}

// ============================================================================
// HELPERS
// ============================================================================

func testTime(
	t *testing.T,
	value string,
) time.Time {
	t.Helper()

	location, err :=
		time.LoadLocation("America/Santiago")

	if err != nil {
		t.Fatalf("load timezone: %v", err)
	}

	parsed, err := time.ParseInLocation(
		"2006-01-02 15:04",
		value,
		location,
	)

	if err != nil {
		t.Fatalf(
			"parse test time %q: %v",
			value,
			err,
		)
	}

	return parsed
}

func newActivityOccupancy(
	key string,
	resourceID int,
	activityID int,
	scheduleID int,
	start time.Time,
	end time.Time,
) Occupancy {
	return Occupancy{
		Key:                     key,
		ResourceID:              resourceID,
		Kind:                    OccupancyKindInstitutionalActivity,
		InstitutionalActivityID: &activityID,
		ScheduleID:              &scheduleID,
		Start:                   start,
		End:                     end,
	}
}

func newReservationOccupancy(
	key string,
	resourceID int,
	reservationID int,
	start time.Time,
	end time.Time,
) Occupancy {
	return Occupancy{
		Key:           key,
		ResourceID:    resourceID,
		Kind:          OccupancyKindReservation,
		ReservationID: &reservationID,
		Start:         start,
		End:           end,
	}
}
