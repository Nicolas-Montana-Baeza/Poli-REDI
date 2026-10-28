# Poli-REDI - Base de datos y modelo de persistencia

Fecha de corte documental: 2026-10-27
Referencia: cierre tecnico MVP3 del 2026-10-24.

## Objetivo del documento

Este documento describe el modelo de datos funcional de Poli-REDI y la evolucion de su persistencia.

La fuente de verdad tecnica vigente es PostgreSQL 16. La etapa Azure SQL se conserva como antecedente arquitectonico dentro de `EV-011`.

## Estado actual

Motor vigente:

```txt
PostgreSQL 16
```

Driver backend:

```txt
github.com/jackc/pgx/v5
```

Configuracion:

```env
DATABASE_URL=postgres://...
PGHOST=
PGPORT=
PGDATABASE=
PGUSER=
PGPASSWORD=
PGSSLMODE=
```

El backend rechaza URLs de base de datos con esquema distinto de `postgres` o `postgresql`.

## Estructura canonica

```txt
database/postgres/bootstrap/
database/postgres/migrations/
database/postgres/seed/
database/postgres/check/
```

Migraciones vigentes:

```txt
PG16_0001_mvp1_baseline.sql
PG16_0002_mvp1_indexes.sql
PG16_0003_mvp1_invariants.sql
PG16_0004_mvp2_group_participants.sql
PG16_0005_mvp2_institutional_scheduling.sql
PG16_0006_mvp2_institutional_availability.sql
PG16_0007_mvp2_schedule_exceptions.sql
PG16_0008_mvp2_schedule_exception_availability.sql
PG16_0009_full_notifications.sql
PG16_0010_mvp2_group_resource_rules.sql
PG16_0011_mvp3_reservation_venue_guard.sql
PG16_0012_mvp3_availability_block_overlap.sql
PG16_0013_mvp3_occurrence_revisions.sql
PG16_0014_mvp3_user_access_audit.sql
PG16_0015_mvp3_user_role_audit.sql
```

La migracion `0010` agrega minimo por recurso grupal, snapshot del minimo,
Sala Multiuso como recurso grupal y una politica prospectiva con deadline de
60 minutos. `0009` mantiene la tabla fisica de notificaciones, aunque sus
generadores generales permanecen fuera del cierre funcional MVP2.

La cadena MVP2 puede verificarse sobre una base descartable mediante
`infra/local/quadlet/verify-mvp2-ephemeral.sh` y
`database/postgres/check/PG16_verify_mvp2.sql`.

### Ampliaciones MVP3

- `PG16_0011`: impide reservas activas en sedes o recursos inactivos.
- `PG16_0012`: evita solapamientos de bloqueos administrativos activos mediante una restriccion GiST. Falla si existen conflictos previos.
- `PG16_0013`: incorpora historial de revisiones de excepciones de programacion institucional.
- `PG16_0014`: registra bloqueos y desbloqueos de usuarios en `user_access_audit`.
- `PG16_0015`: registra asignaciones y revocaciones de roles en `user_role_audit`.

### Limitacion del instalador

`infra/local/quadlet/install.sh` incorpora actualmente las migraciones
`PG16_0001` a `PG16_0010`, pero no las cinco migraciones MVP3.

Los volumenes PostgreSQL existentes no vuelven a ejecutar los scripts
de inicializacion. Las migraciones posteriores deben aplicarse mediante
un procedimiento controlado, con respaldo, comprobaciones previas
y verificacion posterior.

La presencia de una migracion en Git no demuestra que haya sido
aplicada en una instancia concreta.

## Genealogia EV-011

1. PostgreSQL inicial - 2026-05-23.
2. Migracion deliberada a Azure SQL - 2026-07-03.
3. Nueva baseline PostgreSQL 16 - 2026-08-14.
4. Integracion estable MVP1 PostgreSQL - 2026-08-17.

La etapa Azure SQL fue valida durante julio de 2026 y produjo evidencia real de QA. No debe borrarse de la historia, pero tampoco describirse como motor actual.

## SQL Server legacy

Los archivos en la raiz de `database/` (`schema.sql`, `seed.sql`, `drop.sql`, `seed_today_temp.sql`, `queries*.sql`) pertenecen a la etapa SQL Server/Azure SQL.

El backend compilable ya no conserva consultas T-SQL. Talleres y actividades programadas legacy fueron retirados; notificaciones y publicacion de politicas fueron migradas a PostgreSQL. Los scripts SQL Server se conservan solo como historia arquitectonica.

