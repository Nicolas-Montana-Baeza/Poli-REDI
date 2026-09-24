package repositories

import (
	"context"
	"database/sql"
	"errors"

	"poli-redi-api/internal/appscope"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

var ErrNotificationNotFound = errors.New(
	"notificacion no encontrada",
)

type notificationExecer interface {
	ExecContext(
		context.Context,
		string,
		...any,
	) (sql.Result, error)
}

type reservationNotificationSpec struct {
	Title   string
	Message string
	Type    string
}

func notificationsEnabled() bool {
	return appscope.HasMVP3()
}

func createNotificationTx(
	ctx context.Context,
	execer notificationExecer,
	userID int,
	reservationID int,
	spec reservationNotificationSpec,
) error {
	if !notificationsEnabled() {
		return nil
	}

	_, err := execer.ExecContext(
		ctx,
		`
		INSERT INTO notifications (
			user_id,
			reservation_id,
			title,
			message,
			type
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5
		)
		`,
		userID,
		reservationID,
		spec.Title,
		spec.Message,
		spec.Type,
	)

	return err
}

func participantNotificationForTransition(
	oldReservationStatus string,
	newReservationStatus string,
	oldGroupCondition string,
	newGroupCondition string,
) (reservationNotificationSpec, bool) {
	switch {
	case oldReservationStatus == "PENDING" &&
		newReservationStatus == "CONFIRMED":

		return reservationNotificationSpec{
			Title: "Reserva confirmada",
			Message: "Tu reserva alcanzó el mínimo de participantes " +
				"y quedó confirmada.",
			Type: "RESERVATION_CONFIRMED",
		}, true

	case oldReservationStatus == "CONFIRMED" &&
		newReservationStatus == "CONFIRMED" &&
		oldGroupCondition == GroupConditionHealthy &&
		newGroupCondition == GroupConditionAtRisk:

		return reservationNotificationSpec{
			Title: "Reserva en riesgo",
			Message: "Tu reserva sigue confirmada, pero quedó bajo " +
				"el mínimo de participantes.",
			Type: "RESERVATION_AT_RISK",
		}, true

	case oldReservationStatus == "CONFIRMED" &&
		newReservationStatus == "CONFIRMED" &&
		oldGroupCondition == GroupConditionAtRisk &&
		newGroupCondition == GroupConditionHealthy:

		return reservationNotificationSpec{
			Title: "Reserva recuperada",
			Message: "Tu reserva recuperó el mínimo de participantes " +
				"requerido.",
			Type: "RESERVATION_RECOVERED",
		}, true
	}

	return reservationNotificationSpec{}, false
}

var ErrNotificationTargetNotFound = errors.New(
	"destinatario de notificacion no encontrado",
)

func administrativeCancellationNotification(
	actorIsAdmin bool,
	actorUserID int,
	ownerUserID int,
) (reservationNotificationSpec, bool) {
	if !actorIsAdmin ||
		actorUserID <= 0 ||
		ownerUserID <= 0 ||
		actorUserID == ownerUserID {

		return reservationNotificationSpec{}, false
	}

	return reservationNotificationSpec{
		Title:   "Reserva cancelada",
		Message: "Tu reserva fue cancelada por administración.",
		Type:    "RESERVATION_CANCELLED",
	}, true
}

func institutionalCancellationNotification() reservationNotificationSpec {
	return reservationNotificationSpec{
		Title: "Reserva cancelada",
		Message: "Tu reserva fue cancelada debido a una decisión de " +
			"programación institucional.",
		Type: "RESERVATION_CANCELLED_INSTITUTIONAL",
	}
}

func createReservationOwnerNotificationTx(
	ctx context.Context,
	execer notificationExecer,
	reservationID int,
	spec reservationNotificationSpec,
) error {
	if !notificationsEnabled() {
		return nil
	}

	result, err := execer.ExecContext(
		ctx,
		`
		INSERT INTO notifications (
			user_id,
			reservation_id,
			title,
			message,
			type
		)
		SELECT
			reservation.user_id,
			reservation.id,
			$2,
			$3,
			$4
		FROM reservations reservation
		WHERE reservation.id = $1
		`,
		reservationID,
		spec.Title,
		spec.Message,
		spec.Type,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if affected != 1 {
		return ErrNotificationTargetNotFound
	}

	return nil
}

func MarkNotificationRead(
	userID int,
	notificationID int,
) (models.Notification, error) {
	var notification models.Notification

	err := database.DB.QueryRowContext(
		context.Background(),
		`
		UPDATE notifications
		SET is_read = true
		WHERE id = $1
		  AND user_id = $2
		RETURNING
			id,
			user_id,
			reservation_id,
			title,
			message,
			type,
			is_read,
			created_at
		`,
		notificationID,
		userID,
	).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.ReservationID,
		&notification.Title,
		&notification.Message,
		&notification.Type,
		&notification.IsRead,
		&notification.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return models.Notification{}, ErrNotificationNotFound
	}

	if err != nil {
		return models.Notification{}, err
	}

	return notification, nil
}

func GetNotificationsByUserID(
	userID int,
) ([]models.Notification, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			id,
			user_id,
			reservation_id,
			title,
			message,
			type,
			is_read,
			created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY
			is_read ASC,
			created_at DESC
		LIMIT 20
		`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	notifications := []models.Notification{}

	for rows.Next() {
		var notification models.Notification

		if err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.ReservationID,
			&notification.Title,
			&notification.Message,
			&notification.Type,
			&notification.IsRead,
			&notification.CreatedAt,
		); err != nil {
			return nil, err
		}

		notifications = append(
			notifications,
			notification,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}
