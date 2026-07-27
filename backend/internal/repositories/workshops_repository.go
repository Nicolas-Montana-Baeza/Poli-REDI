package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

var (
	ErrWorkshopNotFound        = errors.New("taller no encontrado o no disponible")
	ErrWorkshopCapacity        = errors.New("el taller no tiene cupos disponibles")
	ErrWorkshopScheduleInvalid = errors.New("el taller no tiene un horario valido")
	ErrWorkshopInternal        = errors.New("no se pudo procesar la inscripción al taller")
)

type WorkshopScheduleConflictError struct {
	Title        string
	DayText      string
	ScheduleText string
}

func (e *WorkshopScheduleConflictError) Error() string {
	return fmt.Sprintf("el horario se superpone con %s", e.Title)
}

func GetActiveWorkshopsForUser(userID int) ([]models.Workshop, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			w.id,
			w.resource_id,
			w.title,
			COALESCE(w.description, '') AS description,
			COALESCE(w.location, '') AS location,
			COALESCE(w.instructor_name, '') AS instructor_name,
			w.day_text,
			w.schedule_text,
			w.capacity,
			(
				SELECT COUNT(*)
				FROM dbo.workshop_enrollments we
				WHERE we.workshop_id = w.id
				  AND we.status = 'CONFIRMED'
			) AS enrolled_count,
			w.is_active,
			CASE
				WHEN EXISTS (
					SELECT 1
					FROM dbo.workshop_enrollments user_we
					WHERE user_we.workshop_id = w.id
					  AND user_we.user_id = @p1
					  AND user_we.status = 'CONFIRMED'
				)
				THEN CAST(1 AS BIT)
				ELSE CAST(0 AS BIT)
			END AS is_enrolled,
			w.created_at,
			w.updated_at
		FROM dbo.workshops w
		WHERE w.is_active = 1
		ORDER BY w.title ASC, w.day_text ASC, w.schedule_text ASC;
		`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	workshops := []models.Workshop{}

	for rows.Next() {
		var workshop models.Workshop

		err := rows.Scan(
			&workshop.ID,
			&workshop.ResourceID,
			&workshop.Title,
			&workshop.Description,
			&workshop.Location,
			&workshop.InstructorName,
			&workshop.DayText,
			&workshop.ScheduleText,
			&workshop.Capacity,
			&workshop.EnrolledCount,
			&workshop.IsActive,
			&workshop.IsEnrolled,
			&workshop.CreatedAt,
			&workshop.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		workshops = append(workshops, workshop)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return workshops, nil
}

func GetActiveWorkshops() ([]models.Workshop, error) {
	return GetActiveWorkshopsForUser(0)
}

func EnrollUserInWorkshop(
	workshopID int,
	userID int,
) (models.Workshop, bool, error) {
	ctx := context.Background()
	tx, err := database.DB.BeginTx(
		ctx,
		&sql.TxOptions{Isolation: sql.LevelSerializable},
	)

	if err != nil {
		return models.Workshop{}, false, err
	}

	defer tx.Rollback()

	var capacity int
	var enrolledCount int
	var isAlreadyEnrolled bool

	// Serializa todas las inscripciones de un usuario para evitar carreras entre talleres.
	var lockedUserID int
	if err := tx.QueryRowContext(ctx,
		`SELECT id FROM dbo.users WITH (UPDLOCK, HOLDLOCK) WHERE id = @p1;`,
		userID,
	).Scan(&lockedUserID); err != nil {
		return models.Workshop{}, false, err
	}

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT
			w.capacity,
			(
				SELECT COUNT(*)
				FROM dbo.workshop_enrollments we
				WHERE we.workshop_id = w.id
				  AND we.status = 'CONFIRMED'
			) AS enrolled_count,
			CASE
				WHEN EXISTS (
					SELECT 1
					FROM dbo.workshop_enrollments user_we
					WHERE user_we.workshop_id = w.id
					  AND user_we.user_id = @p2
					  AND user_we.status = 'CONFIRMED'
				)
				THEN CAST(1 AS BIT)
				ELSE CAST(0 AS BIT)
			END AS is_already_enrolled
		FROM dbo.workshops w WITH (UPDLOCK, HOLDLOCK)
		WHERE w.id = @p1
		  AND w.is_active = 1;
		`,
		workshopID,
		userID,
	).Scan(
		&capacity,
		&enrolledCount,
		&isAlreadyEnrolled,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Workshop{}, false, ErrWorkshopNotFound
		}
		return models.Workshop{}, false, err
	}

	if isAlreadyEnrolled {
		if err := tx.Commit(); err != nil {
			return models.Workshop{}, false, err
		}
		workshop, err := GetWorkshopForUser(workshopID, userID)
		return workshop, false, err
	}

	if enrolledCount >= capacity {
		return models.Workshop{}, false, ErrWorkshopCapacity
	}

	var occurrenceCount int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM dbo.workshop_occurrences WITH (UPDLOCK, HOLDLOCK)
		WHERE workshop_id = @p1
		  AND weekday_iso BETWEEN 1 AND 7
		  AND start_minute >= 0 AND end_minute <= 1440
		  AND start_minute < end_minute;`,
		workshopID,
	).Scan(&occurrenceCount); err != nil {
		return models.Workshop{}, false, err
	}
	if occurrenceCount == 0 {
		return models.Workshop{}, false, ErrWorkshopScheduleInvalid
	}

	var conflict WorkshopScheduleConflictError
	err = tx.QueryRowContext(ctx, `
		SELECT TOP (1) existing_w.title, existing_w.day_text, existing_w.schedule_text
		FROM dbo.workshop_enrollments existing_e WITH (UPDLOCK, HOLDLOCK)
		JOIN dbo.workshops existing_w WITH (HOLDLOCK)
		  ON existing_w.id = existing_e.workshop_id AND existing_w.is_active = 1
		JOIN dbo.workshop_occurrences existing_o WITH (HOLDLOCK)
		  ON existing_o.workshop_id = existing_w.id
		JOIN dbo.workshop_occurrences target_o WITH (HOLDLOCK)
		  ON target_o.workshop_id = @p2
		 AND target_o.weekday_iso = existing_o.weekday_iso
		 AND existing_o.start_minute < target_o.end_minute
		 AND target_o.start_minute < existing_o.end_minute
		WHERE existing_e.user_id = @p1
		  AND existing_e.status = 'CONFIRMED'
		  AND existing_e.workshop_id <> @p2
		ORDER BY existing_w.title;`,
		userID, workshopID,
	).Scan(&conflict.Title, &conflict.DayText, &conflict.ScheduleText)
	if err == nil {
		return models.Workshop{}, false, &conflict
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return models.Workshop{}, false, err
	}

	result, err := tx.ExecContext(
		ctx,
		`
		UPDATE dbo.workshop_enrollments
		   SET status = 'CONFIRMED'
		 WHERE id = (
			SELECT TOP (1) id FROM dbo.workshop_enrollments WITH (UPDLOCK, HOLDLOCK)
			WHERE workshop_id = @p1 AND user_id = @p2 AND status = 'CANCELLED'
			ORDER BY id DESC
		 );
		`,
		workshopID,
		userID,
	)

	if err != nil {
		return models.Workshop{}, false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return models.Workshop{}, false, err
	}
	if affected == 0 {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO dbo.workshop_enrollments (workshop_id, user_id, status)
			VALUES (@p1, @p2, 'CONFIRMED');`,
			workshopID, userID,
		); err != nil {
			return models.Workshop{}, false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return models.Workshop{}, false, err
	}

	workshop, err := GetWorkshopForUser(workshopID, userID)
	return workshop, true, err
}

