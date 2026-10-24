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
	ErrUserRoleUnauthorized  = errors.New("solo administradores activos pueden asignar roles")
	ErrUserRoleNotFound      = errors.New("usuario no encontrado")
	ErrUserRoleSelf          = errors.New("no puedes cambiar tu propio rol administrativo")
	ErrUserRoleUnchanged     = errors.New("el usuario ya tiene ese rol")
	ErrUserRoleLastAdmin     = errors.New("no puedes quitar el rol al ultimo administrador activo")
	ErrUserRoleBlockedTarget = errors.New("desbloquea la cuenta antes de otorgar el rol administrador")
	ErrUserRoleConflict      = errors.New("cambio concurrente; actualiza la lista y vuelve a intentarlo")
)

func classifyUserRoleError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && (pgErr.Code == "40001" || pgErr.Code == "40P01") {
		return ErrUserRoleConflict
	}
	return err
}

// La asignacion y su auditoria son atomicas. La serializacion impide que dos
// administradores se retiren los ultimos permisos en transacciones concurrentes.
func ChangeAdminUserRole(actorID, targetID int, grantAdmin bool, reason string) (models.AdminUserRoleResult, error) {
	empty := models.AdminUserRoleResult{}
	if actorID <= 0 || targetID <= 0 {
		return empty, ErrUserRoleNotFound
	}
	if actorID == targetID {
		return empty, ErrUserRoleSelf
	}

	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return empty, err
	}
	defer tx.Rollback()

	var actorAdmin, actorBlocked bool
	var actorName string
	err = tx.QueryRowContext(ctx, `SELECT is_admin,is_blocked,full_name FROM users WHERE id=$1 FOR UPDATE`, actorID).
		Scan(&actorAdmin, &actorBlocked, &actorName)
	if errors.Is(err, sql.ErrNoRows) {
		return empty, ErrUserRoleUnauthorized
	}
	if err != nil {
		return empty, classifyUserRoleError(err)
	}
	if !actorAdmin || actorBlocked {
		return empty, ErrUserRoleUnauthorized
	}

	var targetAdmin, targetBlocked bool
	err = tx.QueryRowContext(ctx, `SELECT is_admin,is_blocked FROM users WHERE id=$1 FOR UPDATE`, targetID).
		Scan(&targetAdmin, &targetBlocked)
	if errors.Is(err, sql.ErrNoRows) {
		return empty, ErrUserRoleNotFound
	}
	if err != nil {
		return empty, classifyUserRoleError(err)
	}
	if targetAdmin == grantAdmin {
		return empty, ErrUserRoleUnchanged
	}
	if grantAdmin && targetBlocked {
		return empty, ErrUserRoleBlockedTarget
	}

	if targetAdmin && !grantAdmin && !targetBlocked {
		var activeAdmins int
		err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE is_admin=true AND is_blocked=false`).Scan(&activeAdmins)
		if err != nil {
			return empty, classifyUserRoleError(err)
		}
		if activeAdmins <= 1 {
			return empty, ErrUserRoleLastAdmin
		}
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE users SET is_admin=$2, updated_at=CURRENT_TIMESTAMP WHERE id=$1`, targetID, grantAdmin)
	if err != nil {
		return empty, classifyUserRoleError(err)
	}

	var result models.AdminUserRoleResult
	err = tx.QueryRowContext(ctx, `SELECT id,email,full_name,COALESCE(rut,''),is_admin,is_blocked,
        COALESCE(entra_oid,''),COALESCE(tenant_id,'') FROM users WHERE id=$1`, targetID).Scan(
		&result.User.ID, &result.User.Email, &result.User.FullName, &result.User.RUT,
		&result.User.IsAdmin, &result.User.IsBlocked, &result.User.OID, &result.User.TenantID)
	if err != nil {
		return empty, classifyUserRoleError(err)
	}

	err = tx.QueryRowContext(ctx, `INSERT INTO user_role_audit
        (target_user_id,actor_user_id,previous_is_admin,new_is_admin,reason)
        VALUES ($1,$2,$3,$4,$5)
        RETURNING id,target_user_id,actor_user_id,previous_is_admin,new_is_admin,reason,changed_at`,
		targetID, actorID, targetAdmin, grantAdmin, reason).Scan(
		&result.Audit.ID, &result.Audit.TargetUserID, &result.Audit.ActorUserID,
		&result.Audit.PreviousIsAdmin, &result.Audit.NewIsAdmin,
		&result.Audit.Reason, &result.Audit.ChangedAt)
	if err != nil {
		return empty, classifyUserRoleError(err)
	}
	result.Audit.ActorName = actorName

	if err := tx.Commit(); err != nil {
		return empty, classifyUserRoleError(err)
	}
	return result, nil
}

func GetUserRoleAudit(targetID int) ([]models.UserRoleAudit, error) {
	ctx := context.Background()
	var exists bool
	if err := database.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, targetID).
		Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserRoleNotFound
	}
	rows, err := database.DB.QueryContext(ctx, `
        SELECT a.id,a.target_user_id,a.actor_user_id,u.full_name,
               a.previous_is_admin,a.new_is_admin,a.reason,a.changed_at
        FROM user_role_audit a JOIN users u ON u.id=a.actor_user_id
        WHERE a.target_user_id=$1
        ORDER BY a.changed_at DESC,a.id DESC LIMIT 100`, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []models.UserRoleAudit{}
	for rows.Next() {
		var item models.UserRoleAudit
		if err := rows.Scan(&item.ID, &item.TargetUserID, &item.ActorUserID,
			&item.ActorName, &item.PreviousIsAdmin, &item.NewIsAdmin,
			&item.Reason, &item.ChangedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
