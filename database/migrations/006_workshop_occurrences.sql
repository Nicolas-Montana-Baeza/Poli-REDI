SET XACT_ABORT ON;
GO

IF OBJECT_ID('dbo.workshops','U') IS NULL OR OBJECT_ID('dbo.workshop_enrollments','U') IS NULL
    THROW 56000, 'Faltan tablas prerequisite workshops/workshop_enrollments.', 1;
IF COL_LENGTH('dbo.workshops','id') IS NULL OR COL_LENGTH('dbo.workshops','is_active') IS NULL
   OR COL_LENGTH('dbo.workshops','capacity') IS NULL
   OR COL_LENGTH('dbo.workshop_enrollments','id') IS NULL
   OR COL_LENGTH('dbo.workshop_enrollments','workshop_id') IS NULL
   OR COL_LENGTH('dbo.workshop_enrollments','user_id') IS NULL
   OR COL_LENGTH('dbo.workshop_enrollments','status') IS NULL
    THROW 56001, 'Las tablas prerequisite tienen estructura incompatible.', 1;
GO

IF OBJECT_ID('dbo.workshop_occurrences','U') IS NULL
BEGIN
    CREATE TABLE dbo.workshop_occurrences (
        id INT IDENTITY(1,1) NOT NULL CONSTRAINT pk_workshop_occurrences PRIMARY KEY,
        workshop_id INT NOT NULL,
        weekday_iso TINYINT NOT NULL,
        start_minute SMALLINT NOT NULL,
        end_minute SMALLINT NOT NULL
    );
END;
GO

IF EXISTS (
    SELECT required.name
    FROM (VALUES ('id','int',4,0),('workshop_id','int',4,0),
                 ('weekday_iso','tinyint',1,0),('start_minute','smallint',2,0),
                 ('end_minute','smallint',2,0)) required(name,type_name,max_length,is_nullable)
    LEFT JOIN sys.columns c ON c.object_id=OBJECT_ID('dbo.workshop_occurrences') AND c.name=required.name
    LEFT JOIN sys.types t ON t.user_type_id=c.user_type_id
    WHERE c.column_id IS NULL OR t.name<>required.type_name
       OR c.max_length<>required.max_length OR c.is_nullable<>required.is_nullable
)
    THROW 56002, 'workshop_occurrences tiene columnas incompatibles.', 1;
IF COLUMNPROPERTY(OBJECT_ID('dbo.workshop_occurrences'),'id','IsIdentity')<>1
    THROW 56003, 'workshop_occurrences.id debe ser IDENTITY.', 1;
IF EXISTS (
    SELECT 1 FROM sys.key_constraints
    WHERE parent_object_id=OBJECT_ID('dbo.workshop_occurrences') AND type='PK'
      AND name<>'pk_workshop_occurrences'
)
    THROW 56004, 'workshop_occurrences tiene una PK divergente.', 1;
GO

IF NOT EXISTS (SELECT 1 FROM sys.key_constraints WHERE parent_object_id=OBJECT_ID('dbo.workshop_occurrences') AND type='PK')
    ALTER TABLE dbo.workshop_occurrences ADD CONSTRAINT pk_workshop_occurrences PRIMARY KEY(id);
GO
IF NOT EXISTS (SELECT 1 FROM sys.foreign_keys WHERE parent_object_id=OBJECT_ID('dbo.workshop_occurrences') AND name='fk_workshop_occurrences_workshop')
    ALTER TABLE dbo.workshop_occurrences ADD CONSTRAINT fk_workshop_occurrences_workshop
      FOREIGN KEY(workshop_id) REFERENCES dbo.workshops(id) ON DELETE CASCADE;
GO
IF NOT EXISTS (SELECT 1 FROM sys.check_constraints WHERE parent_object_id=OBJECT_ID('dbo.workshop_occurrences') AND name='ck_workshop_occurrences_weekday')
    ALTER TABLE dbo.workshop_occurrences ADD CONSTRAINT ck_workshop_occurrences_weekday CHECK(weekday_iso BETWEEN 1 AND 7);
