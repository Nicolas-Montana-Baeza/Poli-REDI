# Cambios respecto de los archivos recibidos

## Correcciones funcionales

1. **Auditoría nullable:** `reservation_target_audit.old_target_participants` y `new_target_participants` pasan a aceptar `NULL`, coherente con `reservations.target_participants`.
2. **Triggers redundantes:** se retiran `trg_reservations_pending_conflicts`, `trg_blocks_pending_conflicts` y `trg_scheduled_activities_pending_conflicts`. Las definiciones canónicas ya validan `PENDING` y `CONFIRMED`; mantener ambos conjuntos duplicaba errores y podía anular la semántica compartida de `OPEN_USE`.
3. **RUT futuro:** se incorpora `trg_users_rut_validate`, por lo que nuevos INSERT/UPDATE también validan formato canónico y dígito verificador. El trigger `write_once` continúa impidiendo cambios posteriores.
4. **Duraciones por política:** el esquema ya no usa una comprobación global de tabla vacía; completa cada combinación política/duración faltante.
5. **Inmutabilidad temporal:** una política puede pasar de vigente a cerrada una sola vez, pero no puede reabrirse ni cambiar posteriormente su `effective_to`.
6. **Auditoría consistente:** la auditoría de participantes es append-only y se evita el doble registro de reservas provocado por el trigger de `updated_at`.
7. **Talleres:** un índice único filtrado impide dos inscripciones `CONFIRMED` del mismo usuario al mismo taller.
8. **Snapshot grupal:** las reservas grupales nuevas deben copiar la capacidad vigente del recurso; las no grupales no pueden almacenar ese valor.

## Seguridad operacional

- `drop.sql` exige `allow_destructive_reset=1`.
- `seed.sql` exige `allow_development_seed=1`.
- `seed_today_temp.sql` exige `allow_today_seed=1`.
- `queries.sql` deja las modificaciones deshabilitadas por defecto.

## Rendimiento y mantenibilidad

- Índices para solapes por recurso/usuario, participaciones, bloqueos, actividades y talleres.
- Nombres normalizados sin sufijos `(1)`, `(2)` o `(3)`.
- README reorganizado por escenario de instalación y upgrade.
- Nuevo `verify.sql` de solo lectura.
- Nueva migración `009` para llevar una base ya migrada al estado canónico mejorado.

## Límite de la revisión

Se realizó revisión estática, comprobación de estructura por lotes `GO`, inventario de objetos y consistencia entre esquema y migraciones. No se ejecutaron los scripts contra una instancia real de Azure SQL/SQL Server, por lo que deben probarse en una copia recuperable antes del despliegue.
