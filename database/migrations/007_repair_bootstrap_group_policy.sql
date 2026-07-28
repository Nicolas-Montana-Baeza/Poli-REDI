SET XACT_ABORT ON;
SET NOCOUNT ON;
GO

IF OBJECT_ID('dbo.reservation_policies','U') IS NULL
   OR OBJECT_ID('dbo.reservation_policy_resources','U') IS NULL
   OR OBJECT_ID('dbo.reservation_policy_group_resources','U') IS NULL
   OR OBJECT_ID('dbo.reservation_policy_durations','U') IS NULL
   OR OBJECT_ID('dbo.resources','U') IS NULL
    THROW 57000,'Preflight: faltan objetos de politica o recursos.',1;
GO

IF EXISTS(
    SELECT expected.resource_id
    FROM (VALUES(1),(2),(7)) expected(resource_id)
    LEFT JOIN dbo.resources r ON r.id=expected.resource_id
    WHERE r.id IS NULL OR r.is_active=0 OR r.reservation_mode='OPEN_USE'
       OR r.capacity IS NULL OR r.capacity<10
)
    THROW 57001,'Preflight: Cancha 1, 2 o 3 no admite politica grupal.',1;
GO

BEGIN TRY
    SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
    BEGIN TRANSACTION;

    DECLARE @key NVARCHAR(100)=N'repair-bootstrap-group-policy-v1';
    DECLARE @now DATETIME2(0)=SYSUTCDATETIME();
    DECLARE @current INT;
    DECLARE @new INT;

    SELECT @new=id
    FROM dbo.reservation_policies WITH(UPDLOCK,HOLDLOCK)
    WHERE idempotency_key=@key;

    IF @new IS NULL
    BEGIN
        SELECT TOP(1) @current=id
        FROM dbo.reservation_policies WITH(UPDLOCK,HOLDLOCK)
        WHERE is_published=1
          AND effective_from<=@now
          AND (effective_to IS NULL OR effective_to>@now)
        ORDER BY effective_from DESC,id DESC;

        IF @current IS NULL
            THROW 57002,'No existe una politica vigente publicada.',1;

        IF NOT EXISTS(
            SELECT 1
            FROM dbo.reservation_policies
            WHERE id=@current AND idempotency_key IS NULL
        ) OR EXISTS(
            SELECT 1
            FROM dbo.reservation_policy_group_resources
            WHERE policy_id=@current
        )
            THROW 57003,'La politica vigente no corresponde al bootstrap vacio; no se modifico.',1;

        UPDATE dbo.reservation_policies
        SET effective_to=@now
        WHERE id=@current;

        INSERT INTO dbo.reservation_policies(
            reservable_window_days,request_frequency_days,
            confirmation_deadline_minutes,minimum_participants,
            opening_minute,closing_minute,slot_interval_minutes,
            effective_from,created_by_user_id,idempotency_key,
            idempotency_payload_hash,is_published
        )
        SELECT reservable_window_days,request_frequency_days,
            confirmation_deadline_minutes,minimum_participants,
            opening_minute,closing_minute,slot_interval_minutes,
            @now,NULL,@key,
            '7777777777777777777777777777777777777777777777777777777777777777',0
        FROM dbo.reservation_policies
        WHERE id=@current;

        SET @new=SCOPE_IDENTITY();

        INSERT INTO dbo.reservation_policy_durations(policy_id,duration_minutes)
        SELECT @new,duration_minutes
        FROM dbo.reservation_policy_durations
        WHERE policy_id=@current;

        INSERT INTO dbo.reservation_policy_resources(policy_id,resource_id)
        SELECT @new,resource_id
        FROM dbo.reservation_policy_resources
        WHERE policy_id=@current;

        INSERT INTO dbo.reservation_policy_group_resources(policy_id,resource_id)
        SELECT @new,resource_id
        FROM (VALUES(1),(2),(7)) expected(resource_id);

        UPDATE dbo.reservation_policies
        SET is_published=1
        WHERE id=@new;
    END;

    IF NOT EXISTS(
        SELECT 1
        FROM dbo.reservation_policies
        WHERE id=@new AND is_published=1 AND effective_to IS NULL
    ) OR (
        SELECT COUNT(*)
        FROM dbo.reservation_policy_group_resources
        WHERE policy_id=@new AND resource_id IN(1,2,7)
    )<>3
        THROW 57004,'Postcheck transaccional: la politica reparada no quedo vigente.',1;

    COMMIT TRANSACTION;
END TRY
BEGIN CATCH
    IF XACT_STATE()<>0 ROLLBACK TRANSACTION;
    THROW;
END CATCH;
GO

SELECT
    CONVERT(bit,CASE WHEN EXISTS(
        SELECT 1
        FROM dbo.reservation_policies p
        WHERE p.idempotency_key='repair-bootstrap-group-policy-v1'
          AND p.is_published=1 AND p.effective_to IS NULL
          AND (SELECT COUNT(*) FROM dbo.reservation_policy_group_resources g
               WHERE g.policy_id=p.id AND g.resource_id IN(1,2,7))=3
    ) THEN 1 ELSE 0 END) AS bootstrap_group_policy_repaired;
GO
