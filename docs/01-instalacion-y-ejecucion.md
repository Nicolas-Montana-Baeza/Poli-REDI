# Poli-REDI - Instalacion y ejecucion local

## Objetivo del documento

Este documento explica como preparar y ejecutar Poli-REDI en ambiente local. Tambien registra configuraciones necesarias, comandos de verificacion y puntos pendientes de confirmar durante la revision tecnica del proyecto.

## Componentes del proyecto

Poli-REDI se divide en tres componentes principales:

- Frontend: aplicacion Vue 3 ubicada en `frontend/`.
- Backend: API Go/Fiber ubicada en `backend/`.
- Base de datos: scripts PostgreSQL ubicados en `database/`.

## Requisitos previos

Para ejecutar el proyecto localmente se requiere tener instalado:

- Node.js y npm para el frontend.
- Go para el backend.
- PostgreSQL para la base de datos.
- Git para control de versiones.

Versiones observadas en el proyecto:

- Frontend: Vite, Vue 3, Pinia y Vue Router.
- Backend: Go segun `backend/go.mod`.
- Base de datos: PostgreSQL.

## Estructura relevante

```txt
Poli-REDI/
  backend/
    cmd/main.go
    go.mod
    internal/
  database/
    schema.sql
    seed.sql
    drop.sql
  frontend/
    package.json
    vite.config.js
    src/
```

## Configuracion de base de datos

El backend requiere una variable de entorno llamada `DATABASE_URL`.

Ejemplo de formato:

```env
DATABASE_URL=postgres://usuario:password@localhost:5432/poli_redi?sslmode=disable
```

La conexion se configura en:

```txt
backend/internal/database/database.go
```

Si `DATABASE_URL` no esta definida, el backend finaliza la ejecucion con error.

## Scripts SQL disponibles

En la carpeta `database/` existen los siguientes archivos:

- `schema.sql`: crea extensiones, funciones, tablas, indices, restricciones, triggers y vistas.
- `seed.sql`: inserta datos iniciales de usuarios, recursos, actividades, reservas, participantes, infracciones, bloqueos y notificaciones.
- `drop.sql`: elimina objetos de base de datos.
- `schema_0.1.sql`: version alternativa o anterior del esquema.

## Preparar base de datos local

Pasos sugeridos:

1. Crear una base de datos PostgreSQL para el proyecto.
2. Ejecutar `database/schema.sql`.
3. Ejecutar `database/seed.sql` si se requieren datos de prueba.
4. Configurar `DATABASE_URL` en el entorno del backend.

Ejemplo conceptual:

```bash
psql -d poli_redi -f database/schema.sql
psql -d poli_redi -f database/seed.sql
```

Nota: el nombre de la base de datos, usuario y password dependen del entorno local de cada integrante.

## Configuracion del backend

El backend carga variables de entorno desde un archivo `.env` si existe. Si no existe, usa las variables definidas en el sistema.

Archivo principal:

```txt
backend/cmd/main.go
```

Variables identificadas:

```env
PORT=3000
DATABASE_URL=postgres://usuario:password@localhost:5432/poli_redi?sslmode=disable
ENTRA_TENANT_ID=...
ENTRA_API_CLIENT_ID=...
ENTRA_ISSUER=...
```

`PORT` es opcional. Si no se define, el backend usa `3000`.

Las variables `ENTRA_*` son necesarias para validar tokens de Microsoft Entra ID en rutas protegidas.

## Ejecutar backend

Desde la carpeta `backend/`:

```bash
go mod download
go run ./cmd
```

Si todo esta configurado correctamente, la API deberia quedar disponible en:

```txt
http://localhost:3000
```

## Verificar backend

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

## Rutas backend identificadas

Ruta publica:

```txt
GET /api/health
```

Rutas protegidas por autenticacion:

```txt
GET /api/me
GET /api/resources
GET /api/reservations
POST /api/reservations
PATCH /api/reservations/cancel
```

Para probar rutas protegidas se requiere un token Bearer valido.

## Configuracion del frontend

El frontend esta ubicado en:

```txt
frontend/
```

Scripts disponibles en `frontend/package.json`:

```json
{
  "dev": "vite",
  "build": "vite build",
  "preview": "vite preview"
}
```

El cliente HTTP esta configurado en:

```txt
frontend/src/services/api.js
```

La API base actual es:

```txt
http://localhost:3000/api
```

## Variables frontend

La configuracion de autenticacion usa variables de entorno de Vite:

```env
VITE_ENTRA_TENANT_ID=...
VITE_ENTRA_CLIENT_ID=...
VITE_ENTRA_REDIRECT_URI=http://localhost:5173/auth/callback
VITE_ENTRA_API_SCOPE=...
```

Estas variables se consumen en:

```txt
frontend/src/auth/msalConfig.js
```

## Ejecutar frontend

Desde la carpeta `frontend/`:

```bash
npm install
npm run dev
```

Por defecto, Vite suele levantar la aplicacion en:

```txt
http://localhost:5173
```

## Compilar frontend

Desde la carpeta `frontend/`:

```bash
npm run build
```

Este comando sirve para validar que la aplicacion compile correctamente.

## Vista previa de produccion

Luego de compilar:

```bash
npm run preview
```

## Flujo de ejecucion local recomendado

1. Levantar PostgreSQL.
2. Crear la base de datos local.
3. Ejecutar `database/schema.sql`.
4. Ejecutar `database/seed.sql` si se quieren datos de prueba.
5. Crear o configurar variables de entorno del backend.
6. Ejecutar backend en `http://localhost:3000`.
7. Crear o configurar variables de entorno del frontend.
8. Ejecutar frontend en `http://localhost:5173`.
9. Verificar `GET /api/health`.
10. Probar inicio de sesion y pantallas principales.

## Puntos pendientes de confirmar

- Version exacta de Go instalada en el equipo de desarrollo.
- Version exacta de Node.js recomendada para el proyecto.
- Nombre oficial de la base de datos local.
- Si existe un archivo `.env.example` pendiente de crear.
- Si las variables de Microsoft Entra ID son definitivas o de prueba.
- Si el frontend debe permitir cambiar la URL de la API por variable de entorno.
- Si las rutas protegidas tienen algun modo de prueba local sin autenticacion real.

## Observaciones de revision

- El backend depende obligatoriamente de PostgreSQL para iniciar.
- El endpoint `/api/health` tambien requiere que el backend haya conectado correctamente a la base de datos, porque la conexion se realiza antes de registrar y servir rutas.
- El frontend usa una URL de API fija: `http://localhost:3000/api`.
- La autenticacion esta integrada con Microsoft Entra ID mediante MSAL en frontend y validacion JWT en backend.
- Para facilitar futuros desarrollos, conviene crear archivos `.env.example` para backend y frontend.

## Siguiente documento sugerido

`docs/02-arquitectura.md`

Ese documento deberia explicar como se comunican frontend, backend, autenticacion y base de datos.
