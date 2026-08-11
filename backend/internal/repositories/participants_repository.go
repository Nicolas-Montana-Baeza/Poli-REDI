package repositories

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/joinsecret"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/validators"
	"time"

	mssql "github.com/microsoft/go-mssqldb"
)

var ErrInvalidJoinCode = errors.New("solicitud grupal no encontrada")
var ErrParticipantIneligible = errors.New("la cuenta debe estar activa y tener RUT registrado")
var ErrGroupCapacity = errors.New("la solicitud alcanzo su capacidad")
var ErrOwnerCannotWithdraw = errors.New("el solicitante no puede retirarse")
var ErrParticipationDeadline = errors.New("el plazo de confirmacion ya vencio")
var ErrParticipantConflict = errors.New("ya tienes una reserva activa en ese horario")

func codeHash(code string) string { s := sha256.Sum256([]byte(code)); return hex.EncodeToString(s[:]) }

func assembleReservationProgress(reservationID int, status string, count, minimum, target, capacity int, start time.Time, deadlineMinutes int, isOwner, isMember bool) models.ReservationProgress {
	deadline := businessclock.ConfirmationDeadline(start, deadlineMinutes)
	return models.ReservationProgress{
		ReservationID: reservationID, Status: status, ParticipantCount: count,
		MinimumParticipants: minimum, TargetParticipants: target, Capacity: capacity,
		ConfirmationDeadline: deadline,
		CanEditTarget:        isOwner && !businessclock.Now().After(deadline),
		IsOwner:              isOwner, IsMember: isMember,
	}
}

// GetReservationProgress retrieves the current progress and participation metrics of a group reservation using its join code.
// It verifies the join code and returns the status, target participants, and deadlines.
// Before querying, it expires any pending group reservations that have passed their deadline.
func GetReservationProgress(code string, userID int) (models.ReservationProgress, error) {
	if err := ExpirePendingGroupReservations(businessclock.Now()); err != nil {
		return models.ReservationProgress{}, err
	}
	var p models.ReservationProgress
	var start time.Time
	var deadlineMinutes int
	err := database.DB.QueryRowContext(context.Background(), `SELECT r.id,r.status,COUNT(CASE WHEN pa.status='CONFIRMED' THEN 1 END),pol.minimum_participants,r.group_capacity_snapshot,COALESCE(r.target_participants,r.group_capacity_snapshot),r.start_time,pol.confirmation_deadline_minutes,CASE WHEN EXISTS(SELECT 1 FROM dbo.participants mine WHERE mine.reservation_id=r.id AND mine.user_id=@p2 AND mine.status='CONFIRMED') THEN 1 ELSE 0 END,CASE WHEN r.user_id=@p2 THEN 1 ELSE 0 END FROM dbo.reservations r INNER JOIN dbo.reservation_policies pol ON pol.id=r.policy_id LEFT JOIN dbo.participants pa ON pa.reservation_id=r.id WHERE r.join_code_hash=@p1 AND r.group_capacity_snapshot IS NOT NULL AND r.status IN ('PENDING','CONFIRMED') GROUP BY r.id,r.status,pol.minimum_participants,r.group_capacity_snapshot,r.target_participants,r.start_time,pol.confirmation_deadline_minutes,r.user_id`, codeHash(code), userID).Scan(&p.ReservationID, &p.Status, &p.ParticipantCount, &p.MinimumParticipants, &p.Capacity, &p.TargetParticipants, &start, &deadlineMinutes, &p.IsMember, &p.IsOwner)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrInvalidJoinCode
	}
	return assembleReservationProgress(p.ReservationID, p.Status, p.ParticipantCount, p.MinimumParticipants, p.TargetParticipants, p.Capacity, start, deadlineMinutes, p.IsOwner, p.IsMember), err
}

const userActiveOverlapSQL = `SELECT TOP (1) 1
	FROM dbo.reservations existing WITH(UPDLOCK,HOLDLOCK)
	WHERE existing.id<>@p2
	  AND existing.status IN('PENDING','CONFIRMED')
	  AND existing.start_time < DATEADD(MINUTE, @p3, @p4)
	  AND DATEADD(MINUTE, existing.duration_minutes, existing.start_time) > @p5
	  AND (
		existing.user_id=@p1
		OR EXISTS (
			SELECT 1
			FROM dbo.participants membership WITH(UPDLOCK,HOLDLOCK)
			WHERE membership.reservation_id=existing.id
			  AND membership.user_id=@p1
			  AND membership.status='CONFIRMED'
		)
	  )`

