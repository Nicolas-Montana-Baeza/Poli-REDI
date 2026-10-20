package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrBlockNotFound            = errors.New("bloqueo no encontrado")
	ErrBlockResourceUnavailable = errors.New("recurso o sede no disponible")
	ErrBlockOverlap             = errors.New("el horario se solapa con otra ocupacion")
)

const adminBlockColumns = `
    b.id, b.resource_id, r.name, b.created_by_user_id,
    b.block_type, COALESCE(b.reason, ''), b.start_time, b.end_time,
    b.is_active, b.created_at`

func scanAdminBlock(scanner interface{ Scan(...any) error }) (models.AdminAvailabilityBlock, error) {
	var block models.AdminAvailabilityBlock
	err := scanner.Scan(
		&block.ID, &block.ResourceID, &block.ResourceName,
		&block.CreatedByUserID, &block.BlockType, &block.Reason,
		&block.StartTime, &block.EndTime, &block.IsActive, &block.CreatedAt,
	)
	return block, err
}

// classifyBlockWriteError evita filtrar detalles SQL hacia el cliente.
func classifyBlockWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23P01", "P1010", "P1012":
			return ErrBlockOverlap
		case "23503":
			return ErrBlockResourceUnavailable
		}
	}
	return err
}

// CreateAdminAvailabilityBlock mantiene el candado de recurso/sede hasta
// commit. Los triggers heredados serializan reservas/actividades por recurso,
// y PG16_0012 protege tambien inserciones concurrentes bloque-bloque.
func CreateAdminAvailabilityBlock(input models.AdminAvailabilityBlock) (models.AdminAvailabilityBlock, error) {
	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return models.AdminAvailabilityBlock{}, err
	}
	defer tx.Rollback()

	var resourceID int
	err = tx.QueryRowContext(ctx, `
        SELECT r.id FROM resources r
        JOIN venues v ON v.id = r.venue_id
        WHERE r.id = $1 AND r.is_active AND v.is_active
        FOR SHARE OF r, v`, input.ResourceID).Scan(&resourceID)
	if errors.Is(err, sql.ErrNoRows) {
		return models.AdminAvailabilityBlock{}, ErrBlockResourceUnavailable
	}
	if err != nil {
		return models.AdminAvailabilityBlock{}, err
	}

	block, err := scanAdminBlock(tx.QueryRowContext(ctx, `
        WITH inserted AS (
            INSERT INTO availability_blocks
                (resource_id, created_by_user_id, block_type, reason, start_time, end_time)
            VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6)
            RETURNING *
        )
        SELECT `+adminBlockColumns+`
        FROM inserted b JOIN resources r ON r.id = b.resource_id`,
		input.ResourceID, input.CreatedByUserID, input.BlockType,
		input.Reason, input.StartTime, input.EndTime))
	if err != nil {
		return models.AdminAvailabilityBlock{}, classifyBlockWriteError(err)
	}
	if err = tx.Commit(); err != nil {
		return models.AdminAvailabilityBlock{}, classifyBlockWriteError(err)
	}
	return block, nil
}

// ListAdminAvailabilityBlocks incluye activos e inactivos para auditoria.
// El intervalo es semiabierto, igual que el contrato de disponibilidad.
func ListAdminAvailabilityBlocks(from, to time.Time) ([]models.AdminAvailabilityBlock, error) {
	ctx := context.Background()
	rows, err := database.DB.QueryContext(ctx, `
        SELECT `+adminBlockColumns+`
        FROM availability_blocks b JOIN resources r ON r.id = b.resource_id
        WHERE b.start_time < $2 AND b.end_time > $1
        ORDER BY b.start_time ASC, b.id ASC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]models.AdminAvailabilityBlock, 0)
	for rows.Next() {
		block, err := scanAdminBlock(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, block)
	}
	return result, rows.Err()
}

// La desactivacion es idempotente y conserva el registro historico.
func DeactivateAdminAvailabilityBlock(id int) (models.AdminAvailabilityBlock, error) {
	block, err := scanAdminBlock(database.DB.QueryRowContext(context.Background(), `
        WITH updated AS (
            UPDATE availability_blocks SET is_active = false
            WHERE id = $1
            RETURNING *
        )
        SELECT `+adminBlockColumns+`
        FROM updated b JOIN resources r ON r.id = b.resource_id`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return models.AdminAvailabilityBlock{}, ErrBlockNotFound
	}
	return block, err
}
