# Instalación, despliegue y recuperación

**Audiencia:** desarrollo, QA, DevOps y operación

**Propósito:** reproducir el entorno, evolucionar la base única y recuperar una publicación

**Estado:** guía operativa canónica, corte 2026-08-11

**Fuente:** configuración del repositorio y `Propuesta nueva/05`

## Resumen

El sistema usa Vue, Go, Microsoft Entra ID y Azure SQL. La instalación local puede emplear autenticación de desarrollo, pero Azure siempre debe usar Entra y secretos protegidos. Una base existente evoluciona en orden con migraciones `001`–`008`; `009` está excluida por no estar aprobada.

Ningún despliegue se considera recuperable sin versión identificable, respaldo o estrategia equivalente, postcheck y smoke test.

## Prerrequisitos

- Node.js y npm compatibles con `frontend/package.json`.
- Go compatible con `backend/go.mod`.
- Docker para la ejecución local opcional.
- Acceso autorizado a Azure SQL y Microsoft Entra ID para el ambiente online.
- Cliente T-SQL que interprete separadores `GO`.

## Configuración segura

Crear archivos locales desde las plantillas del repositorio. No versionar valores reales.

| Backend | Uso |
|---|---|
| `PORT`, `APP_TIMEZONE` | Servicio y zona de negocio (`America/Santiago`). |
| `CORS_ALLOWED_ORIGINS` | Orígenes explícitamente autorizados. |
| `DB_SERVER`, `DB_PORT`, `DB_NAME` | Destino de base de datos. |
| `DB_USER`, `DB_PASSWORD` | Credenciales protegidas. |
| `DB_ENCRYPT`, `DB_TRUST_SERVER_CERTIFICATE` | TLS; en Azure se exige cifrado y no confiar certificados arbitrarios. |
| `ENTRA_TENANT_ID`, `ENTRA_API_CLIENT_ID`, `ENTRA_ISSUER` | Validación de tokens. |
| `DEV_AUTH_ENABLED` | Solo desarrollo local; siempre `false` online. |

| Frontend | Uso |
|---|---|
| `VITE_API_BASE_URL`, `VITE_API_TIMEOUT_MS` | API y timeout. |
| `VITE_APP_TIMEZONE` | Zona presentada al usuario. |
| `VITE_ENTRA_TENANT_ID`, `VITE_ENTRA_CLIENT_ID` | Aplicación SPA. |
| `VITE_ENTRA_REDIRECT_URI`, `VITE_ENTRA_POST_LOGOUT_REDIRECT_URI` | Redirecciones registradas. |
| `VITE_ENTRA_API_SCOPE` | Scope de la API. |
| `VITE_DEV_AUTH_ENABLED` | Solo desarrollo local; siempre `false` online. |

## Ejecución local

### Backend

```powershell
cd backend
go mod download
go run ./cmd
```

Comprobar `GET http://localhost:3000/api/health`.

### Frontend

```powershell
cd frontend
npm install
npm run dev
```

Abrir `http://localhost:5173`. Backend y frontend deben coincidir en modo de autenticación.

### Contenedores

Si se usa Docker, validar puertos, variables y persistencia antes de atribuir un error a la aplicación. Actualizar una imagen no modifica automáticamente Azure SQL ni corrige CORS o cuotas del servicio.

## Base de datos única

Para una base nueva se usa el paquete de instalación vigente. Para una base existente se ejecutan únicamente migraciones incrementales, nunca un script destructivo.

| Orden | Migración | Propósito resumido |
|---:|---|---|
| 001 | `001_mvp2_group_reservations.sql` | Estructura inicial de reserva grupal. |
| 002 | `002_mvp2_group_reservations_followup.sql` | Ajustes de seguimiento. |
| 003 | `003_mvp2_admin_resources.sql` | Base para recursos administrativos. |
| 004 | `004_mvp2_group_join_code_crypto.sql` | Código de unión protegido. |
| 005 | `005_mvp2_group_join_code_crypto_hotfix.sql` | Corrección compatible del cifrado. |
| 006 | `006_mvp2_group_reservation_policy.sql` | Política y snapshots grupales. |
| 007 | `007_user_schedule_overlap_integrity.sql` | Integridad de agenda personal. |
| 008 | `008_workshop_enrollment_history.sql` | Episodios e historial de talleres. |