func userHasActiveOverlapTx(ctx context.Context, tx *sql.Tx, userID, reservationID int, start time.Time, durationMinutes int) (bool, error) {
	var found int
	err := tx.QueryRowContext(
		ctx,
		userActiveOverlapSQL,
		userID,
		reservationID,
		durationMinutes,
		start,
		start,
	).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func ChangeParticipation(code string, userID int, confirm bool) (models.ReservationProgress, error) {
	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return models.ReservationProgress{}, mapParticipationDatabaseError(err)
	}
	defer tx.Rollback()
	var reservationID, capacity, minimum, deadlineMinutes, ownerID int
	var oldReservationStatus, cancellationReason string
	var target int
	var start time.Time
	var durationMinutes int
	err = tx.QueryRowContext(ctx, `SELECT r.id,r.group_capacity_snapshot,p.minimum_participants,COALESCE(r.target_participants,r.group_capacity_snapshot),r.status,r.start_time,r.duration_minutes,p.confirmation_deadline_minutes,r.user_id,COALESCE(r.cancellation_reason,'') FROM dbo.reservations r WITH(UPDLOCK,HOLDLOCK) INNER JOIN dbo.reservation_policies p ON p.id=r.policy_id WHERE r.join_code_hash=@p1 AND r.group_capacity_snapshot IS NOT NULL AND r.status IN('PENDING','CONFIRMED','CANCELLED')`, codeHash(code)).Scan(&reservationID, &capacity, &minimum, &target, &oldReservationStatus, &start, &durationMinutes, &deadlineMinutes, &ownerID, &cancellationReason)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ReservationProgress{}, ErrInvalidJoinCode
	}
	if err != nil {
		return models.ReservationProgress{}, mapParticipationDatabaseError(err)
	}
	if oldReservationStatus == models.ReservationStatusCancelled {
		if cancellationReason == "CONFIRMATION_DEADLINE" {
			return models.ReservationProgress{}, ErrParticipationDeadline
		}
		return models.ReservationProgress{}, ErrInvalidJoinCode
	}
	deadline := businessclock.ConfirmationDeadline(start, deadlineMinutes)
	if participationDeadlineClosed(businessclock.Now(), deadline) {
		if oldReservationStatus == models.ReservationStatusPending {
			if _, err = expirePendingGroupTx(ctx, tx, reservationID, ownerID, minimum); err != nil {
				return models.ReservationProgress{}, err
			}
		}
		if err = tx.Commit(); err != nil {
			return models.ReservationProgress{}, err
		}
		return models.ReservationProgress{}, ErrParticipationDeadline
	}
	var rut string
	var blocked bool
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(rut,''),is_blocked FROM dbo.users WITH(UPDLOCK,HOLDLOCK) WHERE id=@p1`, userID).Scan(&rut, &blocked); err != nil {
		return models.ReservationProgress{}, err
	}
	if blocked || !validators.HasRUT(rut) {
		return models.ReservationProgress{}, ErrParticipantIneligible
	}
	if confirm {
		overlaps, err := userHasActiveOverlapTx(ctx, tx, userID, reservationID, start, durationMinutes)
		if err != nil {
			return models.ReservationProgress{}, err
		}
		if overlaps {
			return models.ReservationProgress{}, ErrParticipantConflict
		}
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
	mutate, newStatus, calculatedStatus, transitionErr := participantTransition(participantExists, oldStatus, isOwner, priorCount, target, minimum, confirm)
	if transitionErr != nil {
		return models.ReservationProgress{}, transitionErr
	}
	if !mutate {
		if err = tx.Commit(); err != nil {
			return models.ReservationProgress{}, err
		}
		return assembleReservationProgress(reservationID, calculatedStatus, priorCount, minimum, target, capacity, start, deadlineMinutes, ownerID == userID, participantExists && oldStatus == "CONFIRMED"), nil
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
		return models.ReservationProgress{}, mapParticipationDatabaseError(err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO dbo.reservation_participant_audit(reservation_id,actor_user_id,participant_user_id,action,previous_status,new_status,previous_reservation_status,new_reservation_status) VALUES(@p1,@p2,@p2,@p3,@p4,@p5,@p6,@p7)`, reservationID, userID, action, sql.NullString{String: oldStatus, Valid: oldStatus != ""}, newStatus, oldReservationStatus, newReservationStatus); err != nil {
		return models.ReservationProgress{}, err
	}
	if err = tx.Commit(); err != nil {
		return models.ReservationProgress{}, mapParticipationDatabaseError(err)
	}
	return assembleReservationProgress(reservationID, newReservationStatus, count, minimum, target, capacity, start, deadlineMinutes, ownerID == userID, newStatus == "CONFIRMED"), nil
}