Consultar `database/README.md` para la clasificacion operativa actual.

## Nota sobre las entidades descritas abajo

Las secciones siguientes contienen un modelo conceptual heredado.
No representan necesariamente el esquema fisico PostgreSQL vigente.

En PostgreSQL 16, `reservations.start_time` y `reservations.end_time`
utilizan `timestamptz`. La zona de negocio se define mediante
`APP_TIMEZONE=America/Santiago`.

El tipo `DATETIME2` pertenece a la arquitectura historica Azure SQL.
La fuente de verdad del esquema vigente son las migraciones `PG16_*`.

Las secciones siguientes conservan el modelo conceptual de entidades construido durante el proyecto. Cuando aparezcan tipos o mecanismos especificamente SQL Server (`DATETIME2`, `dbo`, `UPDLOCK`, `HOLDLOCK`), deben interpretarse como descripcion de la version Azure SQL previa, no como contrato de la persistencia PostgreSQL vigente.

## Tablas principales

### `venues`

Representa una sede o recinto institucional. Permite asociar recursos deportivos a ubicaciones fisicas.

Campos relevantes:

- `id`
- `name`
- `address_line`
- `commune`
- `city`
- `region`
- `country`
- `latitude`
- `longitude`
- `is_active`

### `users`

Representa usuarios autenticados mediante Microsoft Entra ID.

Campos relevantes:

- `id`
- `email`
- `full_name`
- `is_admin`
- `is_blocked`
- `entra_oid`
- `tenant_id`

### `resources`

Representa recursos deportivos asociados a una sede.

Ejemplos:

- Cancha de Futbol
- Cancha de Basquetbol
- Piscina
- Gimnasio
- Sala Multiuso

Modos de reserva:

- `RESERVABLE`
- `OPEN_USE`
- `INFORMATIVE`
- `ADMIN_ONLY`

`OPEN_USE` representa recursos de uso libre, como piscina o gimnasio, donde varias reservas pueden coexistir para medir asistencia o intensidad de uso sin bloquear el recurso completo.

`image_url` permite asociar una URL `http`, `https` o ruta local iniciada en `/` para mostrar imagenes del recurso en dashboard y catalogo. Es opcional y puede quedar `NULL`.

### `activities`

Representa actividades asociadas a reservas o programacion institucional.

Ejemplos:

- Futbol
- Basquetbol
- Natacion
- Entrenamiento Libre
- Yoga
- Campeonato

### `reservations`

Representa reservas realizadas por usuarios.

Estados permitidos:

- `PENDING`
- `CONFIRMED`
- `CANCELLED`
- `REJECTED`
- `EXPIRED`

Campos temporales relevantes:

- `start_time`: `DATETIME2`, sin zona horaria embebida.
- `duration_minutes`: duracion numerica usada para calcular el termino.

Contrato temporal de reservas para MVP 1:

- `start_time` y los rangos de disponibilidad guardan hora institucional de muro en `America/Santiago`.
- `DATETIME2` no contiene zona. Un valor SQL `2026-07-14 10:30:00` significa 10:30 de Chile, no 10:30 UTC.
- El backend asigna `APP_TIMEZONE` al leer estos campos y responde RFC 3339 con el offset real. Ejemplo de invierno: `2026-07-14T10:30:00-04:00`.
- Un request sin offset, por ejemplo `2026-07-14T10:30:00`, se interpreta en `APP_TIMEZONE`.
- Un request con `Z` u otro offset se convierte primero a la hora equivalente de Santiago.
- `created_at`, `updated_at` y otros campos generados mediante `SYSUTCDATETIME()` continuan representando UTC.
- La estrategia conserva las horas de las reservas existentes y evita una conversion masiva de datos.

### `participants`

Representa participantes asociados a una reserva.

Estados permitidos:

- `PENDING`
- `CONFIRMED`
- `REJECTED`
- `CANCELLED`

### `availability_blocks`

Representa bloqueos administrativos, mantenciones o cierres sobre un recurso.

Tipos permitidos:

- `MAINTENANCE`
- `ADMINISTRATIVE`
- `EVENT`
- `CLOSED`
- `OTHER`

### `scheduled_activities`

Representa programacion institucional, como clases, talleres, eventos, campeonatos o entrenamientos.

Tipos permitidos:

- `CLASS`
- `WORKSHOP`
- `EVENT`
- `CHAMPIONSHIP`
- `TRAINING`
- `OTHER`

