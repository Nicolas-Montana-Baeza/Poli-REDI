# Poli-REDI - Guia de despliegue y redeploy

## Objetivo

Esta guia deja un flujo estable para levantar Poli-REDI en local y redeployar la demo online en Azure sin depender de pasos improvisados.

Arquitectura actual:

- Frontend: Vue/Vite en Azure Static Web Apps.
- Backend: Go/Fiber en Azure App Service con contenedor Docker.
- Base de datos: Azure SQL Database.
- Autenticacion: Microsoft Entra ID.
- CI/CD frontend: GitHub Actions.

URLs actuales:

```txt
Frontend:
https://purple-ground-0205c9f10.7.azurestaticapps.net/

Backend:
https://poli-redi.azurewebsites.net

Health check:
https://poli-redi.azurewebsites.net/api/health
```

## 1. Regla rapida

Usar este criterio antes de redeployar:

| Cambio realizado | Accion necesaria |
| --- | --- |
| Solo frontend | Push a `main`; GitHub Actions redeploya Static Web Apps |
| Solo variables `VITE_*` | Reejecutar workflow o hacer commit vacio |
| Solo backend | Construir nueva imagen Docker, publicarla, actualizar tag en App Service y reiniciar |
| Backend y frontend | Probar local, push a `main`, esperar frontend, luego redeploy backend |
| Base de datos | Ejecutar `drop.sql`, `schema.sql`, `seed.sql` solo en ambiente de prueba o con respaldo |
| Datos demo de hoy | Ejecutar `database/seed_today_temp.sql` despues del seed normal |

## 2. Variables obligatorias

### 2.1 Backend local

Archivo: `backend/.env`

```env
PORT=3000
CORS_ALLOWED_ORIGINS=http://localhost:5173

DB_SERVER=poli-redi-server.database.windows.net
DB_PORT=1433
DB_NAME=poli-redi-database
DB_USER=poli-redi-admin
DB_PASSWORD=
DB_ENCRYPT=true
DB_TRUST_SERVER_CERTIFICATE=false
APP_TIMEZONE=America/Santiago

ENTRA_TENANT_ID=
ENTRA_API_CLIENT_ID=
ENTRA_ISSUER=

DEV_AUTH_ENABLED=true
```

Notas:

- `DB_PASSWORD` nunca debe quedar versionado.
- Para pruebas locales rapidas se puede usar `DEV_AUTH_ENABLED=true`.
- Para probar Microsoft Entra ID real en local, usar `DEV_AUTH_ENABLED=false`.
- Tambien se puede usar `AZURE_SQL_CONNECTION_STRING` en vez de las variables `DB_*`.

### 2.2 Frontend local

Archivo: `frontend/.env`

```env
VITE_API_BASE_URL=http://localhost:3000/api
VITE_API_TIMEOUT_MS=30000
VITE_APP_TIMEZONE=America/Santiago

VITE_ENTRA_TENANT_ID=
VITE_ENTRA_CLIENT_ID=
VITE_ENTRA_REDIRECT_URI=http://localhost:5173/auth/callback
VITE_ENTRA_POST_LOGOUT_REDIRECT_URI=http://localhost:5173/login
VITE_ENTRA_API_SCOPE=api://ENTRA_API_CLIENT_ID/access_as_user

VITE_DEV_AUTH_ENABLED=true
```

### 2.3 Backend Azure App Service

Ruta en Azure Portal:

```txt
App Service > poli-redi > Settings > Environment variables
```

Variables:

```env
PORT=3000
CORS_ALLOWED_ORIGINS=https://purple-ground-0205c9f10.7.azurestaticapps.net
APP_TIMEZONE=America/Santiago

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

DEV_AUTH_ENABLED=false
```

Reglas:

- En nube, `DEV_AUTH_ENABLED` debe ser `false`.
- `CORS_ALLOWED_ORIGINS` debe contener solo origenes necesarios. Para demo publica basta la URL de Static Web Apps.
- Si se necesita probar local contra backend online temporalmente, agregar `http://localhost:5173` solo durante la prueba y retirarlo despues.
- Reiniciar App Service despues de cambiar variables.

Contrato temporal:

- Azure SQL conserva hora institucional de muro en sus columnas `DATETIME2` de agenda.
- El backend interpreta esas columnas con `APP_TIMEZONE=America/Santiago` y serializa con offset.
- Una zona invalida impide iniciar el backend con un error de configuracion claro.
- Antes de validar `RES-009`, comprobar online que una reserva creada para una hora de Chile conserve la misma hora al recargar.

### 2.4 Frontend Azure Static Web Apps

Ruta en GitHub:

```txt
Repository > Settings > Secrets and variables > Actions > Variables
```

Variables:

```env
VITE_API_BASE_URL=https://poli-redi.azurewebsites.net/api
VITE_API_TIMEOUT_MS=30000
VITE_APP_TIMEZONE=America/Santiago

VITE_ENTRA_TENANT_ID=
VITE_ENTRA_CLIENT_ID=
VITE_ENTRA_REDIRECT_URI=https://purple-ground-0205c9f10.7.azurestaticapps.net/auth/callback
VITE_ENTRA_POST_LOGOUT_REDIRECT_URI=https://purple-ground-0205c9f10.7.azurestaticapps.net/login
VITE_ENTRA_API_SCOPE=api://ENTRA_API_CLIENT_ID/access_as_user

VITE_DEV_AUTH_ENABLED=false
```

Secret requerido:

```txt
AZURE_STATIC_WEB_APPS_API_TOKEN_PURPLE_GROUND_0205C9F10
```

Workflow:

```txt
.github/workflows/azure-static-web-apps-purple-ground-0205c9f10.yml
```

## 3. Ejecucion local

Backend:

```bash
cd backend
go run cmd/main.go
```

Validar:

```txt
http://localhost:3000/api/health
```

Frontend:

```bash
cd frontend
npm install
npm run dev
```

Abrir:

```txt
http://localhost:5173
```

Pruebas recomendadas antes de redeploy:

```bash
cd backend
go test ./...

cd ../frontend
npm run build
```

Estado 2026-07-14:

- `go test ./...` finaliza correctamente, pero aun no descubre archivos de prueba; no debe interpretarse como cobertura hasta cerrar `QA-001`.
- El frontend todavia no tiene `npm run test:run`; se agregara en `QA-002`.
- `npm run build` completa el build de produccion.

## 4. Base de datos

Scripts principales:

```txt
database/drop.sql
database/schema.sql
database/seed.sql
```

Flujo para base limpia de desarrollo:

1. Ejecutar `drop.sql`.
2. Ejecutar `schema.sql`.
3. Ejecutar `seed.sql`.
4. Opcional para pruebas de hoy: ejecutar `seed_today_temp.sql`.
5. Levantar backend y validar `/api/health`.
6. Abrir frontend y revisar disponibilidad/reservas.

Precauciones:

- No ejecutar `drop.sql` en una base con datos reales sin respaldo.
- `seed_today_temp.sql` es solo un overlay temporal de prueba.
- El seed base debe mantenerse estable.

## 5. Redeploy frontend

El frontend se publica automaticamente al hacer push a `main`.

Flujo:

```bash
cd frontend
npm run build

cd ..
git status
git add .
git commit -m "mensaje del cambio"
git push origin main
```

Luego revisar:

```txt
GitHub > Actions > Azure Static Web Apps CI/CD
```

Si solo cambiaron variables `VITE_*`, hay que forzar un nuevo build:

```bash
git commit --allow-empty -m "chore: trigger frontend redeploy"
git push origin main
```

Validar:

- La accion termina sin errores.
- La app carga en la URL de Static Web Apps.
- El frontend llama a `https://poli-redi.azurewebsites.net/api`, no a `localhost`.

## 6. Redeploy backend

El backend se publica como imagen Docker en Azure App Service.

### 6.1 Construir y publicar imagen

Desde `backend/`:

```bash
docker build -t TU_REGISTRY_O_USUARIO/poli-redi-api:TAG .
docker push TU_REGISTRY_O_USUARIO/poli-redi-api:TAG
```

Usar siempre un `TAG` nuevo. Evitar `latest` para no pelear con cache.

Ejemplos:

```txt
2026-07-14-01
main-042
mvp1-final-01
```

### 6.2 Actualizar Azure App Service

En Azure Portal:

```txt
App Service > poli-redi > Deployment Center
```

Actualizar:

- Registry.
- Image.
- Tag.

Luego reiniciar:

```txt
App Service > poli-redi > Overview > Restart
```

Validar:

```txt
https://poli-redi.azurewebsites.net/api/health
```

Respuesta esperada:

```json
{
  "message": "Poli-REDI API funcionando",
  "status": "ok"
}
```

## 7. Checklist antes de publicar

Antes de considerar un redeploy como usable para demo:

- `go test ./...` pasa.
- `npm run build` pasa.
- Una vez implementados `QA-001` y `QA-002`, ambas suites descubren y ejecutan pruebas reales.
- `DEV_AUTH_ENABLED=false` en App Service.
- `VITE_DEV_AUTH_ENABLED=false` en GitHub Actions Variables.
- `CORS_ALLOWED_ORIGINS` no queda abierto con `*`.
- `APP_TIMEZONE=America/Santiago` esta configurada una vez implementado `RES-009`.
- `DB_PASSWORD` vive solo en Azure App Service o `.env` local.
- El frontend online apunta al backend online.
- `/api/health` responde `ok`.
- Login Microsoft funciona.
- `/api/me` responde `200` para usuario valido.
- Usuario normal no ve rutas admin.
- Crear y cancelar reserva funciona en ambiente de prueba.
- Una reserva creada con hora de Chile conserva inicio/termino y categoria temporal en frontend online.

## 8. Problemas frecuentes

### 8.1 `ERR_CONNECTION_REFUSED`

El frontend intenta llamar a un backend apagado o a `localhost` desde un ambiente incorrecto.

Revisar:

- Backend local encendido si se usa `localhost`.
- `VITE_API_BASE_URL` en GitHub Actions si ocurre online.
- App Service iniciado si ocurre en Azure.

### 8.2 Error CORS

Revisar `CORS_ALLOWED_ORIGINS` en App Service.

Para demo online debe incluir:

```txt
https://purple-ground-0205c9f10.7.azurestaticapps.net
```

Reiniciar App Service despues del cambio.

### 8.3 `AADSTS50011`

La Redirect URI no esta registrada en Microsoft Entra ID.

Registrar:

```txt
http://localhost:5173/auth/callback
https://purple-ground-0205c9f10.7.azurestaticapps.net/auth/callback
```

### 8.4 Login correcto pero vuelve a `/login`

Revisar en Network:

```txt
GET /api/me
```

Interpretacion:

- `200`: revisar estado/router frontend.
- `401`: revisar scope, audience o issuer.
- `403`: usuario bloqueado o sin permisos.
- Error CORS: revisar `CORS_ALLOWED_ORIGINS`.

### 8.5 Cambios backend no aparecen

Probables causas:

- App Service sigue apuntando a un tag antiguo.
- Se uso `latest` y Azure mantuvo cache.
- Falta reiniciar App Service.

Solucion:

1. Construir imagen con tag nuevo.
2. Publicar imagen.
3. Actualizar tag en Deployment Center.
4. Reiniciar App Service.
5. Validar `/api/health`.

### 8.6 Pantalla 404 al abrir rutas internas

Revisar:

```txt
frontend/public/staticwebapp.config.json
```

Debe existir fallback a `index.html` para que Vue Router maneje rutas como `/login`, `/availability` o `/auth/callback`.

## 9. Seguridad operativa

- No subir `.env`.
- No copiar passwords en README, docs, issues ni capturas.
- Mantener secretos backend en Azure App Service.
- Mantener variables frontend no secretas en GitHub Actions Variables.
- Rotar la clave de Azure SQL si fue compartida fuera de un canal seguro.
- Mantener modo local (`DEV_AUTH_ENABLED`) desactivado en nube.
- Evitar CORS amplio en despliegue publico.

## 10. Estado para MVP 1

Con esta guia, el MVP 1 tiene un flujo documentado para:

- Levantar entorno local.
- Configurar variables seguras.
- Probar backend/frontend.
- Redeployar frontend.
- Redeployar backend Docker.
- Validar la demo online.

El endurecimiento productivo institucional queda como mejora posterior si Poli-REDI pasa de demo a operacion formal.