func mapParticipationDatabaseError(err error) error {
	var sqlErr mssql.Error
	if errors.As(err, &sqlErr) && sqlErr.Number == 51023 {
		return ErrParticipantConflict
	}
	return err
}

func expirePendingGroupTx(ctx context.Context, tx *sql.Tx, reservationID, ownerID, minimum int) (bool, error) {
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM dbo.participants WITH(UPDLOCK,HOLDLOCK) WHERE reservation_id=@p1 AND status='CONFIRMED'`, reservationID).Scan(&count); err != nil {
		return false, err
	}
	if count >= minimum {
		return false, nil
	}
	result, err := tx.ExecContext(ctx, `UPDATE dbo.reservations SET status='CANCELLED',cancellation_reason='CONFIRMATION_DEADLINE',updated_at=SYSUTCDATETIME() WHERE id=@p1 AND status='PENDING'`, reservationID)
	if err != nil {
		return false, err
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		return false, nil
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO dbo.reservation_group_expirations(reservation_id,participant_count,minimum_participants) VALUES(@p1,@p2,@p3)`, reservationID, count, minimum); err != nil {
		return false, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO dbo.notifications(user_id,reservation_id,title,message,type) VALUES(@p1,@p2,'Solicitud grupal cancelada','No se alcanzo el minimo antes del plazo.','RESERVATION_CANCELLED')`, ownerID, reservationID)
	return true, err
}

func ExpirePendingGroupReservations(now time.Time) error {
	rows, err := database.DB.QueryContext(context.Background(), `SELECT r.id,r.user_id,r.start_time,p.confirmation_deadline_minutes,p.minimum_participants FROM dbo.reservations r INNER JOIN dbo.reservation_policies p ON p.id=r.policy_id WHERE r.status='PENDING' AND r.group_capacity_snapshot IS NOT NULL`)
	if err != nil {
		return err
	}
	type candidate struct {
		id, owner, minutes, minimum int
		start                       time.Time
	}
	var candidates []candidate
	for rows.Next() {
		var c candidate
		if err = rows.Scan(&c.id, &c.owner, &c.start, &c.minutes, &c.minimum); err != nil {
			rows.Close()
			return err
		}
		if now.After(businessclock.ConfirmationDeadline(c.start, c.minutes)) {
			candidates = append(candidates, c)
		}
	}
	rows.Close()
	for _, c := range candidates {
		ctx := context.Background()
		tx, beginErr := database.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if beginErr != nil {
			return beginErr
		}
		var status string
		if beginErr = tx.QueryRowContext(ctx, `SELECT status FROM dbo.reservations WITH(UPDLOCK,HOLDLOCK) WHERE id=@p1`, c.id).Scan(&status); beginErr == nil && status == "PENDING" {
			_, beginErr = expirePendingGroupTx(ctx, tx, c.id, c.owner, c.minimum)
		}
		if beginErr == nil {
			beginErr = tx.Commit()
		} else {
			tx.Rollback()
		}
		if beginErr != nil {
			return beginErr
		}
	}
	return nil
}

func GetOwnerJoinCode(reservationID, userID int) (string, error) {
	var nonce, ciphertext []byte
	var version int
	err := database.DB.QueryRowContext(context.Background(), `SELECT s.nonce,s.ciphertext,s.key_version FROM dbo.reservations r INNER JOIN dbo.reservation_join_code_secrets s ON s.reservation_id=r.id WHERE r.id=@p1 AND r.user_id=@p2 AND r.group_capacity_snapshot IS NOT NULL AND r.status IN('PENDING','CONFIRMED')`, reservationID, userID).Scan(&nonce, &ciphertext, &version)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrInvalidJoinCode
	}
	if err != nil {
		return "", err
	}
	return joinsecret.Decrypt(nonce, ciphertext, version, reservationID)
}

func RotateOwnerJoinCode(reservationID, userID int) (string, error) {
	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var owner int
	if err = tx.QueryRowContext(ctx, `SELECT user_id FROM dbo.reservations WITH(UPDLOCK,HOLDLOCK) WHERE id=@p1 AND group_capacity_snapshot IS NOT NULL AND status IN('PENDING','CONFIRMED')`, reservationID).Scan(&owner); errors.Is(err, sql.ErrNoRows) || owner != userID {
		return "", ErrInvalidJoinCode
	}
	if err != nil {
		return "", err
	}
	raw := make([]byte, 18)
	if _, err = rand.Read(raw); err != nil {
		return "", err
	}
	code := base64.RawURLEncoding.EncodeToString(raw)
	nonce, ciphertext, version, err := joinsecret.Encrypt(code, reservationID)
	if err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE dbo.reservations SET join_code_hash=@p2,updated_at=SYSUTCDATETIME() WHERE id=@p1`, reservationID, codeHash(code)); err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `MERGE dbo.reservation_join_code_secrets AS t USING(SELECT @p1 reservation_id) s ON t.reservation_id=s.reservation_id WHEN MATCHED THEN UPDATE SET key_version=@p2,nonce=@p3,ciphertext=@p4,rotated_at=SYSUTCDATETIME() WHEN NOT MATCHED THEN INSERT(reservation_id,key_version,nonce,ciphertext) VALUES(@p1,@p2,@p3,@p4);`, reservationID, version, nonce, ciphertext); err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return code, nil
}