### `workshops`

Representa talleres deportivos recurrentes disponibles para inscripcion de estudiantes.

Cada taller queda asociado a un `resource_id`. Esa relacion permite que backend y frontend lo traten como ocupacion del recurso durante sus dias y horarios recurrentes.

Campos relevantes:

- `id`
- `title`
- `description`
- `day_text`
- `schedule_text`
- `location`
- `instructor_name`
- `capacity`
- `is_active`

### `workshop_enrollments`

Representa inscripciones de usuarios a talleres deportivos.

Estados permitidos:

- `CONFIRMED`
- `CANCELLED`

### `violations`

Representa infracciones o incumplimientos de usuarios.

Tipos permitidos:

- `NO_SHOW`
- `LATE_CANCEL`
- `MISUSE`
- `PARTICIPANTS_NOT_MET`
- `OTHER`

### `notifications`

Representa notificaciones internas del sistema.

Tipos permitidos:

- `RESERVATION_CREATED`
- `RESERVATION_CONFIRMED`
- `RESERVATION_CANCELLED`
- `RESERVATION_MODIFIED`
- `REMINDER`
- `SYSTEM`

### `audit_logs`

Representa auditoria y trazabilidad de acciones relevantes.

## Reglas de negocio en base de datos

Las descripciones siguientes conservan algunas reglas de la etapa
SQL Server/Azure SQL. Sus nombres de triggers no deben interpretarse
automaticamente como nombres actuales de objetos PostgreSQL.

En PostgreSQL las reglas se implementan mediante constraints,
indices, triggers y servicios Go.

MVP3 incorpora, entre otras protecciones, la validacion de sedes
y recursos de `PG16_0011` y la exclusion de bloqueos de `PG16_0012`.

### Reglas de reservas

El trigger `trg_reservations_validate_conflicts` valida que:

- Un usuario bloqueado no pueda crear reservas.
- Un recurso inactivo no pueda ser reservado.
- Un recurso `INFORMATIVE` no pueda ser reservado.
- Un recurso `ADMIN_ONLY` solo pueda ser reservado por administradores.
- No existan reservas confirmadas solapadas para el mismo recurso.
- Un usuario no tenga dos reservas confirmadas en el mismo horario.
- No se reserve durante un bloqueo activo.
- No se reserve durante una actividad institucional programada.

Limites conocidos:

- Las reglas de solapamiento dependen de estados activos como `CONFIRMED`; por eso el endpoint publico no debe permitir que el cliente escoja `status` (`RES-010`).
- La base solo comprueba que `duration_minutes` sea positivo. Jornada, pasos y duraciones permitidas deben validarse en servicio (`RES-011`).
- Las comparaciones de pasado, en curso y finalizada pertenecen a la capa de negocio y deben usar la zona definida en `RES-009`.

### Reglas de bloqueos

El trigger `trg_blocks_validate_conflicts` valida que:

- No existan bloqueos activos solapados sobre el mismo recurso.
- No se cree un bloqueo sobre una reserva confirmada existente.

### Reglas de programacion institucional

El trigger `trg_scheduled_activities_validate_conflicts` valida que:

- No existan actividades programadas solapadas sobre el mismo recurso.
- Una actividad programada no se cruce con un bloqueo activo.
- Una actividad programada no se cruce con una reserva confirmada.

### Reglas de talleres

La base de datos registra talleres activos y sus inscripciones. El backend complementa estas reglas con transaccion serializable para validar cupos antes de insertar una inscripcion.

- Cada taller apunta a un recurso existente.
- Cada taller debe tener capacidad mayor a cero.
- Una inscripcion apunta a un taller y a un usuario existente.
- Solo puede existir una inscripcion `CONFIRMED` por usuario y taller.
- El contador de inscritos se calcula desde `workshop_enrollments`.

### Auditoria y notificaciones

- `trg_violations_notify`: genera una notificacion cuando se registra una infraccion.
- `trg_reservations_audit`: registra cambios sobre reservas en `audit_logs`.

## Vistas

### `vw_resource_usage`

Resume uso de recursos por sede y recurso.

### `vw_peak_hours`

Resume cantidad de reservas confirmadas por hora del dia.

### `vw_user_violations`

Resume infracciones por usuario.

### `vw_resource_calendar`

Unifica reservas, bloqueos y actividades programadas para mostrar disponibilidad/calendario.

## Diagrama ER