`009_user_schedule_overlap_integrity_v2.sql` es una propuesta no aprobada. No se ejecuta, no se incluye en una instalación y no se usa para declarar cerrado MVP 2.

### Procedimiento de migración

1. Identificar base, esquema, versión de aplicación y responsable por rol.
2. Crear respaldo recuperable y probar acceso al mecanismo de restauración.
3. Ejecutar el precheck de la migración en una copia recuperable.
4. Resolver cualquier definición inesperada; no forzar DDL ni crear una segunda base.
5. Ejecutar una migración por vez en orden.
6. Ejecutar postcheck funcional y de integridad.
7. Repetir la migración para demostrar idempotencia cuando el script la prometa.
8. Conservar resultado, fecha, ambiente y versión como evidencia.
9. Repetir el procedimiento controlado en Azure SQL.

La estrategia y las condiciones de reversión se detallan en [04-base-de-datos-y-migraciones.md](04-base-de-datos-y-migraciones.md) y [ADR-002](decisiones/ADR-002-evolucion-base-unica.md).

## Verificación antes de publicar

```powershell
cd backend
go test ./... -count=1
```

```powershell
cd frontend
npm test
npm run build
```

Además:

- revisar CORS y redirect URIs;
- comprobar que autenticación de desarrollo esté desactivada;
- revisar secretos, scopes y certificados;
- verificar inventario de migraciones aplicado;
- ejecutar flujos críticos del MVP correspondiente;
- registrar artefactos, commit y resultados.

## Despliegue Azure

| Componente | Publicación | Verificación |
|---|---|---|
| Frontend | Artefacto/commit identificable y variables `VITE_*` de build | Carga, rutas SPA, login y consumo API. |
| Backend | Imagen con tag inmutable; evitar depender de `latest` | Health, logs seguros, CORS y autorización. |
| Entra ID | Redirect URIs, scope, tenant y audiencia correctos | Usuario normal, admin, logout y rechazo. |
| Azure SQL | Backup, migraciones aprobadas y postcheck | Esquema, integridad, concurrencia y recuperación. |

No reutilizar una evidencia histórica de despliegue como prueba de la versión actual. Las cuotas gratuitas agotadas o servicios detenidos deben registrarse como indisponibilidad del ambiente, no como defecto funcional.

## Recuperación

### Frontend

- volver al artefacto o commit conocido;
- restaurar las variables del build;
- verificar navegación directa y autenticación.

### Backend

- restaurar el tag anterior;
- reiniciar el servicio;
- comprobar health, CORS y logs sin secretos.

### Base de datos

- detener la publicación y abrir una sesión nueva;
- identificar objetos y transacciones afectadas;
- aplicar la reversión aprobada o restaurar el respaldo;
- no ejecutar `drop.sql` ni reconstruir la base única;
- validar aplicación y datos con la versión recuperada.

## Smoke posterior

1. Health y login Entra.
2. Carga de usuario, rol y recursos.
3. Disponibilidad sanitizada.
4. Reserva válida, conflicto y cancelación permitida.
5. Código grupal, progreso y privacidad.
6. Inscripción y desinscripción de taller.
7. Rechazo de rutas administrativas a usuario normal.
8. Navegación SPA, CORS y ausencia de errores internos visibles.

Registrar el resultado en [07-calidad-y-evidencia.md](07-calidad-y-evidencia.md) y [09-checklist-cierre.md](09-checklist-cierre.md). El corte operativo de Google Calendar está en [anexos/operacion/corte-google-calendar.md](anexos/operacion/corte-google-calendar.md).
