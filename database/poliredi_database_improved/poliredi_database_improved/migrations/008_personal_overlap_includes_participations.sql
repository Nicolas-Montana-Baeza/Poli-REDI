-- ============================================================================
-- POLI-REDI | Azure SQL Database / SQL Server
-- Archivo: 008_personal_overlap_includes_participations.sql
-- Propósito: Incluir participaciones confirmadas en los solapes personales.
-- Reejecución: Idempotente; revisar PRECHECK y POSTCHECK.
-- Requiere cliente con soporte para separadores GO.
-- ============================================================================

SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
SET NOCOUNT ON;
SET XACT_ABORT ON;
GO

/*
La agenda personal se compone de reservas propias y participaciones CONFIRMED.
Los extremos contiguos son validos: solo se rechazan intervalos con solape real.
Los triggers se ejecutan dentro de la transaccion que origina cada escritura.
*/

IF OBJECT_ID('dbo.reservations', 'U') IS NULL
   OR OBJECT_ID('dbo.participants', 'U') IS NULL
    THROW 58000, 'Preflight: faltan reservas o participantes.', 1;
GO

CREATE OR ALTER TRIGGER dbo.trg_reservations_validate_participant_overlap
ON dbo.reservations
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        INNER JOIN dbo.reservations existing WITH (UPDLOCK, HOLDLOCK)
            ON existing.id <> i.id
           AND existing.status IN ('PENDING', 'CONFIRMED')
           AND i.start_time < DATEADD(MINUTE, existing.duration_minutes, existing.start_time)
           AND DATEADD(MINUTE, i.duration_minutes, i.start_time) > existing.start_time
        WHERE i.status IN ('PENDING', 'CONFIRMED')
          AND EXISTS (
              SELECT 1
              FROM dbo.participants membership WITH (UPDLOCK, HOLDLOCK)
              WHERE membership.reservation_id = existing.id
                AND membership.user_id = i.user_id
                AND membership.status = 'CONFIRMED'
          )
    )
        THROW 51023, 'El usuario ya participa en una reserva activa en ese horario.', 1;
END;
GO

CREATE OR ALTER TRIGGER dbo.trg_participants_validate_personal_overlap
ON dbo.participants
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    IF EXISTS (
        SELECT 1
        FROM inserted membership
        INNER JOIN dbo.reservations joined_reservation
            ON joined_reservation.id = membership.reservation_id
           AND joined_reservation.status IN ('PENDING', 'CONFIRMED')
        INNER JOIN dbo.reservations existing WITH (UPDLOCK, HOLDLOCK)
            ON existing.id <> joined_reservation.id
           AND existing.status IN ('PENDING', 'CONFIRMED')
           AND joined_reservation.start_time < DATEADD(MINUTE, existing.duration_minutes, existing.start_time)
           AND DATEADD(MINUTE, joined_reservation.duration_minutes, joined_reservation.start_time) > existing.start_time
        WHERE membership.status = 'CONFIRMED'
          AND (
              existing.user_id = membership.user_id
              OR EXISTS (
                  SELECT 1
                  FROM dbo.participants other_membership WITH (UPDLOCK, HOLDLOCK)
                  WHERE other_membership.reservation_id = existing.id
                    AND other_membership.user_id = membership.user_id
                    AND other_membership.status = 'CONFIRMED'
              )
          )
    )
        THROW 51023, 'El usuario ya tiene una reserva o participacion activa en ese horario.', 1;
END;
GO

IF OBJECT_DEFINITION(OBJECT_ID('dbo.trg_reservations_validate_participant_overlap'))
       NOT LIKE N'%membership.status = ''CONFIRMED''%'
   OR OBJECT_DEFINITION(OBJECT_ID('dbo.trg_reservations_validate_participant_overlap'))
       NOT LIKE N'%i.start_time < DATEADD(MINUTE, existing.duration_minutes, existing.start_time)%'
   OR OBJECT_DEFINITION(OBJECT_ID('dbo.trg_participants_validate_personal_overlap'))
       NOT LIKE N'%existing.user_id = membership.user_id%'
   OR OBJECT_DEFINITION(OBJECT_ID('dbo.trg_participants_validate_personal_overlap'))
       NOT LIKE N'%other_membership.status = ''CONFIRMED''%'
    THROW 58001, 'Postcheck: los guards de agenda personal no quedaron instalados.', 1;
GO

SELECT
    CONVERT(bit, 1) AS personal_overlap_includes_confirmed_participations,
    CONVERT(bit, CASE WHEN OBJECT_DEFINITION(OBJECT_ID('dbo.trg_participants_validate_personal_overlap'))
        LIKE N'%joined_reservation.start_time < DATEADD(MINUTE, existing.duration_minutes, existing.start_time)%'
        AND OBJECT_DEFINITION(OBJECT_ID('dbo.trg_participants_validate_personal_overlap'))
        LIKE N'%DATEADD(MINUTE, joined_reservation.duration_minutes, joined_reservation.start_time) > existing.start_time%'
        THEN 1 ELSE 0 END) AS contiguous_intervals_allowed;
GO
