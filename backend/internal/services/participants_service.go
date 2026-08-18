package services

import (
	"database/sql"
	"errors"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

// Errores públicos del servicio de participantes.
//
// Se reutilizan los errores de dominio definidos en repositories para que
// handlers pueda distinguir correctamente errores de validación, permisos
// y conflictos sin depender directamente de la capa de persistencia.
var (
	ErrInvalidJoinCode           = repositories.ErrInvalidJoinCode
	ErrParticipantIneligible     = repositories.ErrParticipantIneligible
	ErrParticipationWindowClosed = repositories.ErrParticipationWindowClosed
	ErrOwnerCannotWithdraw       = repositories.ErrOwnerCannotWithdraw
	ErrGroupCapacity             = repositories.ErrGroupCapacity
	ErrInvalidGroupConfig        = repositories.ErrInvalidGroupConfig
	ErrReservationNotJoinable    = repositories.ErrReservationNotJoinable
)

// GetReservationProgressByJoinCode obtiene el estado actual de una reserva
// grupal utilizando su join code.
//
// El código recibido nunca se compara directamente con un valor almacenado
// en PostgreSQL. El repository normaliza el código y calcula su hash SHA-256
// antes de localizar la reserva.
//
// La respuesta incluye información derivada como:
//
//   - cantidad actual de participantes;
//   - mínimo requerido;
//   - capacidad máxima;
//   - condición grupal;
//   - si el usuario autenticado ya pertenece al grupo;
//   - si el usuario autenticado es el owner.
func GetReservationProgressByJoinCode(
	code string,
	userID int,
) (models.ReservationProgress, error) {

	if userID <= 0 {
		return models.ReservationProgress{},
			errors.New("usuario autenticado es obligatorio")
	}

	return repositories.GetReservationProgress(
		code,
		userID,
	)
}

// JoinGroupReservation incorpora al usuario autenticado como participante
// confirmado de una reserva grupal.
//
// Toda la lógica crítica se ejecuta dentro del repository mediante una
// transacción SERIALIZABLE:
//
//   - valida el join code;
//   - bloquea la reserva durante la modificación;
//   - verifica elegibilidad del usuario;
//   - respeta capacidad y ventanas temporales;
//   - actualiza el conteo de participantes;
//   - confirma la reserva cuando alcanza el mínimo;
//   - registra el cambio en la auditoría.
//
// El servicio no permite especificar otro userID desde el cliente; siempre
// debe utilizarse el usuario autenticado por Poli-REDI.
func JoinGroupReservation(
	code string,
	userID int,
) (models.ReservationProgress, error) {

	if userID <= 0 {
		return models.ReservationProgress{},
			errors.New("usuario autenticado es obligatorio")
	}

	return repositories.ChangeParticipation(
		code,
		userID,
		true,
	)
}

// LeaveGroupReservation retira al usuario autenticado de una reserva grupal.
//
// Una reserva que ya alcanzó CONFIRMED nunca vuelve automáticamente a
// PENDING cuando pierde participantes.
//
// Si el grupo cae bajo el mínimo:
//
//	reservation.status = CONFIRMED
//	groupCondition      = AT_RISK
//
// Esto permite al owner recuperar participantes antes del plazo definido
// por la política.
//
// El owner no puede abandonar el grupo mediante este flujo: debe cancelar
// la reserva completa.
func LeaveGroupReservation(
	code string,
	userID int,
) (models.ReservationProgress, error) {

	if userID <= 0 {
		return models.ReservationProgress{},
			errors.New("usuario autenticado es obligatorio")
	}

	return repositories.ChangeParticipation(
		code,
		userID,
		false,
	)
}

// GetReservationParticipantsForUser obtiene los participantes de una
// reserva grupal aplicando control de acceso.
//
// Por privacidad, la lista completa no queda disponible para cualquier
// participante que conozca el join code.
//
// Solo pueden consultarla:
//
//   - el owner de la reserva;
//   - un administrador.
//
// La autorización se basa en reservation.user_id, que representa al creador
// y owner institucional de la reserva.
func GetReservationParticipantsForUser(
	reservationID int,
	requestedBy models.LocalAuthUser,
) ([]models.ReservationParticipant, error) {

	if reservationID <= 0 {
		return nil, ErrReservationNotFound
	}

	reservation, err :=
		repositories.GetReservationByID(
			reservationID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrReservationNotFound
	}

	if err != nil {
		return nil, err
	}

	if !requestedBy.IsAdmin &&
		reservation.UserID != requestedBy.ID {

		return nil, ErrReservationForbidden
	}

	return repositories.GetReservationParticipants(
		reservationID,
	)
}

// RotateReservationJoinCodeForUser rota el código de invitación de una
// reserva grupal aplicando control de acceso.
//
// Solo pueden realizar esta operación:
//
//   - el owner de la reserva;
//   - un administrador.
//
// El servicio autoriza primero por reservation.user_id y recién después
// delega la rotación al repository. De esta forma, conocer el ID de una
// reserva no permite invalidar su código de invitación.
//
// El código en texto plano se devuelve únicamente como resultado de esta
// operación. PostgreSQL conserva solamente su hash.
func RotateReservationJoinCodeForUser(
	reservationID int,
	requestedBy models.LocalAuthUser,
) (string, error) {

	if reservationID <= 0 {
		return "", ErrReservationNotFound
	}

	reservation, err :=
		repositories.GetReservationByID(
			reservationID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrReservationNotFound
	}

	if err != nil {
		return "", err
	}

	// ---------------------------------------------------------------------
	// Autorización.
	// ---------------------------------------------------------------------
	//
	// Un usuario normal solo puede rotar el código de una reserva que él
	// mismo creó. Los administradores pueden hacerlo como operación de
	// soporte institucional.
	if !requestedBy.IsAdmin &&
		reservation.UserID != requestedBy.ID {

		return "", ErrReservationForbidden
	}

	return repositories.RotateReservationJoinCode(
		reservationID,
	)
}
