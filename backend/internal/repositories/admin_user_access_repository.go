package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

var (
	ErrUserAccessUnauthorized = errors.New("solo administradores activos pueden gestionar accesos")
	ErrUserAccessNotFound     = errors.New("usuario no encontrado")
	ErrUserAccessSelf         = errors.New("no puedes bloquear ni desbloquear tu propia cuenta")
	ErrUserAccessUnchanged    = errors.New("el usuario ya tiene ese estado de acceso")
	ErrUserAccessLastAdmin    = errors.New("la operación dejaría sin administradores activos al sistema")
	ErrUserAccessConflict     = errors.New("la cuenta cambió simultáneamente; actualiza e inténtalo nuevamente")
)

// ChangeUserAccess registra el nuevo estado y su auditoría en la misma transacción.
// SERIALIZABLE protege la comprobación del último administrador ante escrituras concurrentes.
// FOR UPDATE asegura que el administrador que autoriza permanece activo durante el cambio.
func ChangeUserAccess(actorID, targetID int, blocked bool, reason string) (models.AdminUserAccessResult, error) {
	empty := models.AdminUserAccessResult{}
	if actorID <= 0 || targetID <= 0 {
		return empty, ErrUserAccessNotFound
	}
	if actorID == targetID {
		return empty, ErrUserAccessSelf
	}

	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return empty, err
	}
	defer tx.Rollback()

	var actorAdmin, actorBlocked bool
	err = tx.QueryRowContext(ctx, `SELECT is_admin, is_blocked FROM users WHERE id=$1 FOR UPDATE`, actorID).Scan(&actorAdmin, &actorBlocked)
	if errors.Is(err, sql.ErrNoRows) {
		return empty, ErrUserAccessUnauthorized
	}
	if err != nil {
		return empty, err
	}
	if !actorAdmin || actorBlocked {
		return empty, ErrUserAccessUnauthorized
	}

	var targetAdmin, targetBlocked bool
	err = tx.QueryRowContext(ctx, `SELECT is_admin, is_blocked FROM users WHERE id=$1 FOR UPDATE`, targetID).Scan(&targetAdmin, &targetBlocked)
	if errors.Is(err, sql.ErrNoRows) {
		return empty, ErrUserAccessNotFound
	}
	if err != nil {
		return empty, classifyUserAccessError(err)
	}
	if targetBlocked == blocked {
		return empty, ErrUserAccessUnchanged
	}

	if targetAdmin && blocked {
		var admins int
		err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE is_admin=true AND is_blocked=false`).Scan(&admins)
		if err != nil {
			return empty, err
		}
		if admins <= 1 {
			return empty, ErrUserAccessLastAdmin
		}
	}

	_, err = tx.ExecContext(ctx, `UPDATE users SET is_blocked=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$1`, targetID, blocked)
	if err != nil {
		return empty, classifyUserAccessError(err)
	}

	var result models.AdminUserAccessResult
	err = tx.QueryRowContext(ctx, `SELECT id,email,full_name,COALESCE(rut,''),is_admin,is_blocked,
        COALESCE(entra_oid,''),COALESCE(tenant_id,'') FROM users WHERE id=$1`, targetID).Scan(
		&result.User.ID, &result.User.Email, &result.User.FullName, &result.User.RUT,
		&result.User.IsAdmin, &result.User.IsBlocked, &result.User.OID, &result.User.TenantID)
	if err != nil {
		return empty, classifyUserAccessError(err)
	}

	err = tx.QueryRowContext(ctx, `
        INSERT INTO user_access_audit
            (target_user_id, actor_user_id, previous_is_blocked, new_is_blocked, reason)
        VALUES ($1,$2,$3,$4,$5)
        RETURNING id, target_user_id, actor_user_id, previous_is_blocked,
                  new_is_blocked, reason, changed_at
    `, targetID, actorID, targetBlocked, blocked, reason).Scan(
		&result.Audit.ID, &result.Audit.TargetUserID, &result.Audit.ActorUserID,
		&result.Audit.PreviousIsBlocked, &result.Audit.NewIsBlocked,
		&result.Audit.Reason, &result.Audit.ChangedAt)
	if err != nil {
		return empty, classifyUserAccessError(err)
	}

	if err = tx.Commit(); err != nil {
		return empty, classifyUserAccessError(err)
	}
	return result, nil
}

func classifyUserAccessError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && (pgErr.Code == "40001" || pgErr.Code == "40P01") {
		return ErrUserAccessConflict
	}
	return err
}

func GetUserAccessAudit(targetID int) ([]models.UserAccessAudit, error) {
	ctx := context.Background()
	var exists bool
	if err := database.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, targetID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserAccessNotFound
	}

	rows, err := database.DB.QueryContext(ctx, `
        SELECT audit.id,audit.target_user_id,audit.actor_user_id,actor.full_name,
            audit.previous_is_blocked,audit.new_is_blocked,audit.reason,audit.changed_at
        FROM user_access_audit audit
        JOIN users actor ON actor.id=audit.actor_user_id
        WHERE audit.target_user_id=$1
        ORDER BY audit.changed_at DESC,audit.id DESC LIMIT 100`, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []models.UserAccessAudit{}
	for rows.Next() {
		var value models.UserAccessAudit
		if err := rows.Scan(&value.ID, &value.TargetUserID, &value.ActorUserID,
			&value.ActorName, &value.PreviousIsBlocked, &value.NewIsBlocked,
			&value.Reason, &value.ChangedAt); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}
