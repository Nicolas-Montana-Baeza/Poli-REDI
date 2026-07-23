# Poli-REDI - Instalacion y ejecucion local

## Objetivo

Este documento explica como levantar Poli-REDI en ambiente local usando el stack vigente: frontend Vue/Vite, backend Go/Fiber y Azure SQL Database.

## Componentes

- Frontend: aplicacion Vue 3 en `frontend/`.
- Backend: API Go/Fiber en `backend/`.
- Base de datos: Azure SQL Database con scripts T-SQL en `database/`.
- Autenticacion: Microsoft Entra ID, con modo local controlado para pruebas.

## Requisitos

- Node.js y npm.
- Go compatible con `backend/go.mod`.
- Acceso a Azure SQL Database.
- Aplicacion Microsoft Entra ID configurada para frontend y API.

## Variables del backend

Crear `backend/.env` desde `backend/.env.example`.

```env
PORT=3000

DB_SERVER=poli-redi-server.database.windows.net
DB_PORT=1433
DB_NAME=poli-redi-database
DB_USER=poli-redi-admin
DB_PASSWORD=
DB_ENCRYPT=true
DB_TRUST_SERVER_CERTIFICATE=false

ENTRA_TENANT_ID=
ENTRA_API_CLIENT_ID=
ENTRA_ISSUER=
CORS_ALLOWED_ORIGINS=http://localhost:5173

DEV_AUTH_ENABLED=false
```

`DB_PASSWORD` debe vivir solo en `.env` local o en variables de entorno del servicio de despliegue.

Tambien se puede usar `AZURE_SQL_CONNECTION_STRING` como alternativa a las variables `DB_*`.

## Variables del frontend

Crear `frontend/.env` con:

```env
VITE_API_BASE_URL=http://localhost:3000/api
VITE_API_TIMEOUT_MS=30000

VITE_ENTRA_TENANT_ID=
VITE_ENTRA_CLIENT_ID=
VITE_ENTRA_REDIRECT_URI=http://localhost:5173/auth/callback
VITE_ENTRA_POST_LOGOUT_REDIRECT_URI=http://localhost:5173/login
VITE_ENTRA_API_SCOPE=
```

## Preparar base de datos

Los scripts vigentes son T-SQL para Azure SQL Database:

```txt
database/drop.sql
database/schema.sql
database/seed.sql
```

Orden recomendado para una base limpia:

1. Ejecutar `database/drop.sql` si se necesita limpiar objetos existentes.
2. Ejecutar `database/schema.sql`.
3. Ejecutar `database/seed.sql` para cargar datos iniciales.

Para una base MVP 1 ya existente, no repetir el flujo destructivo de
instalacion limpia. Es obligatorio seguir
[`database/migrations/README.md`](../database/migrations/README.md): crear el
backup, abrir una sesion nueva, usar una herramienta compatible con `GO`,
ejecutar `001_mvp2_group_participants.sql`, despues
`002_mvp2_target_participants.sql` y finalmente
`003_open_use_frequency_scope.sql`, comprobando el `POSTCHECK` de cada una
antes de continuar. Esa guia es la unica fuente operativa para la recuperacion o
actualizacion de la base existente.
4. Configurar `backend/.env`.
5. Levantar backend y validar `/api/health`.

No usar `DATABASE_URL`, `pgx`, `psql` ni cadenas PostgreSQL para el entorno actual.

## Ejecutar backend

Desde `backend/`:

```bash
go mod download
go run ./cmd
```

La API queda disponible en:

```txt
http://localhost:3000
```

Validacion publica:

```txt
GET http://localhost:3000/api/health
```

## Ejecutar frontend

Desde `frontend/`:

```bash
npm install
npm run dev
```

La aplicacion queda disponible normalmente en:

```txt
http://localhost:5173
```

## Modo local de prueba

Para probar sin Microsoft Entra ID real:

```env
DEV_AUTH_ENABLED=true
```

Con esta bandera, el frontend muestra accesos locales de prueba y el backend acepta headers `X-Dev-Auth-*`.

No activar `DEV_AUTH_ENABLED=true` en ambientes publicos o productivos.

## Rutas principales

Ruta publica:

```txt
GET /api/health
```

Rutas protegidas:

```txt
GET /api/me
PATCH /api/me/rut
GET /api/resources
GET /api/activities
GET /api/workshops
POST /api/workshops/:id/enroll
GET /api/availability/reservations
GET /api/reservations
GET /api/reservations/mine
POST /api/reservations
PATCH /api/reservations/cancel
GET /api/users
PATCH /api/resources/:id/image
GET /api/notifications
```

## Validaciones tecnicas

Backend:

```bash
cd backend
go test ./...
```

Frontend:

```bash
cd frontend
npm run build
```

## Checklist local rapido

1. Configurar `backend/.env`.
2. Configurar `frontend/.env`.
3. Levantar backend en `http://localhost:3000`.
4. Validar `GET /api/health`.
5. Levantar frontend en `http://localhost:5173`.
6. Iniciar sesion con Entra ID o modo local.
7. Confirmar carga de recursos y actividades.
8. Crear y cancelar una reserva de prueba.
9. Confirmar que usuario normal no accede a rutas admin.

## Nota historica

El proyecto tuvo una etapa anterior documentada con PostgreSQL. Esa informacion queda solo como antecedente historico en `docs/00-revision-inicial.md` y `docs/03-base-de-datos.md`. El estado vigente usa Azure SQL Database y `github.com/microsoft/go-mssqldb`.
