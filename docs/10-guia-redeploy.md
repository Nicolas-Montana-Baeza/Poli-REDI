# Poli-REDI - Guia de ejecucion y redeploy

## Objetivo

Esta guia resume como volver a ejecutar Poli-REDI en local y como redeployar la demo online en Azure.

La configuracion actual usa:

- Frontend: Vue/Vite en Azure Static Web Apps.
- Backend: Go/Fiber en Azure App Service con Docker.
- Base de datos: Azure SQL Database.
- Autenticacion: Microsoft Entra ID.
- CI/CD frontend: GitHub Actions.

## URLs actuales

```txt
Frontend Azure:
https://purple-ground-0205c9f10.7.azurestaticapps.net/

Backend Azure:
https://poli-redi.azurewebsites.net

Health check:
https://poli-redi.azurewebsites.net/api/health
```

## 1. Ejecutar en local

### 1.1 Requisitos

- Node.js y npm.
- Go compatible con `backend/go.mod`.
- Docker, solo si se quiere probar imagen local.
- Acceso a Azure SQL Database.
- App registrations de Microsoft Entra ID para frontend y API.

### 1.2 Variables del backend local

Crear o revisar `backend/.env`.

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

ENTRA_TENANT_ID=
ENTRA_API_CLIENT_ID=
ENTRA_ISSUER=

DEV_AUTH_ENABLED=true
```

Notas:

- `DB_PASSWORD` no debe subirse al repositorio.
- Para pruebas locales rapidas se puede usar `DEV_AUTH_ENABLED=true`.
- Para probar Microsoft Entra ID real en local, usar `DEV_AUTH_ENABLED=false`.

### 1.3 Variables del frontend local

Crear o revisar `frontend/.env`.

```env
VITE_API_BASE_URL=http://localhost:3000/api
VITE_API_TIMEOUT_MS=30000

VITE_ENTRA_TENANT_ID=
VITE_ENTRA_CLIENT_ID=
VITE_ENTRA_REDIRECT_URI=http://localhost:5173/auth/callback
VITE_ENTRA_POST_LOGOUT_REDIRECT_URI=http://localhost:5173/login
VITE_ENTRA_API_SCOPE=api://ENTRA_API_CLIENT_ID/access_as_user

VITE_DEV_AUTH_ENABLED=true
```

### 1.4 Ejecutar backend local

```bash
cd backend
go run ./cmd
```

Validar:

```txt
http://localhost:3000/api/health
```

### 1.5 Ejecutar frontend local

```bash
cd frontend
npm install
npm run dev
```

Abrir:

```txt
http://localhost:5173
```

### 1.6 Validaciones locales recomendadas

```bash
cd backend
go test ./...

cd ../frontend
npm run build
```

Flujos minimos:

1. Entrar a la aplicacion.
2. Validar `/api/me`.
3. Abrir Disponibilidad.
4. Crear una reserva.
5. Ver Mis Reservas.
6. Cancelar una reserva.
7. Ver Dashboard y panel admin con usuario administrador.

## 2. Variables de nube

### 2.1 Azure App Service backend

En Azure Portal:

```txt
App Service poli-redi > Settings > Environment variables
```

Variables esperadas:

```env
PORT=3000
CORS_ALLOWED_ORIGINS=https://purple-ground-0205c9f10.7.azurestaticapps.net,http://localhost:5173

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

Notas:

- Si App Service usa un puerto distinto por configuracion de contenedor, mantener el mismo puerto que expone la imagen.
- En la configuracion actual el Dockerfile expone `3000`.
- Reiniciar App Service despues de cambiar variables.

### 2.2 GitHub Actions variables del frontend

En GitHub:

```txt
Repository > Settings > Secrets and variables > Actions > Variables
```

Variables esperadas:

```env
VITE_API_BASE_URL=https://poli-redi.azurewebsites.net/api
VITE_API_TIMEOUT_MS=30000

VITE_ENTRA_TENANT_ID=
VITE_ENTRA_CLIENT_ID=
VITE_ENTRA_REDIRECT_URI=https://purple-ground-0205c9f10.7.azurestaticapps.net/auth/callback
VITE_ENTRA_POST_LOGOUT_REDIRECT_URI=https://purple-ground-0205c9f10.7.azurestaticapps.net/login
VITE_ENTRA_API_SCOPE=api://ENTRA_API_CLIENT_ID/access_as_user

VITE_DEV_AUTH_ENABLED=false
```

El workflow que consume estas variables es:

```txt
.github/workflows/azure-static-web-apps-purple-ground-0205c9f10.yml
```

### 2.3 GitHub Actions secrets

El secret de Azure Static Web Apps debe existir:

```txt
AZURE_STATIC_WEB_APPS_API_TOKEN_PURPLE_GROUND_0205C9F10
```

Azure lo crea al conectar la Static Web App con GitHub.

## 3. Microsoft Entra ID

### 3.1 App frontend

En el tenant correcto, revisar:

```txt
Microsoft Entra ID > App registrations > Poli-REDI Frontend > Authentication
```

Redirect URIs esperadas para SPA:

```txt
http://localhost:5173/auth/callback
https://purple-ground-0205c9f10.7.azurestaticapps.net/auth/callback
```

### 3.2 App API

En el tenant correcto, revisar:

```txt
Microsoft Entra ID > App registrations > Poli-REDI API > Expose an API
```

Debe existir un scope equivalente a:

```txt
api://ENTRA_API_CLIENT_ID/access_as_user
```

El backend usa `ENTRA_API_CLIENT_ID` para validar el `aud` del token.

## 4. Redeploy del frontend

El frontend se redeploya con GitHub Actions al hacer push a `main`.

Flujo normal:

