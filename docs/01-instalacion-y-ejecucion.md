# Poli-REDI - Instalacion y ejecucion local

Fecha de corte documental: 2026-10-27
Referencia: cierre tecnico MVP3 del 2026-10-24.

## Objetivo

Esta guía describe el entorno de desarrollo y verificación:

- frontend Vue 3 / Vite;
- backend Go / Fiber;
- PostgreSQL 16;
- Podman;
- Microsoft Entra ID o autenticacion local controlada.

Esta guia distingue tres operaciones independientes: preparar una base nueva, migrar una base existente y actualizar el backend/frontend. Desplegar codigo no aplica automaticamente migraciones PostgreSQL.

## Requisitos

- Linux o WSL.
- Podman rootless.
- cgroup v2.
- systemd de usuario.
- Go.
- Node.js y npm.
- Git.

## Despliegue integrado vigente

El entorno desplegado completo usa Podman Compose desde el repositorio
`~/projects/poliredi-infra`. Incluye PostgreSQL 16, backend y Caddy/frontend,
arranque systemd y publicación mediante Tailscale Funnel.

```bash
systemctl --user status poliredi-stack.service --no-pager
podman ps --format 'table {{.Names}}\t{{.Status}}'
curl --fail --insecure https://localhost:8443/api/health
```

Consultar `docs/10-guia-redeploy.md` para construir, actualizar, respaldar y
recuperar ese stack.

## Quadlet de desarrollo y precaucion con el volumen compartido

El repositorio conserva `infra/local/quadlet/install.sh` para crear una instancia
PostgreSQL de desarrollo. Sin embargo, la configuracion Quadlet y el stack
Podman Compose documentado referencian el **mismo volumen**:
`poliredi-postgres-mvp1-data` y el mismo puerto host `127.0.0.1:55432`.

Por tanto, **no es una instancia aislada del stack integrado**. No ejecutar
`bash infra/local/quadlet/install.sh install` mientras el stack utilice ese
volumen. Nunca iniciar dos servidores PostgreSQL sobre un mismo directorio de
datos ni borrar/recrear ese volumen como procedimiento de actualizacion.

En un directorio de datos inicialmente vacio, el instalador Quadlet copia el
bootstrap local, las migraciones `PG16_0001` a `PG16_0010` y el seed MVP1.
No contempla las cinco migraciones posteriores de MVP3.

Variables privadas del mecanismo Quadlet:

```txt
~/.config/containers/systemd/poliredi-postgres.env
~/.config/poli-redi/backend.env
```

Para pruebas reproducibles sin afectar la instancia habitual existe un
verificador **efimero de MVP2**. La provision completa de MVP3 con aislamiento,
respaldo y comprobaciones se mantiene pendiente en `INFRA-007`.

## Estado de migraciones

Migraciones PostgreSQL presentes:

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

El instalador Quadlet automatiza la cadena `0001` a `0010` en un volumen
nuevo y genera `MVP_SCOPE=mvp1` como valor conservador.

Los scripts de inicializacion de PostgreSQL se ejecutan al inicializar un
directorio de datos vacio; no se repiten al reiniciar un volumen existente.

El backend reconoce los alcances `mvp1`, `mvp2`, `mvp3` y `full` mediante
`MVP_SCOPE`. El frontend selecciona la superficie funcional en compilacion con
`VITE_MVP_SCOPE`.

**Cambiar el scope no ejecuta migraciones.** Para habilitar MVP3 debe
comprobarse que `PG16_0011` a `PG16_0015` esten aplicadas, que sus
precondiciones y privilegios sean correctos y que el runtime sea compatible.

## Diferenciar instalacion, migracion y despliegue

### Preparar una base PostgreSQL nueva

Requiere bootstrap, migraciones en orden, seed adecuado y validacion del
esquema. La automatizacion actual de Quadlet llega a MVP2; la de MVP3 esta
pendiente (`INFRA-007`).

### Actualizar una base PostgreSQL existente

Requiere inventariar el esquema aplicado, crear y verificar un respaldo,
revisar incompatibilidades y ejecutar migraciones incrementales mediante un
procedimiento controlado. No reutilizar scripts de inicializacion de volumen
vacio ni suponer que un archivo SQL versionado ya fue aplicado.

