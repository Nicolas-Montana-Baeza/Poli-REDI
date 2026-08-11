-- ============================================================================
-- POLI-REDI | Verificación posterior a despliegue (solo lectura)
-- Todos los indicadores *_ok deben devolver 1.
-- ============================================================================
SET NOCOUNT ON;
SET TRANSACTION ISOLATION LEVEL READ COMMITTED;
GO

SELECT
    CONVERT(bit,CASE WHEN (SELECT COUNT(*) FROM dbo.reservation_policies WHERE effective_to IS NULL AND is_published=1)=1 THEN 1 ELSE 0 END) AS one_current_published_policy_ok,
    CONVERT(bit,CASE WHEN NOT EXISTS(SELECT idempotency_key FROM dbo.reservation_policies WHERE idempotency_key IS NOT NULL GROUP BY idempotency_key HAVING COUNT(*)>1) THEN 1 ELSE 0 END) AS policy_idempotency_unique_ok,
    CONVERT(bit,CASE WHEN NOT EXISTS(SELECT reservation_id,user_id FROM dbo.participants GROUP BY reservation_id,user_id HAVING COUNT(*)>1) THEN 1 ELSE 0 END) AS participants_unique_ok,
    CONVERT(bit,CASE WHEN OBJECT_ID(N'dbo.trg_users_rut_validate',N'TR') IS NOT NULL AND OBJECT_ID(N'dbo.trg_users_rut_write_once',N'TR') IS NOT NULL THEN 1 ELSE 0 END) AS rut_triggers_ok,
    CONVERT(bit,CASE WHEN OBJECT_ID(N'dbo.trg_reservations_validate_conflicts',N'TR') IS NOT NULL THEN 1 ELSE 0 END) AS reservation_conflict_trigger_ok,
    CONVERT(bit,CASE WHEN OBJECT_ID(N'dbo.trg_reservations_group_snapshot_validate',N'TR') IS NOT NULL THEN 1 ELSE 0 END) AS group_snapshot_trigger_ok,
    CONVERT(bit,CASE WHEN OBJECT_ID(N'dbo.trg_reservation_participant_audit_append_only',N'TR') IS NOT NULL AND OBJECT_ID(N'dbo.trg_reservation_target_audit_append_only',N'TR') IS NOT NULL THEN 1 ELSE 0 END) AS audit_append_only_triggers_ok,
    CONVERT(bit,CASE WHEN OBJECT_ID(N'dbo.trg_reservations_validate_participant_overlap',N'TR') IS NOT NULL AND OBJECT_ID(N'dbo.trg_participants_validate_personal_overlap',N'TR') IS NOT NULL THEN 1 ELSE 0 END) AS personal_overlap_triggers_ok,
    CONVERT(bit,CASE WHEN OBJECT_DEFINITION(OBJECT_ID(N'dbo.trg_reservations_audit')) LIKE N'%skip_reservation_audit%' THEN 1 ELSE 0 END) AS reservation_audit_single_write_ok,
    CONVERT(bit,CASE WHEN OBJECT_ID(N'dbo.trg_reservations_pending_conflicts',N'TR') IS NULL AND OBJECT_ID(N'dbo.trg_blocks_pending_conflicts',N'TR') IS NULL AND OBJECT_ID(N'dbo.trg_scheduled_activities_pending_conflicts',N'TR') IS NULL THEN 1 ELSE 0 END) AS redundant_triggers_absent_ok;
GO

SELECT
    CONVERT(bit,CASE WHEN NOT EXISTS(
        SELECT 1 FROM dbo.users WHERE rut IS NOT NULL AND (
            LEN(rut) NOT BETWEEN 9 AND 10 OR LEN(rut)-LEN(REPLACE(rut,N'-',N''))<>1
            OR CHARINDEX(N'-',rut)<>LEN(rut)-1 OR LEFT(rut,LEN(rut)-2) LIKE N'%[^0-9]%'
            OR RIGHT(rut,1) NOT LIKE N'[0-9K]' OR rut<>UPPER(rut) OR rut LIKE N'%[ .]%'
        )
    ) THEN 1 ELSE 0 END) AS rut_format_data_ok,
    CONVERT(bit,CASE WHEN NOT EXISTS(SELECT rut FROM dbo.users WHERE rut IS NOT NULL GROUP BY rut HAVING COUNT(*)>1) THEN 1 ELSE 0 END) AS rut_unique_data_ok,
    CONVERT(bit,CASE WHEN NOT EXISTS(
        SELECT 1
        FROM dbo.users AS u
        CROSS APPLY(SELECT LEFT(u.rut,LEN(u.rut)-2),RIGHT(u.rut,1))parts(body,verifier)
        CROSS APPLY(
            SELECT SUM(CONVERT(INT,SUBSTRING(parts.body,LEN(parts.body)-n.n+1,1))*(2+((n.n-1)%6))) weighted_sum
            FROM(VALUES(1),(2),(3),(4),(5),(6),(7),(8))n(n) WHERE n.n<=LEN(parts.body)
        )calc
        CROSS APPLY(SELECT CASE 11-(calc.weighted_sum%11) WHEN 11 THEN N'0' WHEN 10 THEN N'K' ELSE CONVERT(NVARCHAR(1),11-(calc.weighted_sum%11)) END)validation(expected_verifier)
        WHERE u.rut IS NOT NULL AND validation.expected_verifier<>parts.verifier
    ) THEN 1 ELSE 0 END) AS rut_check_digit_data_ok,
    CONVERT(bit,CASE WHEN NOT EXISTS(
        SELECT 1 FROM dbo.reservation_policies AS p
        CROSS JOIN(VALUES(30),(60),(90),(120),(150),(180))d(duration_minutes)
        WHERE NOT EXISTS(SELECT 1 FROM dbo.reservation_policy_durations x WHERE x.policy_id=p.id AND x.duration_minutes=d.duration_minutes)
    ) THEN 1 ELSE 0 END) AS policy_durations_complete_ok,
    CONVERT(bit,CASE WHEN NOT EXISTS(
        SELECT 1 FROM sys.columns WHERE object_id=OBJECT_ID(N'dbo.reservation_target_audit')
        AND name IN(N'old_target_participants',N'new_target_participants') AND is_nullable=0
    ) THEN 1 ELSE 0 END) AS target_audit_nullable_ok;
GO

SELECT name,is_disabled,is_not_trusted
FROM sys.foreign_keys
WHERE is_disabled=1 OR is_not_trusted=1
ORDER BY name;
GO

SELECT OBJECT_NAME(object_id) AS table_name,name AS index_name,is_disabled,has_filter
FROM sys.indexes
WHERE name IN(
 N'idx_reservations_resource_status_start',N'idx_reservations_user_status_start',
 N'idx_participants_user_status_reservation',N'idx_availability_blocks_resource_active_start',
 N'idx_scheduled_activities_resource_active_start',N'idx_workshop_enrollments_user_status',N'uq_workshop_enrollments_confirmed'
)
ORDER BY table_name,index_name;
GO
