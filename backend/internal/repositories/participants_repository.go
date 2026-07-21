package repositories

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
	"strings"
)

var ErrInvalidJoinCode = errors.New("solicitud grupal no encontrada")
var ErrParticipantIneligible = errors.New("la cuenta debe estar activa y tener RUT registrado")
var ErrGroupCapacity = errors.New("la solicitud alcanzo su capacidad")
var ErrOwnerCannotWithdraw = errors.New("el solicitante no puede retirarse")

func codeHash(code string) string { s := sha256.Sum256([]byte(code)); return hex.EncodeToString(s[:]) }

func GetReservationProgress(code string, userID int) (models.ReservationProgress, error) {
	var p models.ReservationProgress
	err := database.DB.QueryRowContext(context.Background(), `SELECT r.id,r.status,COUNT(CASE WHEN pa.status='CONFIRMED' THEN 1 END),pol.minimum_participants,r.group_capacity_snapshot,CASE WHEN EXISTS(SELECT 1 FROM dbo.participants mine WHERE mine.reservation_id=r.id AND mine.user_id=@p2 AND mine.status='CONFIRMED') THEN 1 ELSE 0 END FROM dbo.reservations r INNER JOIN dbo.reservation_policies pol ON pol.id=r.policy_id LEFT JOIN dbo.participants pa ON pa.reservation_id=r.id WHERE r.join_code_hash=@p1 AND r.group_capacity_snapshot IS NOT NULL AND r.status IN ('PENDING','CONFIRMED') GROUP BY r.id,r.status,pol.minimum_participants,r.group_capacity_snapshot`, codeHash(code), userID).Scan(&p.ReservationID, &p.Status, &p.ParticipantCount, &p.MinimumParticipants, &p.Capacity, &p.IsMember)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrInvalidJoinCode
	}
	return p, err
}

func ChangeParticipation(code string, userID int, confirm bool) (models.ReservationProgress, error) {
	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return models.ReservationProgress{}, err
	}
	defer tx.Rollback()
	var reservationID, capacity, minimum int
	var oldReservationStatus string
	err = tx.QueryRowContext(ctx, `SELECT r.id,r.group_capacity_snapshot,p.minimum_participants,r.status FROM dbo.reservations r WITH(UPDLOCK,HOLDLOCK) INNER JOIN dbo.reservation_policies p ON p.id=r.policy_id WHERE r.join_code_hash=@p1 AND r.group_capacity_snapshot IS NOT NULL AND r.status IN('PENDING','CONFIRMED')`, codeHash(code)).Scan(&reservationID, &capacity, &minimum, &oldReservationStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ReservationProgress{}, ErrInvalidJoinCode
	}
	if err != nil {
		return models.ReservationProgress{}, err
	}
	var rut string
	var blocked bool
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(rut,''),is_blocked FROM dbo.users WITH(UPDLOCK,HOLDLOCK) WHERE id=@p1`, userID).Scan(&rut, &blocked); err != nil {
		return models.ReservationProgress{}, err
	}
	if blocked || strings.TrimSpace(rut) == "" {
		return models.ReservationProgress{}, ErrParticipantIneligible
	}
	var oldStatus string
	var isOwner bool
	err = tx.QueryRowContext(ctx, `SELECT status,is_owner FROM dbo.participants WITH(UPDLOCK,HOLDLOCK) WHERE reservation_id=@p1 AND user_id=@p2`, reservationID, userID).Scan(&oldStatus, &isOwner)
	participantExists := err == nil
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return models.ReservationProgress{}, err
	}
	action := "CONFIRM"
	if !confirm {
		action = "WITHDRAW"
	}
	var priorCount int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM dbo.participants WITH(UPDLOCK,HOLDLOCK) WHERE reservation_id=@p1 AND status='CONFIRMED'`, reservationID).Scan(&priorCount); err != nil {
		return models.ReservationProgress{}, err
	}
	mutate, newStatus, calculatedStatus, transitionErr := participantTransition(participantExists, oldStatus, isOwner, priorCount, capacity, minimum, confirm)
	if transitionErr != nil {
		return models.ReservationProgress{}, transitionErr
	}
	if !mutate {
		if err = tx.Commit(); err != nil {
			return models.ReservationProgress{}, err
		}
		return models.ReservationProgress{ReservationID: reservationID, Status: calculatedStatus, ParticipantCount: priorCount, MinimumParticipants: minimum, Capacity: capacity, IsMember: participantExists && oldStatus == "CONFIRMED"}, nil
	}
	if !participantExists {
		_, err = tx.ExecContext(ctx, `INSERT INTO dbo.participants(reservation_id,user_id,status,confirmed_at,is_owner) VALUES(@p1,@p2,@p3,CASE WHEN @p3='CONFIRMED' THEN SYSUTCDATETIME() END,0)`, reservationID, userID, newStatus)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE dbo.participants SET status=@p3,confirmed_at=CASE WHEN @p3='CONFIRMED' THEN SYSUTCDATETIME() ELSE confirmed_at END,updated_at=SYSUTCDATETIME() WHERE reservation_id=@p1 AND user_id=@p2`, reservationID, userID, newStatus)
	}
	if err != nil {
		return models.ReservationProgress{}, err
	}
	var count int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM dbo.participants WHERE reservation_id=@p1 AND status='CONFIRMED'`, reservationID).Scan(&count); err != nil {
		return models.ReservationProgress{}, err
	}
	newReservationStatus := "PENDING"
	if count >= minimum {
		newReservationStatus = "CONFIRMED"
	}
	if _, err = tx.ExecContext(ctx, `UPDATE dbo.reservations SET status=@p2,updated_at=SYSUTCDATETIME() WHERE id=@p1`, reservationID, newReservationStatus); err != nil {
		return models.ReservationProgress{}, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO dbo.reservation_participant_audit(reservation_id,actor_user_id,participant_user_id,action,previous_status,new_status,previous_reservation_status,new_reservation_status) VALUES(@p1,@p2,@p2,@p3,@p4,@p5,@p6,@p7)`, reservationID, userID, action, sql.NullString{String: oldStatus, Valid: oldStatus != ""}, newStatus, oldReservationStatus, newReservationStatus); err != nil {
		return models.ReservationProgress{}, err
	}
	if err = tx.Commit(); err != nil {
		return models.ReservationProgress{}, err
	}
	return models.ReservationProgress{ReservationID: reservationID, Status: newReservationStatus, ParticipantCount: count, MinimumParticipants: minimum, Capacity: capacity, IsMember: newStatus == "CONFIRMED"}, nil
}

func GetReservationParticipants(reservationID int) ([]models.ReservationParticipant, error) {
	rows, err := database.DB.QueryContext(context.Background(), `SELECT p.user_id,u.full_name,u.email,COALESCE(u.rut,''),p.is_owner,p.status FROM dbo.participants p INNER JOIN dbo.users u ON u.id=p.user_id WHERE p.reservation_id=@p1 ORDER BY p.is_owner DESC,p.id`, reservationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.ReservationParticipant{}
	for rows.Next() {
		var p models.ReservationParticipant
		if err := rows.Scan(&p.UserID, &p.FullName, &p.Email, &p.RUT, &p.IsOwner, &p.Status); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
