package repositories

import (
	"errors"
	"time"
)

var (
	ErrOwnerCannotWithdraw    = errors.New("el solicitante no puede retirarse de su propia reserva")
	ErrGroupCapacity          = errors.New("la reserva alcanzo su capacidad maxima")
	ErrInvalidGroupConfig     = errors.New("configuracion grupal invalida")
	ErrReservationNotJoinable = errors.New("la reserva no admite cambios de participantes")
)

const (
	GroupConditionPending  = "PENDING_MINIMUM"
	GroupConditionHealthy  = "HEALTHY"
	GroupConditionAtRisk   = "AT_RISK"
	GroupConditionInactive = "INACTIVE"
)

// participantConfirmationWindowOpen aplica un único límite temporal para
// altas en reservas grupales, independientemente de si la reserva aún está
// PENDING o ya alcanzó CONFIRMED. En el instante exacto del deadline la
// ventana ya está cerrada.
func participantConfirmationWindowOpen(
	now time.Time,
	startTime time.Time,
	confirmationDeadlineMinutes int,
) bool {
	deadline := startTime.Add(
		-time.Duration(confirmationDeadlineMinutes) * time.Minute,
	)

	return now.Before(deadline)
}

// participantTransition calcula el cambio de participación sin tocar la BD.
//
// Regla principal MVP2:
//
//   PENDING + alcanza mínimo -> CONFIRMED
//
// Una reserva que ya fue CONFIRMED NO vuelve automáticamente a PENDING
// si posteriormente baja del mínimo.
//
// En ese caso:
//   reservation.status = CONFIRMED
//   group condition     = AT_RISK
//
// Las infracciones por retiro tardío o grupo bajo mínimo se implementarán
// posteriormente como eventos separados.
func participantTransition(
	exists bool,
	oldParticipantStatus string,
	owner bool,
	confirmedCount int,
	capacity int,
	minimum int,
	reservationStatus string,
	confirm bool,
) (
	mutate bool,
	newParticipantStatus string,
	newReservationStatus string,
	err error,
) {
	if capacity <= 0 || minimum <= 0 || minimum > capacity {
		return false, oldParticipantStatus, reservationStatus, ErrInvalidGroupConfig
	}

	if reservationStatus != "PENDING" && reservationStatus != "CONFIRMED" {
		return false, oldParticipantStatus, reservationStatus, ErrReservationNotJoinable
	}

	targetStatus := "CANCELLED"
	if confirm {
		targetStatus = "CONFIRMED"
	}

	// El owner nunca puede abandonar su propia reserva mediante este flujo.
	if !confirm && owner {
		return false, oldParticipantStatus, reservationStatus, ErrOwnerCannotWithdraw
	}

	// Operación idempotente.
	if exists && oldParticipantStatus == targetStatus {
		return false, targetStatus, reservationStatus, nil
	}

	// Retirar a alguien que ni siquiera participa es un no-op.
	if !exists && !confirm {
		return false, targetStatus, reservationStatus, nil
	}

	// Solo cuenta como nuevo confirmado si antes no estaba confirmado.
	if confirm &&
		oldParticipantStatus != "CONFIRMED" &&
		confirmedCount >= capacity {
		return false, oldParticipantStatus, reservationStatus, ErrGroupCapacity
	}

	newCount := confirmedCount

	if confirm {
		if !exists || oldParticipantStatus != "CONFIRMED" {
			newCount++
		}
	} else if exists && oldParticipantStatus == "CONFIRMED" {
		newCount--
	}

	if newCount < 0 {
		newCount = 0
	}

	newReservationStatus = reservationStatus

	switch reservationStatus {

	case "PENDING":
		// Una reserva pendiente se confirma al alcanzar por primera vez
		// el mínimo requerido.
		if newCount >= minimum {
			newReservationStatus = "CONFIRMED"
		}

	case "CONFIRMED":
		// OPCIÓN B:
		//
		// una reserva ya confirmada mantiene CONFIRMED aunque luego
		// quede temporalmente bajo el mínimo.
		newReservationStatus = "CONFIRMED"
	}

	return true, targetStatus, newReservationStatus, nil
}

// participantGroupCondition representa la condición actual del grupo sin
// modificar reservation.status.
//
// Ejemplos:
//
// PENDING + 7/10    -> PENDING_MINIMUM
// CONFIRMED + 10/10 -> HEALTHY
// CONFIRMED + 9/10  -> AT_RISK
func participantGroupCondition(
	reservationStatus string,
	confirmedCount int,
	minimum int,
) string {
	switch reservationStatus {

	case "PENDING":
		return GroupConditionPending

	case "CONFIRMED":
		if confirmedCount < minimum {
			return GroupConditionAtRisk
		}

		return GroupConditionHealthy

	default:
		return GroupConditionInactive
	}
}