func GetWorkshopForUser(
	workshopID int,
	userID int,
) (models.Workshop, error) {
	var workshop models.Workshop

	err := database.DB.QueryRowContext(
		context.Background(),
		`
		SELECT
			w.id,
			w.resource_id,
			w.title,
			COALESCE(w.description, '') AS description,
			COALESCE(w.location, '') AS location,
			COALESCE(w.instructor_name, '') AS instructor_name,
			w.day_text,
			w.schedule_text,
			w.capacity,
			(
				SELECT COUNT(*)
				FROM dbo.workshop_enrollments we
				WHERE we.workshop_id = w.id
				  AND we.status = 'CONFIRMED'
			) AS enrolled_count,
			w.is_active,
			CASE
				WHEN EXISTS (
					SELECT 1
					FROM dbo.workshop_enrollments user_we
					WHERE user_we.workshop_id = w.id
					  AND user_we.user_id = @p2
					  AND user_we.status = 'CONFIRMED'
				)
				THEN CAST(1 AS BIT)
				ELSE CAST(0 AS BIT)
			END AS is_enrolled,
			w.created_at,
			w.updated_at
		FROM dbo.workshops w
		WHERE w.id = @p1;
		`,
		workshopID,
		userID,
	).Scan(
		&workshop.ID,
		&workshop.ResourceID,
		&workshop.Title,
		&workshop.Description,
		&workshop.Location,
		&workshop.InstructorName,
		&workshop.DayText,
		&workshop.ScheduleText,
		&workshop.Capacity,
		&workshop.EnrolledCount,
		&workshop.IsActive,
		&workshop.IsEnrolled,
		&workshop.CreatedAt,
		&workshop.UpdatedAt,
	)

	if err != nil {
		return models.Workshop{}, err
	}

	return workshop, nil
}