**Estado:** diagrama conceptual heredado, pendiente de regeneracion
a partir de las quince migraciones PostgreSQL 16.

No representa exhaustivamente las tablas de politicas versionadas,
unidades institucionales, revisiones y auditorias MVP3.

```mermaid
erDiagram
    VENUES ||--o{ RESOURCES : has
    USERS ||--o{ RESERVATIONS : creates
    USERS ||--o{ PARTICIPANTS : joins
    USERS ||--o{ AVAILABILITY_BLOCKS : creates
    USERS ||--o{ SCHEDULED_ACTIVITIES : creates
    USERS ||--o{ WORKSHOP_ENROLLMENTS : enrolls
    USERS ||--o{ VIOLATIONS : receives
    USERS ||--o{ NOTIFICATIONS : receives
    USERS ||--o{ AUDIT_LOGS : performs

    RESOURCES ||--o{ RESERVATIONS : booked_for
    RESOURCES ||--o{ AVAILABILITY_BLOCKS : blocked_by
    RESOURCES ||--o{ SCHEDULED_ACTIVITIES : scheduled_for

    ACTIVITIES ||--o{ RESERVATIONS : classifies
    ACTIVITIES ||--o{ SCHEDULED_ACTIVITIES : classifies
    WORKSHOPS ||--o{ WORKSHOP_ENROLLMENTS : has

    RESERVATIONS ||--o{ PARTICIPANTS : has
    RESERVATIONS ||--o{ VIOLATIONS : may_generate
    RESERVATIONS ||--o{ NOTIFICATIONS : may_notify

    VENUES {
        int id PK
        string name UK
        string address_line
        string commune
        string city
        string region
        string country
        decimal latitude
        decimal longitude
        boolean is_active
    }

    USERS {
        int id PK
        string email UK
        string full_name
        boolean is_admin
        boolean is_blocked
        string entra_oid
        string tenant_id
    }

    RESOURCES {
        int id PK
        int venue_id FK
        string name
        string type
        string reservation_mode
        string image_url
        int capacity
        boolean is_active
    }

    ACTIVITIES {
        int id PK
        string name UK
        string description
        boolean is_active
    }

    RESERVATIONS {
        int id PK
        int user_id FK
        int resource_id FK
        int activity_id FK
        datetime start_time
        int duration_minutes
        string status
        string cancellation_reason
    }

    PARTICIPANTS {
        int id PK
        int reservation_id FK
        int user_id FK
        string status
        datetime confirmed_at
    }

    AVAILABILITY_BLOCKS {
        int id PK
        int resource_id FK
        int created_by_user_id FK
        string block_type
        string reason
        datetime start_time
        datetime end_time
        boolean is_active
    }

    SCHEDULED_ACTIVITIES {
        int id PK
        int resource_id FK
        int activity_id FK
        int created_by_user_id FK
        string title
        string activity_type
        datetime start_time
        datetime end_time
        string recurrence_rule
        boolean is_active
    }

    WORKSHOPS {
        int id PK
        string title
        string description
        string day_text
        string schedule_text
        string location
        string instructor_name
        int capacity
        boolean is_active
    }

    WORKSHOP_ENROLLMENTS {
        int id PK
        int workshop_id FK
        int user_id FK
        string status
    }

    VIOLATIONS {
        int id PK
        int user_id FK
        int reservation_id FK
        int created_by_user_id FK
        string violation_type
        string description
    }

    NOTIFICATIONS {
        int id PK
        int user_id FK
        int reservation_id FK
        string title
        string message
        string type
        boolean is_read
    }

    AUDIT_LOGS {
        int id PK
        int user_id FK
        string action
        string entity_type
        int entity_id
        string description
    }
```

## Estado verificado

Referencia: cierre tecnico MVP3 del 2026-10-24,
documentado en `docs/mvp3/cierre-mvp3-prototipo.md`.

- Backend Go y PostgreSQL 16 operativos en el prototipo.
- Frontend Vue servido mediante Caddy.
- Autenticacion de desarrollo deshabilitada en el entorno validado.
- Pruebas de integracion PostgreSQL registradas para modulos administrativos.
- 43 pruebas frontend de utilidades aprobadas.
- 73 pruebas de componentes Vue aprobadas.
- Compilacion Vite aprobada.

Estas evidencias no sustituyen una auditoria de las migraciones
efectivamente aplicadas, los privilegios y los datos existentes.

El cierre acredita un prototipo validado, no un sistema certificado
para produccion institucional.

