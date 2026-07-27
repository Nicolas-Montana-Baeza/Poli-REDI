-- ============================================================
-- POLI-REDI - QUERIES DE MANTENCIÓN
-- Consultas útiles para inspección, diagnóstico y ajustes.
-- ============================================================

-- 1. Promover un usuario a administrador de forma idempotente.
UPDATE dbo.users
SET is_admin = 1,
    updated_at = SYSUTCDATETIME()
WHERE email = N'admin@polirediucen.onmicrosoft.com';
GO

-- 2. Revisar reservas pendientes con su recurso y actividad.
SELECT r.id,
       r.start_time,
       r.status,
       u.full_name AS user_name,
       res.name AS resource_name,
       a.name AS activity_name
FROM dbo.reservations AS r
INNER JOIN dbo.users AS u ON u.id = r.user_id
LEFT JOIN dbo.resources AS res ON res.id = r.resource_id
LEFT JOIN dbo.activities AS a ON a.id = r.activity_id
WHERE r.status = 'PENDING'
ORDER BY r.start_time;
GO

-- 3. Listar usuarios bloqueados o administradores.
SELECT id,
       email,
       full_name,
       is_admin,
       is_blocked
FROM dbo.users
WHERE is_admin = 1 OR is_blocked = 1
ORDER BY is_admin DESC, is_blocked DESC, full_name;
GO
