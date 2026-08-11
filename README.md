# Poli-REDI

> **Canon documental vigente desde 2026-08-11:** consultar [`docs/Propuesta nueva/README.md`](docs/Propuesta%20nueva/README.md) y el [`índice de gobierno documental`](docs/Propuesta%20nueva/00-indice-y-gobierno-documental.md). La serie anterior `docs/00` a `docs/04` está supersedida como fuente operativa. El estado vigente reconoce MVP 1 demostrable local —no online—, MVP 2 parcial, MVP 3 parcial y MVP 4 pendiente/acotado. La fecha límite absoluta del prototipo y su documentación es el **2026-12-10**.

Sistema web para gestión de reservas deportivas institucionales.

Poli-REDI permite consultar disponibilidad de recursos deportivos, crear y cancelar reservas, revisar historial, consultar información administrativa básica y visualizar indicadores iniciales de uso. El sistema usa autenticación con Microsoft Entra ID y datos persistidos en Azure SQL Database.

## Alcance MVP 1

El MVP 1 cubre el flujo base de reservas deportivas:

- Login institucional con Microsoft Entra ID.
- Login local de prueba para desarrollo.
- Registro único de RUT para usuarios normales mediante modal obligatorio; después queda visible en modo solo lectura.
- Consulta de disponibilidad por recurso y fecha.
- Creación de reservas con usuario autenticado.
- Selección de actividades desde catálogo aprobado.
- Listado de mis reservas, detalle, historial básico de reservas y cancelación.
- Consulta administrativa básica de usuarios, recursos e indicadores, sin gestión avanzada.

Quedan fuera del MVP 1 la gestión completa de bloqueos, CRUD avanzado de recursos, infracciones, programación institucional y endurecimiento de despliegue productivo institucional.

El historial incluido en MVP 1 corresponde exclusivamente a reservas propias o
reservas en las que el usuario participa. La consulta de inscripciones a talleres
se incorpora como una ampliacion controlada de MVP 2. El historial de clases,
actividades institucionales y otros eventos se reserva para MVP 3; estos elementos
solo podran formar parte del historial personal cuando exista una relacion
explicita entre el usuario y la actividad.

Estado de cierre: el MVP 1 está funcional como demo local. Zona horaria, estado controlado por servidor, límites de horario/duración y mejoras responsive/accesibles de los flujos críticos cuentan con evidencia local; falta revalidarlos en el ambiente integrado/online. El estado vigente está en el [resumen ejecutivo canónico](docs/Propuesta%20nueva/01-resumen-ejecutivo-y-estado.md).

Para MVP 2 y MVP 3 se aprobaron reglas de ventana, frecuencia, participantes y prioridad institucional. El flujo grupal tiene evidencia local de objetivo, progreso, código/enlace, confirmación, retiro, reconfirmación, deadline inclusivo y expiración `CANCELLED`, pero no está cerrado online. La expiración genera una notificación específica localmente. La notificación específica asociada a prioridad depende de MVP 3; el sistema core de notificaciones, lectura, destinos y demás eventos corresponde a MVP 4.

Registro histórico de revisión técnica del 2026-07-30 —no corresponde al estado vigente—: el detalle de reserva se reutiliza en
Disponibilidad, Mis Reservas, Historial y confirmacion mediante codigo; las
tarjetas personales son seleccionables completas y los dialogos restauran foco,
admiten Escape y operacion por teclado. El dashboard evita duplicar la proxima
reserva. Los talleres y sus inscripciones propias son alcance de MVP 2; clases y
otros eventos institucionales permanecen en MVP 3. La gestión avanzada de
campeonatos y la detección automatizada de abuso quedan fuera del prototipo.

La secuencia de migraciones aprobada comprende `001` a `008`, en ese orden. Cada
script debe ejecutarse completo, incluidos sus prechecks cuando los incorpora.
`007` repara solamente la política bootstrap inequívocamente reconocible, sin
alterar reservas históricas; `008` incorpora la defensa de solapes personales
definida por su contrato. `007` y `008` requieren backup, precheck, postcheck,
reejecución idempotente y validación en Azure SQL antes del despliegue. `009` está
explícitamente excluida: es una propuesta no aprobada y no forma parte de la
secuencia ejecutable.

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
- Microsoft Entra ID para validación de tokens JWT

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
  docs/         Documentación técnica del proyecto
  frontend/     Aplicacion Vue/Vite
  files/        Archivos de apoyo para datos