```bash
git status
git add .
git commit -m "mensaje del cambio"
git push origin main
```

Luego revisar:

```txt
GitHub > Actions > Azure Static Web Apps CI/CD
```

El workflow debe:

- Instalar dependencias del frontend.
- Ejecutar `npm run build`.
- Publicar `frontend/dist`.
- Incluir `frontend/public/staticwebapp.config.json` en el build.

### 4.1 Rebuild sin cambios funcionales

Si solo se cambiaron variables `VITE_*`, se debe volver a correr el workflow. Se puede usar un commit vacio:

```bash
git commit --allow-empty -m "chore: trigger frontend redeploy"
git push origin main
```

## 5. Redeploy del backend con Docker

El backend se despliega como contenedor en Azure App Service.

### 5.1 Construir imagen

Desde la raiz del proyecto o desde `backend/`, construir usando el Dockerfile del backend.

Ejemplo con Docker Hub:

```bash
cd backend
docker build -t TU_USUARIO/poli-redi-api:TAG .
docker push TU_USUARIO/poli-redi-api:TAG
```

Ejemplo con Azure Container Registry:

```bash
cd backend
docker build -t TU_REGISTRY.azurecr.io/poli-redi-api:TAG .
docker push TU_REGISTRY.azurecr.io/poli-redi-api:TAG
```

Usar un `TAG` nuevo por despliegue evita que Azure conserve una imagen antigua en cache.

Ejemplos de tags:

```txt
v1
v2
2026-07-06
main-001
```

### 5.2 Apuntar App Service a la imagen

En Azure Portal:

```txt
App Service poli-redi > Deployment Center
```

Revisar:

- Registry.
- Image.
- Tag.

Guardar cambios y reiniciar:

```txt
App Service poli-redi > Overview > Restart
```

### 5.3 Validar backend

Abrir:

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

## 6. Redeploy completo recomendado

Cuando cambia frontend y backend:

1. Ejecutar pruebas locales.
2. Hacer commit de cambios.
3. Push a `main`.
4. Esperar deploy del frontend en GitHub Actions.
5. Construir nueva imagen Docker del backend.
6. Publicar imagen en el registry.
7. Actualizar tag en Azure App Service.
8. Reiniciar App Service.
9. Validar `/api/health`.
10. Validar frontend online.

Comandos base:

```bash
cd backend
go test ./...

cd ../frontend
npm run build

cd ..
git status
git add .
git commit -m "mensaje del cambio"
git push origin main
```

Luego:

```bash
cd backend
docker build -t TU_REGISTRY_O_USUARIO/poli-redi-api:TAG .
docker push TU_REGISTRY_O_USUARIO/poli-redi-api:TAG
```

## 7. Checklist de demo online

Validar:

- `https://poli-redi.azurewebsites.net/api/health` responde `ok`.
- `https://purple-ground-0205c9f10.7.azurestaticapps.net/` carga la app.
- Login Microsoft redirige correctamente.
- `/api/me` responde `200`.
- Dashboard carga sin errores criticos.
- Disponibilidad carga recursos y reservas.
- Crear reserva funciona.
- Mis Reservas muestra la reserva.
- Cancelar reserva funciona.
- Usuario normal no entra a rutas admin.
- Usuario admin ve panel administrador.

## 8. Problemas frecuentes

### 8.1 Pantalla 404 al abrir `/login` o `/auth/callback`

Revisar que exista:

```txt
frontend/public/staticwebapp.config.json
```

Debe contener fallback a `/index.html`.

### 8.2 Error `AADSTS50011`

La Redirect URI no esta registrada en Entra ID.

Agregar en la app frontend:

```txt
https://purple-ground-0205c9f10.7.azurestaticapps.net/auth/callback
```

### 8.3 Error `Invalid URL` en frontend

Faltan variables `VITE_*` durante el build.

Revisar GitHub Actions Variables y redeployar frontend.

### 8.4 Error CORS

Revisar en Azure App Service:

```env
CORS_ALLOWED_ORIGINS=https://purple-ground-0205c9f10.7.azurestaticapps.net,http://localhost:5173
```

Reiniciar App Service despues del cambio.

### 8.5 Login correcto pero vuelve a `/login`

Revisar en Network:

```txt
GET https://poli-redi.azurewebsites.net/api/me
```

Interpretacion:

- `200`: token y backend funcionan; revisar router/estado frontend.
- `401`: revisar scope, audience o issuer.
- `403`: usuario bloqueado o token sin email usable.
- Error CORS: revisar `CORS_ALLOWED_ORIGINS`.

### 8.6 Frontend apunta a `localhost`

El frontend fue compilado sin `VITE_API_BASE_URL` de nube.

Actualizar GitHub Actions Variables y correr el workflow de nuevo.

### 8.7 Backend no refleja cambios recientes

Si se usa Docker con tag `latest`, Azure puede conservar cache.

Solucion recomendada:

- Publicar imagen con tag nuevo.
- Actualizar el tag en Deployment Center.
- Reiniciar App Service.

## 9. Seguridad operativa

- No subir `.env`.
- No guardar passwords en README ni docs.
- Mantener secretos en Azure App Service o GitHub Secrets.
- Mantener variables no secretas del frontend en GitHub Actions Variables.
- Rotar la clave de Azure SQL si fue compartida fuera del entorno seguro.
- Usar `DEV_AUTH_ENABLED=false` y `VITE_DEV_AUTH_ENABLED=false` en nube.

## 10. Estado actual

Con esta configuracion, Poli-REDI queda en estado de demo online funcional:

- MVP 1 cerrado funcionalmente y desplegado.
- Parte importante de MVP 2 ya demostrable.
- Despliegue productivo institucional formal queda como endurecimiento futuro.
