package services

import (
	"errors"
	"sort"
	"strings"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

func GetCurrentReservationPolicy() (models.ReservationPolicy, error) {
	return repositories.GetCurrentReservationPolicyComplete()
}

func GetReservationPolicyHistory() ([]models.ReservationPolicy, error) {
	return repositories.GetReservationPolicyHistory()
}

func PublishReservationPolicy(request models.PublishReservationPolicyRequest, createdBy int, key string) (models.ReservationPolicy, bool, error) {
	key = strings.TrimSpace(key)
	if key == "" || len(key) > 100 {
		return models.ReservationPolicy{}, false, errors.New("Idempotency-Key es obligatorio y debe tener hasta 100 caracteres")
	}
	if createdBy <= 0 {
		return models.ReservationPolicy{}, false, errors.New("administrador autenticado es obligatorio")
	}
	if request.ReservableWindowDays <= 0 || request.RequestFrequencyDays <= 0 || request.ConfirmationDeadlineMinutes < 0 || request.MinimumParticipants <= 0 {
		return models.ReservationPolicy{}, false, errors.New("los parametros de la politica no son validos")
	}
	if request.OpeningMinute < 0 || request.ClosingMinute > 1439 || request.OpeningMinute >= request.ClosingMinute || request.SlotIntervalMinutes <= 0 || request.SlotIntervalMinutes > 1440 {
		return models.ReservationPolicy{}, false, errors.New("la jornada configurada no es valida")
	}
	if len(request.AllowedDurations) == 0 {
		return models.ReservationPolicy{}, false, errors.New("debe existir al menos una duracion permitida")
	}
	for _, value := range request.AllowedDurations {
		if value <= 0 {
			return models.ReservationPolicy{}, false, errors.New("las duraciones permitidas deben ser mayores que cero")
		}
	}
	for _, value := range request.ResourceIDs {
		if value <= 0 {
			return models.ReservationPolicy{}, false, errors.New("los identificadores de recursos deben ser mayores que cero")
		}
	}
	request.AllowedDurations = uniquePositive(request.AllowedDurations)
	request.ResourceIDs = uniquePositive(request.ResourceIDs)
	return repositories.PublishReservationPolicy(request, createdBy, key)
}

func uniquePositive(values []int) []int {
	seen := map[int]bool{}
	result := make([]int, 0, len(values))
	for _, value := range values {
		if value > 0 && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	sort.Ints(result)
	return result
}