```

## Requisitos

- Node.js y npm
- Go compatible con `backend/go.mod`
- Acceso a Azure SQL Database o Docker para ejecutar SQL Server localmente
- Aplicación registrada en Microsoft Entra ID para el frontend y la API

## Configuración del backend

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

# Nota: DB_TRUST_SERVER_CERTIFICATE=true es aceptable solamente en un entorno local de SQL Server
# accesible mediante localhost o 127.0.0.1. En Azure o entornos públicos debe ser false.
```

`DB_PASSWORD` debe existir solo en `backend/.env` local o en las variables de entorno del despliegue. No debe guardarse en archivos versionados.

Tambien se puede usar `AZURE_SQL_CONNECTION_STRING` como alternativa a las variables `DB_*`, segun la plantilla incluida en `backend/.env.example`.

### SQL Server local con Docker

1. Copiar `.env.example` a `.env` en la raiz y reemplazar
   `MSSQL_SA_PASSWORD` por una contraseña local fuerte.
2. Iniciar SQL Server:

```bash
docker compose up -d db
```

3. Configurar `backend/.env` para conectarse al puerto publicado:

```env
DB_SERVER=localhost
DB_PORT=1433
DB_NAME=poli-redi-database
DB_USER=poli-redi-admin
DB_PASSWORD=<mismo valor de MSSQL_SA_PASSWORD>
DB_ENCRYPT=true
DB_TRUST_SERVER_CERTIFICATE=true
```

En el primer inicio se crean la base, el esquema y los datos iniciales. Los
reinicios posteriores conservan la información en el volumen
`poli-redi-mssql`. El contenedor también crea el login local
`poli-redi-admin`, de modo que no es necesario cambiar `DB_USER` al alternar
entre Azure y Docker. Para revisar el estado use `docker compose ps` y
`docker compose logs db`.

Para pruebas locales sin Microsoft, se puede usar:

```env
DEV_AUTH_ENABLED=true
```

Con esta opción, el frontend muestra accesos locales de prueba y el backend acepta headers `X-Dev-Auth-*`.

> Advertencia: `DEV_AUTH_ENABLED=true` debe usarse exclusivamente en entornos de desarrollo local. No activar esta bandera en entornos públicos, de prueba integrados compartidos ni en producción.

### Comprobación antes de desplegar
Antes de cualquier despliegue a Azure o a un entorno de preproducción, verificar que:

```bash
# En el backend
cat backend/.env | grep DEV_AUTH_ENABLED
```

La salida debe ser:

```txt
DEV_AUTH_ENABLED=false
```

Si utiliza App Service o GitHub Actions, confirme también que la variable de entorno está definida como `false` en esos ámbitos.

## Configuración del frontend

Crear un archivo `.env` en `frontend/` con las variables de Vite usadas por la autenticación y la API.

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
`VITE_ENTRA_POST_LOGOUT_REDIRECT_URI` permite usar una URL distinta para local y nube sin cambiar código.

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

Para actualizar una base MVP 1 existente sin reconstruirla, seguir
[`database/migrations/README.md`](database/migrations/README.md) y ejecutar,
en orden y con una herramienta compatible con `GO`:

1. `001_mvp2_group_participants.sql`;
2. `002_mvp2_target_participants.sql`;
3. `003_open_use_frequency_scope.sql`;
4. `004_group_flow_completion.sql`;
5. `005_rut_integrity_and_admin_exemption.sql`;
6. `006_workshop_occurrences.sql`;
7. `007_repair_bootstrap_group_policy.sql`;
8. `008_personal_overlap_includes_participations.sql`.

Ejecutar los prechecks incluidos y detenerse ante cualquier resultado inesperado.
Ante un intento fallido sobre la única base, no ejecutar `drop.sql`, `schema.sql`
ni `seed.sql`. No ejecutar una migración `009`: no pertenece a la secuencia
aprobada.
4. Configurar `backend/.env`.
5. Levantar el backend y validar `/api/health`.

No usar scripts ni cadenas de conexion PostgreSQL para el entorno actual.