GO
IF NOT EXISTS (SELECT 1 FROM sys.check_constraints WHERE parent_object_id=OBJECT_ID('dbo.workshop_occurrences') AND name='ck_workshop_occurrences_minutes')
    ALTER TABLE dbo.workshop_occurrences ADD CONSTRAINT ck_workshop_occurrences_minutes CHECK(start_minute>=0 AND start_minute<1440 AND end_minute>0 AND end_minute<=1440 AND start_minute<end_minute);
GO
IF NOT EXISTS (SELECT 1 FROM sys.key_constraints WHERE parent_object_id=OBJECT_ID('dbo.workshop_occurrences') AND name='uq_workshop_occurrences_slot')
    ALTER TABLE dbo.workshop_occurrences ADD CONSTRAINT uq_workshop_occurrences_slot UNIQUE(workshop_id,weekday_iso,start_minute,end_minute);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE object_id=OBJECT_ID('dbo.workshop_occurrences') AND name='idx_workshop_occurrences_overlap')
    CREATE INDEX idx_workshop_occurrences_overlap ON dbo.workshop_occurrences(weekday_iso,start_minute,end_minute,workshop_id);
GO

IF NOT EXISTS (
    SELECT 1 FROM sys.key_constraints kc
    JOIN sys.index_columns ic ON ic.object_id=kc.parent_object_id AND ic.index_id=kc.unique_index_id
    JOIN sys.columns c ON c.object_id=ic.object_id AND c.column_id=ic.column_id
    WHERE kc.parent_object_id=OBJECT_ID('dbo.workshop_occurrences')
      AND kc.name='pk_workshop_occurrences' AND kc.type='PK'
    GROUP BY kc.name HAVING COUNT(*)=1 AND MAX(c.name)='id'
) THROW 56007, 'La PK de workshop_occurrences es incompatible.', 1;
IF NOT EXISTS (
    SELECT 1 FROM sys.foreign_keys fk
    JOIN sys.foreign_key_columns fkc ON fkc.constraint_object_id=fk.object_id
    JOIN sys.columns pc ON pc.object_id=fkc.parent_object_id AND pc.column_id=fkc.parent_column_id
    JOIN sys.columns rc ON rc.object_id=fkc.referenced_object_id AND rc.column_id=fkc.referenced_column_id
    WHERE fk.parent_object_id=OBJECT_ID('dbo.workshop_occurrences')
      AND fk.name='fk_workshop_occurrences_workshop'
      AND fk.referenced_object_id=OBJECT_ID('dbo.workshops')
      AND pc.name='workshop_id' AND rc.name='id'
      AND fk.delete_referential_action=1 AND fk.is_disabled=0 AND fk.is_not_trusted=0
) THROW 56008, 'La FK de workshop_occurrences es incompatible.', 1;
IF EXISTS (
    SELECT 1 FROM sys.check_constraints
    WHERE parent_object_id=OBJECT_ID('dbo.workshop_occurrences')
      AND name IN ('ck_workshop_occurrences_weekday','ck_workshop_occurrences_minutes')
      AND (is_disabled=1 OR is_not_trusted=1)
) THROW 56009, 'Los CHECK de workshop_occurrences no son confiables.', 1;
IF NOT EXISTS (
    SELECT 1 FROM sys.check_constraints
    WHERE parent_object_id=OBJECT_ID('dbo.workshop_occurrences')
      AND name='ck_workshop_occurrences_weekday'
      AND (
          definition LIKE '%weekday_iso%between%(1)%and%(7)%'
          OR (definition LIKE '%weekday_iso%>=(1)%' AND definition LIKE '%weekday_iso%<=(7)%')
      )
) OR NOT EXISTS (
    SELECT 1 FROM sys.check_constraints
    WHERE parent_object_id=OBJECT_ID('dbo.workshop_occurrences')
      AND name='ck_workshop_occurrences_minutes'
      AND definition LIKE '%start_minute%>=(0)%'
      AND definition LIKE '%end_minute%<=(1440)%'
      AND definition LIKE '%start_minute%<%end_minute%'
) THROW 56009, 'Las definiciones CHECK de workshop_occurrences son incompatibles.', 1;
IF NOT EXISTS (
    SELECT 1 FROM sys.key_constraints
    WHERE parent_object_id=OBJECT_ID('dbo.workshop_occurrences')
      AND name='uq_workshop_occurrences_slot' AND type='UQ'
) THROW 56010, 'La unicidad de occurrence es incompatible.', 1;
IF (
    SELECT COUNT(*) FROM sys.key_constraints kc
    JOIN sys.index_columns ic ON ic.object_id=kc.parent_object_id AND ic.index_id=kc.unique_index_id
    WHERE kc.parent_object_id=OBJECT_ID('dbo.workshop_occurrences')
      AND kc.name='uq_workshop_occurrences_slot' AND ic.is_included_column=0
)<>4 OR EXISTS (
    SELECT 1 FROM sys.key_constraints kc
    JOIN sys.index_columns ic ON ic.object_id=kc.parent_object_id AND ic.index_id=kc.unique_index_id
    JOIN sys.columns c ON c.object_id=ic.object_id AND c.column_id=ic.column_id
    WHERE kc.parent_object_id=OBJECT_ID('dbo.workshop_occurrences')
      AND kc.name='uq_workshop_occurrences_slot'
      AND c.name<>CASE ic.key_ordinal WHEN 1 THEN 'workshop_id' WHEN 2 THEN 'weekday_iso' WHEN 3 THEN 'start_minute' WHEN 4 THEN 'end_minute' END
) THROW 56010, 'Las columnas UNIQUE de occurrence son incompatibles.', 1;
IF NOT EXISTS (
    SELECT 1 FROM sys.indexes WHERE object_id=OBJECT_ID('dbo.workshop_occurrences')
      AND name='idx_workshop_occurrences_overlap' AND is_disabled=0
) THROW 56011, 'El indice de solape es incompatible.', 1;
IF (
    SELECT COUNT(*) FROM sys.indexes ix JOIN sys.index_columns ic
      ON ic.object_id=ix.object_id AND ic.index_id=ix.index_id
    WHERE ix.object_id=OBJECT_ID('dbo.workshop_occurrences')
      AND ix.name='idx_workshop_occurrences_overlap' AND ic.is_included_column=0
)<>4 OR EXISTS (
    SELECT 1 FROM sys.indexes ix JOIN sys.index_columns ic
      ON ic.object_id=ix.object_id AND ic.index_id=ix.index_id
    JOIN sys.columns c ON c.object_id=ic.object_id AND c.column_id=ic.column_id
    WHERE ix.object_id=OBJECT_ID('dbo.workshop_occurrences')
      AND ix.name='idx_workshop_occurrences_overlap'
      AND c.name<>CASE ic.key_ordinal WHEN 1 THEN 'weekday_iso' WHEN 2 THEN 'start_minute' WHEN 3 THEN 'end_minute' WHEN 4 THEN 'workshop_id' END
) THROW 56011, 'Las columnas del indice de solape son incompatibles.', 1;
GO