## Pendientes tecnicos

- Inventariar las migraciones efectivamente aplicadas en cada instancia.
- Actualizar el bootstrap de desarrollo para incluir MVP3.
- Validar la cadena completa en una base efimera.
- Documentar un procedimiento seguro de migracion y rollback.
- Revisar los privilegios efectivos del rol `poliredi_app`.
- Verificar especialmente los permisos de las tablas de auditoria.
- Regenerar el diagrama ER con las entidades PostgreSQL vigentes.
- Ampliar pruebas de integracion, concurrencia y seguridad.
- Definir politicas de retencion de notificaciones y auditorias.
- Completar el alcance pendiente de infracciones (`REP-002`).
- Resolver la discrepancia de RF-022 para reservas `CONFIRMED + AT_RISK`.

No se deben aplicar migraciones sobre una base existente
sin respaldo, inspeccion previa y procedimiento aprobado.

## Politicas versionadas

La administracion prospectiva de politicas se encuentra implementada
para el alcance del prototipo MVP3 (`ADMIN-006`).

La configuracion grupal por recurso se representa mediante
`reservation_policy_group_resources`. Cada reserva conserva su
politica y los snapshots correspondientes.

Las correcciones excepcionales sobre reservas existentes permanecen
fuera del cierre MVP3.

Los parrafos historicos que mencionan `schema.sql`, `seed.sql`
y tipos SQL Server describen la etapa Azure SQL, no el esquema
PostgreSQL actual.

Estado historico: la version Azure SQL fue aprobada y probada durante julio. Estado vigente: PostgreSQL 16 reemplaza esa arquitectura; las reglas deben verificarse contra las migraciones `PG16_*` y las pruebas de integracion actuales.

El modelo agrega:

- `reservation_policies`: versiones inmutables con ventana reservable, frecuencia, plazo, minimo, `effective_from` y `effective_to`.
- `reservation_policy_resources`: relacion unica entre version y recursos permitidos para reservar bajo esa politica. No clasifica recursos para confirmacion grupal.
- `reservation_policy_durations`: snapshot de duraciones permitidas por version.
- `reservations.policy_id`: version aplicada a la solicitud; se agrega nullable, se completa con la version inicial y luego pasa a `NOT NULL`.
- Jornada, intervalo, clave/hash de idempotencia, autoria, vigencias UTC y estado de publicacion por version.
- `reservation_policy_scope_migrations`: marca el bootstrap tecnico, unico y acotado, del alcance heredado de la politica inicial.

Debe existir una sola version vigente. Publicar una politica cierra la anterior y activa inmediatamente la nueva dentro de una transaccion serializable. Indices filtrados impiden dos versiones vigentes y claves de idempotencia duplicadas. Triggers protegen de edicion o eliminacion las politicas publicadas y sus colecciones. Las reservas conservan `policy_id`; las claves foraneas impiden eliminar su politica mientras exista historial relacionado.

En la etapa historica SQL Server/Azure SQL, `schema.sql`
creaba la estructura, la tabla de marca y los triggers del
bootstrap tecnico controlado. Posteriormente, `seed.sql`
cargaba primero los recursos y, dentro de una transaccion, completa una sola vez los recursos permitidos de la politica inicial y registra la marca de migracion. Este flujo `schema.sql` -> `seed.sql` no es el mecanismo de publicacion administrativa.

La persistencia y los endpoints de correcciones excepcionales no estan implementados en este incremento. Su contrato aprobado exige solicitudes futuras activas, vista previa temporal de un solo uso vinculada al administrador, motivo, revalidacion y aplicacion atomica auditada, sin cancelaciones implicitas.

Los conflictos de recursos grupales deben considerar `PENDING` y `CONFIRMED`. Confirmacion, retiro, cancelacion y vencimiento deben serializarse mediante mecanismos transaccionales PostgreSQL. `UPDLOCK/HOLDLOCK` corresponde al diseno Azure SQL historico y no debe introducirse en migraciones nuevas.

## Recomendacion siguiente

1. Verificar la trazabilidad de las quince migraciones PostgreSQL.
2. Actualizar `docs/01-instalacion-y-ejecucion.md`.
3. Revisar `docs/10-guia-redeploy.md` para despliegue MVP3.
4. Reconstruir el diagrama ER desde el esquema vigente.
5. Documentar los privilegios y auditorias de usuarios.
6. Mantener separadas la validacion del prototipo y la
   habilitacion de produccion institucional.
