package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

var ErrReservationPolicyConflict = errors.New("la clave de idempotencia ya fue utilizada con otros datos")

type ReservationPolicyValidationError struct{ message string }

func (e ReservationPolicyValidationError) Error() string { return e.message }

func policyValidation(message string) error {
	return ReservationPolicyValidationError{message: message}
}

func GetCurrentReservationPolicy() (models.ReservationPolicy, error) {
	return repositories.GetCurrentReservationPolicyComplete()
}

func GetReservationPolicyHistory() ([]models.ReservationPolicy, error) {
	return repositories.GetReservationPolicyHistory()
}

func PublishReservationPolicy(request models.PublishReservationPolicyRequest, createdBy int, key string) (models.ReservationPolicy, bool, error) {
	key = strings.TrimSpace(key)
	if key == "" || len(key) > 100 {
		return models.ReservationPolicy{}, false, policyValidation("Idempotency-Key es obligatorio y debe tener hasta 100 caracteres")
	}
	if createdBy <= 0 {
		return models.ReservationPolicy{}, false, policyValidation("administrador autenticado es obligatorio")
	}
	if request.ReservableWindowDays <= 0 || request.RequestFrequencyDays <= 0 || request.ConfirmationDeadlineMinutes < 0 || request.MinimumParticipants <= 0 {
		return models.ReservationPolicy{}, false, policyValidation("los parametros de la politica no son validos")
	}
	if request.OpeningMinute < 0 || request.ClosingMinute > 1439 || request.OpeningMinute >= request.ClosingMinute || request.SlotIntervalMinutes <= 0 || request.SlotIntervalMinutes > 1440 {
		return models.ReservationPolicy{}, false, policyValidation("la jornada configurada no es valida")
	}
	if len(request.AllowedDurations) == 0 {
		return models.ReservationPolicy{}, false, policyValidation("debe existir al menos una duracion permitida")
	}
	for _, value := range request.AllowedDurations {
		if value <= 0 {
			return models.ReservationPolicy{}, false, policyValidation("las duraciones permitidas deben ser mayores que cero")
		}
	}
	for _, value := range request.ResourceIDs {
		if value <= 0 {
			return models.ReservationPolicy{}, false, policyValidation("los identificadores de recursos deben ser mayores que cero")
		}
	}
	request.AllowedDurations = uniquePositive(request.AllowedDurations)
	request.ResourceIDs = uniquePositive(request.ResourceIDs)
	hash, err := reservationPolicyPayloadHash(request)
	if err != nil {
		return models.ReservationPolicy{}, false, err
	}
	policy, replayed, err := repositories.PublishReservationPolicy(request, createdBy, key, hash)
	if errors.Is(err, repositories.ErrIdempotencyPayloadConflict) {
		return models.ReservationPolicy{}, false, ErrReservationPolicyConflict
	}
	if errors.Is(err, repositories.ErrInvalidPolicyResource) {
		return models.ReservationPolicy{}, false, policyValidation("uno o mas recursos no existen o no estan activos")
	}
	return policy, replayed, err
}

func reservationPolicyPayloadHash(request models.PublishReservationPolicyRequest) (string, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
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