BEGIN TRANSACTION;
DECLARE @catalog TABLE(workshop_id INT,weekday_iso TINYINT,start_minute SMALLINT,end_minute SMALLINT,
 PRIMARY KEY(workshop_id,weekday_iso,start_minute,end_minute));
INSERT INTO @catalog VALUES
(1,2,945,1065),(1,4,945,1065),(2,2,930,975),(2,4,930,975),
(3,2,1020,1080),(3,4,1020,1080),(4,1,1020,1065),(4,3,1020,1065),
(5,1,1110,1170),(5,3,1110,1170),(6,2,720,900),(6,4,720,900),
(7,1,855,900),(7,2,855,900),(7,3,855,900),(7,4,855,900),
(8,1,930,975),(8,3,930,975),(9,1,990,1110),(9,5,1050,1140),
(10,3,840,930),(11,2,765,825),(11,3,765,825),(11,4,765,825),
(12,1,1170,1260),(12,3,1080,1170),(13,3,1200,1260),(14,3,1020,1140),
(15,3,1020,1080),(16,2,1140,1260),(16,6,705,780),(17,5,930,990);

IF EXISTS(SELECT 1 FROM dbo.workshops w WHERE w.is_active=1 AND NOT EXISTS(SELECT 1 FROM @catalog c WHERE c.workshop_id=w.id))
    THROW 56005, 'Existe taller activo fuera del catalogo explicito.', 1;