Antes de iniciar el backend, ejecutar `./scripts/configure-join-code-encryption.ps1`. Usar `-Rotate` para agregar una version activa conservando claves anteriores o `-Repair` para reemplazar una configuracion incompleta/invalida; ambos modos son incompatibles entre si. El script valida Git ignore y puntos de reanalisis, escribe atomicamente y crea backups `backend/.env.backup-*` sin mostrar secretos. Las variables resultantes son `JOIN_CODE_ENCRYPTION_KEYS` (`version:base64`, separadas por coma) y `JOIN_CODE_KEY_VERSION`.

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

La aplicación queda disponible normalmente en:

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
npm test
npm run build
```

## Rutas principales

Ruta publica:

```txt
GET /api/health
```

Rutas protegidas por token Bearer —listado orientativo—:

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

El contrato vigente de rutas, permisos y respuestas se mantiene en
[`docs/Propuesta nueva/02-arquitectura-y-contratos.md`](docs/Propuesta%20nueva/02-arquitectura-y-contratos.md)
y sus reglas de negocio en
[`docs/Propuesta nueva/06-flujos-y-reglas-de-negocio.md`](docs/Propuesta%20nueva/06-flujos-y-reglas-de-negocio.md).
En modo `DEV_AUTH_ENABLED=true`, las rutas protegidas también pueden probarse con los headers locales enviados por el frontend de desarrollo.

## Referencia histórica de demo online

Las siguientes direcciones correspondieron a la demo online inicial. No existe
revalidación online vigente posterior a los cambios recientes:

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
5. Guardar RUT una vez, verificar que permite avanzar y que luego se muestra solo lectura.
6. Crear una reserva desde Disponibilidad.
7. Seleccionar una actividad del catalogo o dejarla sin actividad especifica.
8. Revisar Mis Reservas, Detalle e Historial.
9. Cancelar una reserva propia.
10. Entrar como admin local y verificar acceso al panel administrador.
11. Confirmar que usuario normal no ve ni accede a rutas administrativas.

El modal de RUT aparece solamente cuando `/api/me` terminó de cargar, el usuario no es administrador y no posee un RUT válido. El login local no borra el RUT existente. Reenviar el mismo RUT es idempotente; intentar cambiarlo o reutilizar uno de otra cuenta responde `409`.

Un administrador sin RUT puede crear reservas normales o grupales e inscribirse en talleres, pero no puede confirmar como participante de una reserva ajena. En MVP 2, las inscripciones comparan solapes únicamente entre talleres activos con inscripciones `CONFIRMED`; no comparan un taller con una reserva personal ubicada en otro recurso. El alta y la baja se permiten mientras el taller esté activo, sin cutoff adicional; las inscripciones `CANCELLED` y los talleres inactivos no bloquean una nueva inscripción.

La cancelación de una reserva está permitida mientras la reserva no haya finalizado y el actor tenga permiso. Un cutoff configurable es una mejora futura, no un criterio de cierre de MVP 2.

La comparación horaria con la referencia online para `America/Santiago` permanece pendiente de nueva evidencia integrada.

El checklist histórico está en [Checklist demo MVP 1](docs/historico_y_checklists/12-checklist-demo-mvp1.md). La cobertura automatizada sigue siendo parcial y no reemplaza la validación integrada, manual ni online.

## Documentación relacionada

- [Entrada al canon documental](docs/Propuesta%20nueva/README.md).
- [Índice y gobierno documental](docs/Propuesta%20nueva/00-indice-y-gobierno-documental.md).
- [Resumen ejecutivo y estado](docs/Propuesta%20nueva/01-resumen-ejecutivo-y-estado.md).
- [Arquitectura y contratos](docs/Propuesta%20nueva/02-arquitectura-y-contratos.md).
- [Requisitos y trazabilidad](docs/Propuesta%20nueva/03-requisitos-casos-uso-y-trazabilidad.md).
- [Instalación, despliegue y recuperación](docs/Propuesta%20nueva/05-instalacion-despliegue-y-recuperacion.md).

## Seguridad

- No versionar archivos `.env`.
- No guardar passwords, cadenas de conexion reales ni secretos de Entra ID en documentos.
- Usar `backend/.env.example` como plantilla segura.
- Si una clave fue compartida fuera del entorno local seguro, rotarla antes de una entrega o despliegue.
