# Poli-REDI - Instalacion y ejecucion local

Fecha de corte: 2026-08-20

## Objetivo

Esta guia describe el entorno local vigente:

- frontend Vue 3 / Vite;
- backend Go / Fiber;
- PostgreSQL 16;
- Podman Quadlet;
- Microsoft Entra ID o autenticacion local controlada.

## Requisitos

- Linux o WSL.
- Podman rootless.
- cgroup v2.
- systemd de usuario / Quadlet.
- Go.
- Node.js y npm.
- Git.

## Preparar PostgreSQL

Desde la raiz del repositorio:

```bash
bash infra/local/quadlet/install.sh install
```

El instalador:

1. genera credenciales locales;
2. las guarda fuera del repositorio;
3. instala el Quadlet;
4. crea el volumen;
5. inicia PostgreSQL 16;
6. aplica la baseline MVP1;
7. carga el seed MVP1.

La instancia local queda en:

```txt
127.0.0.1:55432
```

Los secretos quedan en:

```txt
~/.config/containers/systemd/poliredi-postgres.env
~/.config/poli-redi/backend.env
```

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
```

El instalador automatiza actualmente `0001` a `0003`.

Las migraciones `0004` a `0008` forman parte de MVP2, pero su aplicacion automatica mediante Quadlet sigue pendiente.

Por eso el instalador genera:

```env
MVP_SCOPE=mvp1
```

como valor seguro.

No usar `MVP_SCOPE=mvp2` sobre una base que no tenga las migraciones MVP2 requeridas.

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

VITE_ENTRA_TENANT_ID=
VITE_ENTRA_CLIENT_ID=
VITE_ENTRA_REDIRECT_URI=http://localhost:5173/auth/callback
VITE_ENTRA_POST_LOGOUT_REDIRECT_URI=http://localhost:5173/login
VITE_ENTRA_API_SCOPE=
```

Luego:

```bash
cd frontend
npm install
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

## Operacion

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
