# Poli-REDI

Sistema web para gestion de reservas deportivas institucionales.

Poli-REDI permite consultar disponibilidad de recursos deportivos, crear y cancelar reservas, revisar historial, administrar usuarios y recursos, y visualizar indicadores iniciales de uso. El sistema usa autenticacion con Microsoft Entra ID y datos persistidos en Azure SQL Database.

## Alcance MVP 1

El MVP 1 cubre el flujo base de reservas deportivas:

- Login institucional con Microsoft Entra ID.
- Login local de prueba para desarrollo.
- Registro/actualizacion de RUT para usuarios normales mediante modal obligatorio.
- Consulta de disponibilidad por recurso y fecha.
- Creacion de reservas con usuario autenticado.
- Seleccion de actividades desde catalogo aprobado.
- Listado de mis reservas, detalle, historial y cancelacion.
- Panel administrador base con usuarios, recursos y reportes iniciales.
- Notificaciones internas basicas.
- Demo online inicial en Azure con frontend, backend, base de datos y autenticacion real.

Quedan fuera del MVP 1 la gestion completa de bloqueos, CRUD avanzado de recursos, infracciones, programacion institucional y endurecimiento de despliegue productivo institucional.

Estado de cierre: el MVP 1 esta funcional como demo, pero permanece reabierto para corregir zona horaria, estado controlado por servidor, limites de horario/duracion, pruebas reales y coherencia responsive/accesible. El orden vigente esta en `docs/09-mvps-roadmap.md`; los bloqueantes principales son `RES-009`, `RES-010`, `RES-011`, `QA-001` y `QA-002`.

## Stack

### Frontend

- Vue 3
- Vite
- Pinia
- Vue Router
- Axios
- MSAL Browser

### Backend

- Go
- Fiber
- Microsoft SQL Server driver for Go (`github.com/microsoft/go-mssqldb`)
- Microsoft Entra ID para validacion de tokens JWT

### Base de datos

- Azure SQL Database
- Scripts T-SQL en `database/`

### Despliegue online inicial

- Frontend en Azure Static Web Apps
- Backend en Azure App Service con Docker
- Variables `VITE_*` inyectadas desde GitHub Actions
- Microsoft Entra ID configurado para local y nube

## Estructura del proyecto

```txt
Poli-REDI/
  backend/      API Go/Fiber
  database/     Scripts T-SQL de esquema, datos iniciales y limpieza
  docs/         Documentacion tecnica del proyecto
  frontend/     Aplicacion Vue/Vite
  files/        Archivos de apoyo para datos
```

## Requisitos

- Node.js y npm
- Go compatible con `backend/go.mod`
- Acceso a una base Azure SQL Database
- Aplicacion registrada en Microsoft Entra ID para el frontend y la API

## Configuracion del backend

Crear `backend/.env` a partir de `backend/.env.example`.

Variables principales:

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
APP_TIMEZONE=America/Santiago

# Solo desarrollo local
DEV_AUTH_ENABLED=false
```

`DB_PASSWORD` debe existir solo en `backend/.env` local o en las variables de entorno del despliegue. No debe guardarse en archivos versionados.

Tambien se puede usar `AZURE_SQL_CONNECTION_STRING` como alternativa a las variables `DB_*`, segun la plantilla incluida en `backend/.env.example`.

Para pruebas locales sin Microsoft, se puede usar:

```env
DEV_AUTH_ENABLED=true
```

Con esta opcion, el frontend muestra accesos locales de prueba y el backend acepta headers `X-Dev-Auth-*`. No activar esta bandera en produccion.

## Configuracion del frontend

Crear un archivo `.env` en `frontend/` con las variables de Vite usadas por la autenticacion y la API.

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

`VITE_API_BASE_URL` es opcional; si no se define, el frontend usa `http://localhost:3000/api`.
`VITE_ENTRA_POST_LOGOUT_REDIRECT_URI` permite usar una URL distinta para local y nube sin cambiar codigo.

