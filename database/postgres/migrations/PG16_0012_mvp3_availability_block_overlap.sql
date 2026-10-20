-- MVP3 / ADMIN-004: dos bloqueos administrativos activos no pueden ocupar
-- el mismo recurso en periodos intersectados. [inicio, fin): se permite
-- que uno termine exactamente cuando comienza el siguiente.
--
-- Importante: si ya existen bloques activos solapados, la migracion falla
-- deliberadamente. Revisar los datos y resolver el conflicto antes de aplicar.
BEGIN;

DO $migration$
DECLARE
    conflicting_resource integer;
BEGIN
    SELECT a.resource_id
      INTO conflicting_resource
      FROM availability_blocks a
      JOIN availability_blocks b
        ON a.resource_id = b.resource_id
       AND a.id < b.id
       AND a.is_active AND b.is_active
       AND tstzrange(a.start_time, a.end_time, '[)')
           && tstzrange(b.start_time, b.end_time, '[)')
     LIMIT 1;

    IF conflicting_resource IS NOT NULL THEN
        RAISE EXCEPTION
            'hay bloqueos activos solapados para el recurso %; resolver antes de PG16_0012',
            conflicting_resource;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ex_availability_blocks_resource_overlap'
          AND conrelid = 'availability_blocks'::regclass
    ) THEN
        ALTER TABLE availability_blocks
            ADD CONSTRAINT ex_availability_blocks_resource_overlap
            EXCLUDE USING gist (
                resource_id WITH =,
                tstzrange(start_time, end_time, '[)') WITH &&
            ) WHERE (is_active);
    END IF;
END;
$migration$;

COMMIT;