### Desplegar backend y frontend

Cambiar imagenes de backend, variables de entorno o archivos compilados del
frontend **no reinicializa PostgreSQL**. El despliegue requiere compatibilidad
con el esquema presente y comprobaciones de salud posteriores.

Consultar `docs/10-guia-redeploy.md` antes de cualquier intervencion operativa.

## Verificacion limpia MVP2

Desde la raiz del repositorio:

```bash
bash infra/local/quadlet/verify-mvp2-ephemeral.sh
```

La verificacion crea otra base y otro volumen, aplica `0001` a `0010`, ejecuta
la comprobacion SQL, la suite Go con integraciones, `go vet`, las pruebas
frontend y el build. El entorno efimero se elimina al terminar y no modifica
la instancia habitual. Valida las migraciones `0001` a `0010` de MVP2, no la cadena completa MVP3.

## Variables del backend

El backend acepta:

```env
DATABASE_URL=postgres://poliredi_app:password@localhost:55432/poliredi?sslmode=disable
```

o:

```env
PGHOST=localhost
PGPORT=55432
PGDATABASE=poliredi
PGUSER=poliredi_app
PGPASSWORD=
PGSSLMODE=disable
```

Variables adicionales:

```env
PORT=3000
APP_TIMEZONE=America/Santiago
MVP_SCOPE=mvp1
CORS_ALLOWED_ORIGINS=http://localhost:5173

ENTRA_TENANT_ID=
ENTRA_API_CLIENT_ID=
ENTRA_ISSUER=

DEV_AUTH_ENABLED=false
```

## Ejecutar backend

Con las variables generadas por el instalador:

```bash
set -a
source "${XDG_CONFIG_HOME:-$HOME/.config}/poli-redi/backend.env"
set +a

cd backend
go run ./cmd
```

Health:

```txt
GET http://localhost:3000/api/health
```

## Ejecutar frontend

Crear `frontend/.env`:

```env
VITE_API_BASE_URL=http://localhost:3000/api
VITE_API_TIMEOUT_MS=30000
VITE_APP_TIMEZONE=America/Santiago
VITE_MVP_SCOPE=mvp1

VITE_ENTRA_TENANT_ID=
VITE_ENTRA_CLIENT_ID=
VITE_ENTRA_REDIRECT_URI=http://localhost:5173/auth/callback
VITE_ENTRA_POST_LOGOUT_REDIRECT_URI=http://localhost:5173/login
VITE_ENTRA_API_SCOPE=
```

Luego:

```bash
cd frontend
npm ci
npm run dev
```

## Autenticacion local

Solo para desarrollo:

```env
DEV_AUTH_ENABLED=true
```

No habilitar en ambientes publicos.

## Validaciones

Backend:

```bash
cd backend
go test ./...
go vet ./...
```

Frontend:

```bash
cd frontend
npm test
npm run build
```

PostgreSQL MVP1:

```bash
bash infra/local/quadlet/verify-mvp1.sh
```

## Operación de la base aislada de desarrollo

```bash
bash infra/local/quadlet/install.sh status
bash infra/local/quadlet/install.sh logs
bash infra/local/quadlet/install.sh stop
bash infra/local/quadlet/install.sh start
```

## SQL Server / Azure SQL legacy

Los archivos:

```txt
database/schema.sql
database/seed.sql
database/drop.sql
database/seed_today_temp.sql
database/queries.sql
database/queries copy.sql
```

pertenecen a una etapa arquitectonica anterior.

No deben utilizarse para inicializar PostgreSQL.

La historia completa se registra en:

```txt
database/README.md
docs/14-evolucion-y-trazabilidad-requisitos.md
```

## Evolucion de persistencia

La genealogia identificada es:

```txt
2026-05-23  PostgreSQL inicial
      ↓
2026-07-03  Azure SQL Database
      ↓
2026-08-14  nueva baseline PostgreSQL 16
      ↓
2026-08-17  MVP1 estable integrado con PostgreSQL
```

PostgreSQL 16 es la arquitectura vigente.