## Base de datos

Los scripts actuales estan preparados para Azure SQL Database:

```txt
database/drop.sql
database/schema.sql
database/seed.sql
```

Para preparar una base limpia:

1. Ejecutar `database/drop.sql` si se necesita limpiar objetos existentes.
2. Ejecutar `database/schema.sql`.
3. Ejecutar `database/seed.sql` para cargar datos iniciales de desarrollo.
4. Configurar `backend/.env`.
5. Levantar el backend y validar `/api/health`.

No usar scripts ni cadenas de conexion PostgreSQL para el entorno actual.

## Ejecutar backend

Desde `backend/`:

```bash
go mod download
go run ./cmd
```

La API queda disponible por defecto en:

```txt
http://localhost:3000
```

Endpoint publico de verificacion:

```txt
GET http://localhost:3000/api/health
```

Respuesta esperada:

```json
{
  "status": "ok",
  "message": "Poli-REDI API funcionando"
}
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

## Validaciones recomendadas

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

## Rutas principales

Ruta publica:

```txt
GET /api/health
```

Rutas protegidas por token Bearer:

```txt
GET /api/me
PATCH /api/me/rut
GET /api/resources
GET /api/activities
GET /api/reservations
GET /api/reservations/mine
POST /api/reservations
PATCH /api/reservations/cancel
GET /api/users
GET /api/notifications
```

En modo `DEV_AUTH_ENABLED=true`, las rutas protegidas tambien pueden probarse con los headers locales enviados por el frontend de desarrollo.

## Demo online

La demo online inicial usa:

```txt
Frontend: https://purple-ground-0205c9f10.7.azurestaticapps.net/
Backend:  https://poli-redi.azurewebsites.net
Health:   https://poli-redi.azurewebsites.net/api/health
```

Para nube, `CORS_ALLOWED_ORIGINS` debe incluir la URL de Static Web Apps y el frontend debe compilarse con `VITE_API_BASE_URL` apuntando al backend Azure.

## Checklist MVP 1

Antes de una demo local:

1. Levantar backend en `http://localhost:3000`.
2. Levantar frontend en `http://localhost:5173`.
3. Entrar con usuario normal local.
4. Confirmar que solicita RUT en un modal si el usuario no tiene uno.
5. Guardar RUT y verificar que permite avanzar.
6. Crear una reserva desde Disponibilidad.
7. Seleccionar una actividad del catalogo o dejarla sin actividad especifica.
8. Revisar Mis Reservas, Detalle e Historial.
9. Cancelar una reserva propia.
10. Entrar como admin local y verificar acceso al panel administrador.
11. Confirmar que usuario normal no ve ni accede a rutas administrativas.
12. Confirmar que la hora de una reserva coincide entre local y demo online para `America/Santiago` una vez implementado `RES-009`.

El checklist ampliado y la evidencia de la revision exhaustiva estan en `docs/12-checklist-demo-mvp1.md`. `go test ./...` actualmente finaliza sin errores, pero todavia no descubre archivos de pruebas; esa brecha se sigue en `QA-001`.

## Documentacion relacionada

- `docs/01-instalacion-y-ejecucion.md`: preparacion y ejecucion local.
- `docs/02-arquitectura.md`: arquitectura general.
- `docs/03-base-de-datos.md`: modelo Azure SQL Database.
- `docs/06-flujo-reservas.md`: flujo funcional de reservas.
- `docs/07-backlog.md`: backlog maestro y estado de tareas.
- `docs/10-guia-redeploy.md`: ejecucion local y redeploy en Azure.
- `docs/11-plan-corte-google-calendar.md`: plan de transicion desde Google Calendar legado.

## Seguridad

- No versionar archivos `.env`.
- No guardar passwords, cadenas de conexion reales ni secretos de Entra ID en documentos.
- Usar `backend/.env.example` como plantilla segura.
- Si una clave fue compartida fuera del entorno local seguro, rotarla antes de una entrega o despliegue.