IF EXISTS(
 SELECT 1 FROM dbo.workshop_occurrences o JOIN dbo.workshops w ON w.id=o.workshop_id AND w.is_active=1
 WHERE NOT EXISTS(SELECT 1 FROM @catalog c WHERE c.workshop_id=o.workshop_id AND c.weekday_iso=o.weekday_iso AND c.start_minute=o.start_minute AND c.end_minute=o.end_minute)
)
    THROW 56006, 'Un taller activo tiene horarios divergentes del catalogo.', 1;

INSERT INTO dbo.workshop_occurrences(workshop_id,weekday_iso,start_minute,end_minute)
SELECT c.workshop_id,c.weekday_iso,c.start_minute,c.end_minute FROM @catalog c
JOIN dbo.workshops w ON w.id=c.workshop_id AND w.is_active=1
WHERE NOT EXISTS(SELECT 1 FROM dbo.workshop_occurrences o WHERE o.workshop_id=c.workshop_id AND o.weekday_iso=c.weekday_iso AND o.start_minute=c.start_minute AND o.end_minute=c.end_minute);
COMMIT TRANSACTION;
GO

CREATE OR ALTER TRIGGER dbo.trg_workshop_enrollments_validate
ON dbo.workshop_enrollments
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    -- El repositorio toma primero un lock de usuario y luego del taller. Ese es
    -- el camino soportado para serializar altas concurrentes sin deadlocks.
    IF EXISTS (
        SELECT 1
        FROM inserted i
        JOIN dbo.workshops w WITH (UPDLOCK, HOLDLOCK) ON w.id=i.workshop_id
        WHERE i.status='CONFIRMED'
          AND (
              w.is_active=0
              OR NOT EXISTS (
                  SELECT 1 FROM dbo.workshop_occurrences o WITH (HOLDLOCK)
                  WHERE o.workshop_id=i.workshop_id
                    AND o.weekday_iso BETWEEN 1 AND 7
                    AND o.start_minute>=0 AND o.end_minute<=1440
                    AND o.start_minute<o.end_minute
              )
          )
    )
        THROW 51300, 'El taller no esta activo o no tiene horario valido.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        JOIN dbo.workshops w WITH (UPDLOCK, HOLDLOCK) ON w.id=i.workshop_id
        CROSS APPLY (
            SELECT COUNT_BIG(*) AS confirmed_count
            FROM dbo.workshop_enrollments e WITH (UPDLOCK, HOLDLOCK)
            WHERE e.workshop_id=i.workshop_id AND e.status='CONFIRMED'
        ) counts
        WHERE i.status='CONFIRMED' AND counts.confirmed_count>w.capacity
    )
        THROW 51301, 'El taller no tiene cupos disponibles.', 1;

    IF EXISTS (
        SELECT 1
        FROM inserted i
        JOIN dbo.workshops target_w WITH (HOLDLOCK)
          ON target_w.id=i.workshop_id AND target_w.is_active=1
        JOIN dbo.workshop_occurrences target_o WITH (HOLDLOCK)
          ON target_o.workshop_id=i.workshop_id
        JOIN dbo.workshop_enrollments existing WITH (UPDLOCK, HOLDLOCK)
          ON existing.user_id=i.user_id AND existing.status='CONFIRMED'
         AND existing.id<>i.id AND existing.workshop_id<>i.workshop_id
        JOIN dbo.workshops existing_w WITH (HOLDLOCK)
          ON existing_w.id=existing.workshop_id AND existing_w.is_active=1
        JOIN dbo.workshop_occurrences existing_o WITH (HOLDLOCK)
          ON existing_o.workshop_id=existing.workshop_id
         AND existing_o.weekday_iso=target_o.weekday_iso
         AND existing_o.start_minute<target_o.end_minute
         AND target_o.start_minute<existing_o.end_minute
        WHERE i.status='CONFIRMED'
    )
        THROW 51300, 'El horario se superpone con otro taller confirmado.', 1;
