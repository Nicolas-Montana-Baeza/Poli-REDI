# Instalación, despliegue y recuperación de Poli-REDI

**Estado:** GUÍA OPERATIVA CANÓNICA

## 1. Prerrequisitos

- Node.js y npm compatibles con el frontend.
- Go compatible con `backend/go.mod`.
- Acceso a Azure SQL Database.
- Aplicaciones Microsoft Entra ID configuradas para SPA y API.
- Herramienta T-SQL compatible con separadores `GO` para migraciones.

## 2. Variables del backend

Crear `backend/.env` desde la plantilla segura:

```env
PORT=3000
CORS_ALLOWED_ORIGINS=http://localhost:5173
APP_TIMEZONE=America/Santiago

DB_SERVER=<servidor>.database.windows.net
DB_PORT=1433
DB_NAME=<base>
DB_USER=<usuario>
DB_PASSWORD=
DB_ENCRYPT=true
DB_TRUST_SERVER_CERTIFICATE=false

ENTRA_TENANT_ID=
ENTRA_API_CLIENT_ID=
ENTRA_ISSUER=
DEV_AUTH_ENABLED=false
```

También puede usarse `AZURE_SQL_CONNECTION_STRING`. Las contraseñas y claves no se versionan.

## 3. Variables del frontend

```env
VITE_API_BASE_URL=http://localhost:3000/api
VITE_API_TIMEOUT_MS=30000
VITE_APP_TIMEZONE=America/Santiago

VITE_ENTRA_TENANT_ID=
VITE_ENTRA_CLIENT_ID=
VITE_ENTRA_REDIRECT_URI=http://localhost:5173/auth/callback
VITE_ENTRA_POST_LOGOUT_REDIRECT_URI=http://localhost:5173/login
VITE_ENTRA_API_SCOPE=
VITE_DEV_AUTH_ENABLED=false
```

## 4. Ejecución local

### Backend

```bash
cd backend
go mod download
go run ./cmd
```

Validar:

```text
GET http://localhost:3000/api/health
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Abrir `http://localhost:5173`.

## 5. Modo de desarrollo

El modo local puede habilitarse solo en equipos de desarrollo. Backend y frontend deben estar alineados. Nunca publicar `DEV_AUTH_ENABLED=true` ni `VITE_DEV_AUTH_ENABLED=true`.

## 6. Preparación de base de datos

- **Base descartable:** seguir instalación limpia del paquete de base de datos.
- **Base existente:** ejecutar migraciones incrementalmente; no usar scripts destructivos.
- Configurar llaves de join code mediante el script del repositorio y conservar versiones anteriores durante rotación.

## 7. Validaciones antes de publicar

```bash
# Backend
go test ./... -count=1

# Frontend
npm test
npm run build
```

Además:

- revisar CORS;
- comprobar redirect URIs de Entra;
- confirmar variables de ambiente;
- verificar que no existan secretos en Git;
- ejecutar checklist manual MVP aplicable;
- conservar evidencia de migraciones y recuperación.

## 8. Estrategia de redeploy

| Cambio | Acción |
|---|---|
| Solo frontend | Push a la rama de despliegue y verificar workflow de Static Web Apps. |
| Variables `VITE_*` | Reejecutar build/deploy; se incorporan en compilación. |
| Solo backend | Construir imagen, publicar tag inmutable, actualizar App Service y reiniciar. |
| Backend y frontend | Probar ambos localmente; desplegar con orden y validar compatibilidad. |
| Base de datos | Ejecutar migración aprobada con backup, postcheck e idempotencia. |

## 9. Reglas de nube

- `DEV_AUTH_ENABLED=false`.
- CORS limitado a orígenes necesarios.
- `DB_TRUST_SERVER_CERTIFICATE=false`.
- Usar secretos del servicio o variables protegidas.
- No exponer cadenas de conexión en logs.
- Usar tags de imagen identificables; evitar depender de `latest` para recuperación.

## 10. Recuperación de despliegue

### Frontend

- volver a un commit o artefacto conocido;
- revisar variables `VITE_*` del build;
- comprobar rutas SPA.

### Backend

- restaurar el tag anterior del contenedor;
- reiniciar App Service;
- verificar `/api/health` y logs sin secretos.

### Base de datos

- detener publicación;
- abrir sesión nueva;
- comprobar estado transaccional;
- inspeccionar objetos;
- restaurar backup cuando no pueda demostrarse compatibilidad;
- no aplicar `drop.sql` en la base única.

## 11. Verificación posterior

1. Health público.
2. Login Entra real.
3. Carga de usuario y recursos.
4. Disponibilidad sanitizada.
5. Reserva y conflicto.
6. Flujo grupal por código.
7. Inscripción y desinscripción de taller.
8. Acceso administrativo y rechazo a usuario normal.
9. CORS y navegación directa de rutas SPA.
10. Revisión de errores y privacidad.

## 12. Estado pendiente

La documentación fuente no demuestra que este flujo se haya ejecutado online después del incremento del 2026-08-04. Debe registrarse una nueva evidencia antes de afirmar disponibilidad actual.
