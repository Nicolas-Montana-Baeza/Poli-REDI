package repositories

import (
	"context"
	"database/sql"
	"errors"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrResourceDuplicate = errors.New(
		"ya existe un recurso con ese nombre en la sede",
	)
	ErrResourceConcurrentUpdate = errors.New(
		"la configuracion cambio durante la operacion; reintenta",
	)
)

type CreateResourceInput struct {
	VenueID         int
	Name            string
	Type            string
	ReservationMode string
	ImageURL        string
	Capacity        *int
	IsActive        bool
}

func createResourceDBError(err error) error {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" &&
			pgErr.ConstraintName == "uq_resources_venue_name" {
			return ErrResourceDuplicate
		}

		if pgErr.Code == "40001" || pgErr.Code == "40P01" {
			return ErrResourceConcurrentUpdate
		}
	}

	return err
}

func CreateResource(input CreateResourceInput) (models.Resource, error) {
	ctx := context.Background()

	tx, err := database.DB.BeginTx(
		ctx,
		&sql.TxOptions{Isolation: sql.LevelSerializable},
	)
	if err != nil {
		return models.Resource{}, createResourceDBError(err)
	}
	defer tx.Rollback()

	// Mismo orden de bloqueo que la administracion de sedes:
	// primero sede, despues escritura del recurso.
	var venueActive bool

	err = tx.QueryRowContext(ctx, `
		SELECT is_active
		FROM venues
		WHERE id = $1
		FOR UPDATE
	`, input.VenueID).Scan(&venueActive)

	if errors.Is(err, sql.ErrNoRows) {
		return models.Resource{}, ErrVenueNotFound
	}
	if err != nil {
		return models.Resource{}, createResourceDBError(err)
	}

	if input.IsActive && !venueActive {
		return models.Resource{}, ErrResourceVenueInactive
	}

	var resourceID int

	err = tx.QueryRowContext(ctx, `
		INSERT INTO resources (
			venue_id,
			name,
			type,
			reservation_mode,
			image_url,
			capacity,
			is_active
		)
		VALUES (
			$1, $2, $3, $4,
			NULLIF($5, ''),
			$6, $7
		)
		RETURNING id
	`,
		input.VenueID,
		input.Name,
		input.Type,
		input.ReservationMode,
		input.ImageURL,
		input.Capacity,
		input.IsActive,
	).Scan(&resourceID)

	if err != nil {
		return models.Resource{}, createResourceDBError(err)
	}

	if err := tx.Commit(); err != nil {
		return models.Resource{}, createResourceDBError(err)
	}

	return GetResourceByID(resourceID)
}