END;
GO

WITH catalog(workshop_id,weekday_iso,start_minute,end_minute) AS (
 SELECT * FROM (VALUES
 (1,2,945,1065),(1,4,945,1065),(2,2,930,975),(2,4,930,975),
 (3,2,1020,1080),(3,4,1020,1080),(4,1,1020,1065),(4,3,1020,1065),
 (5,1,1110,1170),(5,3,1110,1170),(6,2,720,900),(6,4,720,900),
 (7,1,855,900),(7,2,855,900),(7,3,855,900),(7,4,855,900),
 (8,1,930,975),(8,3,930,975),(9,1,990,1110),(9,5,1050,1140),
 (10,3,840,930),(11,2,765,825),(11,3,765,825),(11,4,765,825),
 (12,1,1170,1260),(12,3,1080,1170),(13,3,1200,1260),(14,3,1020,1140),
 (15,3,1020,1080),(16,2,1140,1260),(16,6,705,780),(17,5,930,990)
 ) v(workshop_id,weekday_iso,start_minute,end_minute)
)
SELECT
 CASE WHEN NOT EXISTS(
   SELECT required.name FROM (VALUES('id'),('workshop_id'),('weekday_iso'),('start_minute'),('end_minute')) required(name)
   LEFT JOIN sys.columns c ON c.object_id=OBJECT_ID('dbo.workshop_occurrences') AND c.name=required.name
   WHERE c.column_id IS NULL) THEN 1 ELSE 0 END AS occurrence_columns_ok,
 CASE WHEN OBJECT_ID('dbo.trg_workshop_enrollments_validate','TR') IS NOT NULL THEN 1 ELSE 0 END AS enrollment_trigger_ok,
 CASE WHEN EXISTS(SELECT 1 FROM sys.indexes WHERE object_id=OBJECT_ID('dbo.workshop_occurrences') AND name='idx_workshop_occurrences_overlap') THEN 1 ELSE 0 END AS overlap_index_ok,
 CASE WHEN EXISTS(SELECT 1 FROM sys.foreign_keys WHERE parent_object_id=OBJECT_ID('dbo.workshop_occurrences') AND name='fk_workshop_occurrences_workshop' AND is_disabled=0 AND is_not_trusted=0) THEN 1 ELSE 0 END AS occurrence_fk_ok,
 CASE WHEN (SELECT COUNT(*) FROM sys.check_constraints WHERE parent_object_id=OBJECT_ID('dbo.workshop_occurrences') AND name IN('ck_workshop_occurrences_weekday','ck_workshop_occurrences_minutes') AND is_disabled=0 AND is_not_trusted=0)=2 THEN 1 ELSE 0 END AS occurrence_checks_ok,
 CASE WHEN EXISTS(SELECT 1 FROM sys.key_constraints WHERE parent_object_id=OBJECT_ID('dbo.workshop_occurrences') AND name='uq_workshop_occurrences_slot' AND type='UQ') THEN 1 ELSE 0 END AS occurrence_unique_ok,
 CASE WHEN NOT EXISTS(
   SELECT 1 FROM dbo.workshop_occurrences o JOIN dbo.workshops w ON w.id=o.workshop_id AND w.is_active=1
   WHERE NOT EXISTS(SELECT 1 FROM catalog c WHERE c.workshop_id=o.workshop_id AND c.weekday_iso=o.weekday_iso AND c.start_minute=o.start_minute AND c.end_minute=o.end_minute)
 ) AND NOT EXISTS(
   SELECT 1 FROM catalog c JOIN dbo.workshops w ON w.id=c.workshop_id AND w.is_active=1
   WHERE NOT EXISTS(SELECT 1 FROM dbo.workshop_occurrences o WHERE o.workshop_id=c.workshop_id AND o.weekday_iso=c.weekday_iso AND o.start_minute=c.start_minute AND o.end_minute=c.end_minute)
 ) THEN 1 ELSE 0 END AS active_catalog_ok;
GO
