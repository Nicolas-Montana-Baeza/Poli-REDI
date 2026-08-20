# Poli-REDI - Base de datos

Fecha de corte: 2026-08-20

## Motor vigente

El runtime actual de Poli-REDI utiliza PostgreSQL 16.

El backend se conecta mediante:

```txt
github.com/jackc/pgx/v5
```

La conexion acepta:

- `DATABASE_URL` con esquema `postgres://` o `postgresql://`;
- `PGHOST`;
- `PGPORT`;
- `PGDATABASE`;
- `PGUSER`;
- `PGPASSWORD`;
- `PGSSLMODE`.

`backend/internal/database/database.go` utiliza `pgx` como driver vigente.

## Estructura canonica

La fuente de verdad actual del esquema es:

```txt
database/postgres/
├── bootstrap/
├── check/
├── migrations/
└── seed/
```

Migraciones disponibles:

```txt
PG16_0001_mvp1_baseline.sql
PG16_0002_mvp1_indexes.sql
PG16_0003_mvp1_invariants.sql
PG16_0004_mvp2_group_participants.sql
PG16_0005_mvp2_institutional_scheduling.sql
PG16_0006_mvp2_institutional_availability.sql
PG16_0007_mvp2_schedule_exceptions.sql
PG16_0008_mvp2_schedule_exception_availability.sql
```

## Scripts SQL Server legacy

Los siguientes archivos pertenecen a la etapa Azure SQL / SQL Server:

```txt
database/schema.sql
database/seed.sql
database/drop.sql
database/seed_today_temp.sql
database/queries.sql
database/queries copy.sql
```

Estos archivos contienen T-SQL y no son la fuente de verdad del runtime actual.

Se conservan temporalmente por trazabilidad y porque aun existe codigo legacy pendiente de retiro.

## Estado de automatizacion local

`infra/local/quadlet/install.sh` levanta PostgreSQL 16 con Podman Quadlet.

Actualmente automatiza:

- bootstrap del rol local;
- `PG16_0001`;
- `PG16_0002`;
- `PG16_0003`;
- seed MVP1.

Por lo tanto, el provisionamiento automatico sigue siendo una baseline MVP1.

Las migraciones MVP2 `PG16_0004` a `PG16_0008` ya existen, pero aun deben incorporarse al instalador para disponer de provisionamiento MVP2 automatico.

## Deuda SQL Server activa

EV-011 detecto varios restos de la etapa SQL Server.

### Talleres legacy retirados

La auditoria de 2026-08-20 confirmo que los handlers/services antiguos de talleres habian sido reemplazados por el modulo de talleres institucionales MVP2.

Tambien se retiro el parser textual legacy utilizado durante la creacion de reservas. PostgreSQL `PG16_0006` protege directamente los solapes con programacion institucional, incluidos talleres `WORKSHOP`.

Se retiraron:

```txt
backend/internal/services/workshops_service.go
backend/internal/repositories/workshops_repository.go
backend/internal/handlers/workshops_handlers.go
backend/internal/models/workshop.go
```

Este retiro permitio eliminar `go-mssqldb` como dependencia directa del backend.

### Notificaciones migradas

`backend/internal/repositories/notifications_repository.go` utiliza PostgreSQL. `PG16_0009_full_notifications.sql` crea la persistencia requerida por la superficie `FULL` sin cerrar prematuramente los tipos futuros de notificacion.

### Politicas completamente PostgreSQL

`backend/internal/repositories/reservation_policies_repository.go` utiliza PostgreSQL para lectura, historial y publicacion. La publicacion serializa escritores, mantiene idempotencia y preserva la clasificacion grupal vigente para los recursos que continuan habilitados.

### Actividades programadas legacy retiradas

`backend/internal/repositories/scheduled_activities_repository.go` no tenia consumidores externos y fue reemplazado por la infraestructura `institutional_activities` PostgreSQL MVP2.

El tipo `ScheduledActivity` se conserva porque la infraestructura institucional vigente lo utiliza como DTO de disponibilidad.

## Regla para nuevos desarrollos

Todo cambio nuevo de persistencia debe:

1. utilizar PostgreSQL;
2. implementarse mediante migraciones `PG16_*`;
3. evitar T-SQL nuevo;
4. mantener consistencia con `America/Santiago`;
5. preservar integridad y concurrencia en base de datos cuando corresponda;
6. incorporar pruebas PostgreSQL cuando la regla dependa de persistencia o concurrencia.

La etapa Azure SQL se conserva como historia arquitectonica, no como motor vigente.