var ErrTargetForbidden = errors.New("solo el solicitante puede modificar el objetivo")
var ErrTargetDeadline = errors.New("el plazo para modificar el objetivo ya vencio")
var ErrTargetBelowConfirmed = errors.New("el objetivo no puede ser menor que los participantes confirmados")

func UpdateReservationTarget(reservationID, userID, target int, now time.Time) (models.ReservationProgress, error) {
	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return models.ReservationProgress{}, err
	}
	defer tx.Rollback()
	var owner, minimum, capacity int
	var oldTarget sql.NullInt64
	var status string
	var start time.Time
	var deadlineMinutes int
	err = tx.QueryRowContext(ctx, `SELECT r.user_id,p.minimum_participants,r.group_capacity_snapshot,r.target_participants,r.status,r.start_time,p.confirmation_deadline_minutes FROM dbo.reservations r WITH(UPDLOCK,HOLDLOCK) INNER JOIN dbo.reservation_policies p ON p.id=r.policy_id WHERE r.id=@p1 AND r.group_capacity_snapshot IS NOT NULL`, reservationID).Scan(&owner, &minimum, &capacity, &oldTarget, &status, &start, &deadlineMinutes)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ReservationProgress{}, ErrInvalidJoinCode
	}
	if err != nil {
		return models.ReservationProgress{}, err
	}
	if owner != userID {
		return models.ReservationProgress{}, ErrTargetForbidden
	}
	if status != "PENDING" && status != "CONFIRMED" {
		return models.ReservationProgress{}, errors.New("la solicitud no esta activa")
	}
	deadline := businessclock.ConfirmationDeadline(start, deadlineMinutes)
	if !targetDeadlineOpen(now, deadline) {
		if status == models.ReservationStatusPending {
			if _, err = expirePendingGroupTx(ctx, tx, reservationID, owner, minimum); err != nil {
				return models.ReservationProgress{}, err
			}
			if err = tx.Commit(); err != nil {
				return models.ReservationProgress{}, err
			}
		}
		return models.ReservationProgress{}, ErrTargetDeadline
	}
	var count int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM dbo.participants WITH(UPDLOCK,HOLDLOCK) WHERE reservation_id=@p1 AND status='CONFIRMED'`, reservationID).Scan(&count); err != nil {
		return models.ReservationProgress{}, err
	}
	if err := validateTargetChange(target, minimum, capacity, count); err != nil {
		return models.ReservationProgress{}, err
	}
	oldEffective := capacity
	if oldTarget.Valid {
		oldEffective = int(oldTarget.Int64)
	}
	if target != oldEffective {
		if _, err = tx.ExecContext(ctx, `UPDATE dbo.reservations SET target_participants=@p2,updated_at=SYSUTCDATETIME() WHERE id=@p1`, reservationID, target); err != nil {
			return models.ReservationProgress{}, err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO dbo.reservation_target_audit(reservation_id,actor_user_id,old_target_participants,new_target_participants) VALUES(@p1,@p2,@p3,@p4)`, reservationID, userID, oldEffective, target); err != nil {
			return models.ReservationProgress{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return models.ReservationProgress{}, err
	}
	return assembleReservationProgress(reservationID, status, count, minimum, target, capacity, start, deadlineMinutes, true, true), nil
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
