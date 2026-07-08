package repositories

import (
	"context"
	"database/sql"
	"errors"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func GetActiveWorkshopsForUser(userID int) ([]models.Workshop, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			w.id,
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

func EnrollUserInWorkshop(
	workshopID int,
	userID int,
) (models.Workshop, error) {
	tx, err := database.DB.BeginTx(
		context.Background(),
		&sql.TxOptions{Isolation: sql.LevelSerializable},
	)

	if err != nil {
		return models.Workshop{}, err
	}

	defer tx.Rollback()

	var capacity int
	var isActive bool
	var enrolledCount int
	var isAlreadyEnrolled bool

	err = tx.QueryRowContext(
		context.Background(),
		`
		SELECT
			w.capacity,
			w.is_active,
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
		&isActive,
		&enrolledCount,
		&isAlreadyEnrolled,
	)

	if err != nil {
		return models.Workshop{}, err
	}

	if !isActive {
		return models.Workshop{}, errors.New("el taller no esta disponible")
	}

	if isAlreadyEnrolled {
		return models.Workshop{}, errors.New("ya estas inscrito en este taller")
	}

	if enrolledCount >= capacity {
		return models.Workshop{}, errors.New("el taller no tiene cupos disponibles")
	}

	_, err = tx.ExecContext(
		context.Background(),
		`
		INSERT INTO dbo.workshop_enrollments (
			workshop_id,
			user_id,
			status
		)
		VALUES (@p1, @p2, 'CONFIRMED');
		`,
		workshopID,
		userID,
	)

	if err != nil {
		return models.Workshop{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.Workshop{}, err
	}

	return GetWorkshopForUser(workshopID, userID)
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
