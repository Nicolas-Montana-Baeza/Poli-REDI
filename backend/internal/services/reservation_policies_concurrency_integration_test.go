//go:build integration

package services

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func TestConcurrentPolicyPublishPostgresIntegration(t *testing.T) {
	if os.Getenv("POLIREDI_3C_ISOLATED") != "1" {
		t.Skip("requiere PostgreSQL efimero 3C")
	}

	if os.Getenv("DATABASE_URL") == "" {
		t.Fatal("DATABASE_URL obligatorio")
	}

	if err := database.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(database.Close)

	ctx := context.Background()

	var adminID int

	err := database.DB.QueryRowContext(
		ctx,
		`SELECT id FROM users
		 WHERE lower(email) = 'admin@poliredi.local'`,
	).Scan(&adminID)

	if err != nil {
		t.Fatal(err)
	}

	current, err := GetCurrentReservationPolicy()
	if err != nil {
		t.Fatal(err)
	}

	countPolicies := func() int {
		t.Helper()

		var count int

		err := database.DB.QueryRowContext(
			ctx,
			`SELECT COUNT(*) FROM reservation_policies`,
		).Scan(&count)

		if err != nil {
			t.Fatal(err)
		}

		return count
	}

	before := countPolicies()

	makeRequest := func(increment int) models.PublishReservationPolicyRequest {
		return models.PublishReservationPolicyRequest{
			ExpectedPolicyID:            current.ID,
			ReservableWindowDays:        current.ReservableWindowDays + increment,
			RequestFrequencyDays:        current.RequestFrequencyDays,
			ConfirmationDeadlineMinutes: current.ConfirmationDeadlineMinutes,
			MinimumParticipants:         current.MinimumParticipants,
			OpeningMinute:               current.OpeningMinute,
			ClosingMinute:               current.ClosingMinute,
			SlotIntervalMinutes:         current.SlotIntervalMinutes,
			AllowedDurations:            append([]int{}, current.AllowedDurations...),
			ResourceIDs:                 append([]int{}, current.ResourceIDs...),
		}
	}

	requests := []models.PublishReservationPolicyRequest{
		makeRequest(1),
		makeRequest(2),
	}

	keys := []string{
		"mvp3-3c-concurrency-first",
		"mvp3-3c-concurrency-second",
	}

	results := make([]error, len(requests))

	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := range requests {
		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			<-start

			_, _, results[index] = PublishReservationPolicy(
				requests[index],
				adminID,
				keys[index],
			)
		}(i)
	}

	close(start)
	wg.Wait()

	successes := 0
	conflicts := 0

	for _, err := range results {
		switch {
		case err == nil:
			successes++

		case errors.Is(err, ErrReservationPolicyStale),
			errors.Is(err, ErrReservationPolicyConcurrent):
			conflicts++

		default:
			t.Fatalf("error inesperado: %v", err)
		}
	}

	if successes != 1 || conflicts != 1 {
		t.Fatalf(
			"esperaba 1 publicacion y 1 conflicto: exitos=%d conflictos=%d",
			successes,
			conflicts,
		)
	}

	if countPolicies() != before+1 {
		t.Fatal("se creo una cantidad incorrecta de versiones")
	}

	latest, err := GetCurrentReservationPolicy()
	if err != nil {
		t.Fatal(err)
	}

	if latest.ID == current.ID {
		t.Fatal("la nueva politica no quedo vigente")
	}

	t.Log("Una version publicada; la segunda fue rechazada.")
}
