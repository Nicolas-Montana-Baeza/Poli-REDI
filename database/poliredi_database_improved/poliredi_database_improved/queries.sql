-- ============================================================================
-- POLI-REDI | Consultas de operación y diagnóstico
-- Por defecto no modifica datos. Cambie @apply_changes a 1 solo conscientemente.
-- ============================================================================

SET NOCOUNT ON;
GO

DECLARE @admin_email NVARCHAR(150) = N'admin@polirediucen.onmicrosoft.com';
DECLARE @apply_changes BIT = 0;

SELECT id, email, full_name, is_admin, is_blocked, updated_at
FROM dbo.users
WHERE email = @admin_email;

IF @apply_changes = 1
BEGIN
    UPDATE dbo.users
    SET is_admin = 1,
        updated_at = SYSUTCDATETIME()
    WHERE email = @admin_email
      AND is_admin = 0;

    SELECT @@ROWCOUNT AS users_promoted;
END;
GO

-- Reservas activas próximas, con capacidad y confirmaciones.
SELECT
    r.id,
    r.start_time,
    DATEADD(MINUTE, r.duration_minutes, r.start_time) AS end_time,
    r.status,
    u.full_name AS owner_name,
    res.name AS resource_name,
    a.name AS activity_name,
    r.group_capacity_snapshot,
    COALESCE(r.target_participants, r.group_capacity_snapshot) AS effective_target,
    SUM(CASE WHEN p.status = N'CONFIRMED' THEN 1 ELSE 0 END) AS confirmed_participants
FROM dbo.reservations AS r
INNER JOIN dbo.users AS u ON u.id = r.user_id
LEFT JOIN dbo.resources AS res ON res.id = r.resource_id
LEFT JOIN dbo.activities AS a ON a.id = r.activity_id
LEFT JOIN dbo.participants AS p ON p.reservation_id = r.id
WHERE r.status IN (N'PENDING', N'CONFIRMED')
GROUP BY r.id,r.start_time,r.duration_minutes,r.status,u.full_name,res.name,a.name,r.group_capacity_snapshot,r.target_participants
ORDER BY r.start_time, r.id;
GO

-- Usuarios con privilegios o restricciones.
SELECT id,email,full_name,rut,is_admin,is_blocked,updated_at
FROM dbo.users
WHERE is_admin=1 OR is_blocked=1
ORDER BY is_admin DESC,is_blocked DESC,full_name;
GO

-- Estado de políticas y su alcance.
SELECT p.id,p.effective_from,p.effective_to,p.is_published,p.idempotency_key,
       COUNT(DISTINCT pr.resource_id) AS allowed_resources,
       COUNT(DISTINCT gr.resource_id) AS group_resources,
       COUNT(DISTINCT d.duration_minutes) AS allowed_durations
FROM dbo.reservation_policies AS p
LEFT JOIN dbo.reservation_policy_resources AS pr ON pr.policy_id=p.id
LEFT JOIN dbo.reservation_policy_group_resources AS gr ON gr.policy_id=p.id
LEFT JOIN dbo.reservation_policy_durations AS d ON d.policy_id=p.id
GROUP BY p.id,p.effective_from,p.effective_to,p.is_published,p.idempotency_key
ORDER BY p.effective_from,p.id;
GO
