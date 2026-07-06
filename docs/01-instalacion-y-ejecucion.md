# Poli-REDI - Instalacion y ejecucion local

## Objetivo del documento

Este documento explica como preparar y ejecutar Poli-REDI en ambiente local. Tambien registra configuraciones necesarias, comandos de verificacion y puntos pendientes de confirmar durante la revision tecnica del proyecto.

## Componentes del proyecto

Poli-REDI se divide en tres componentes principales:

- Frontend: aplicacion Vue 3 ubicada en `frontend/`.
- Backend: API Go/Fiber ubicada en `backend/`.
- Base de datos: Azure SQL Database. Los scripts actuales en `database/` pertenecen a la implementacion anterior basada en PostgreSQL.

## Requisitos previos

Para ejecutar el proyecto localmente se requiere tener instalado:

- Node.js y npm para el frontend.
- Go para el backend.
- Azure SQL Database.
- Git para control de versiones.

Versiones observadas en el proyecto:

- Frontend: Vite, Vue 3, Pinia y Vue Router.
- Backend: Go segun `backend/go.mod`.
- Base de datos: Azure SQL Database.

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

## Configuracion de base de datos heredada

El backend actual todavia requiere una variable de entorno llamada `DATABASE_URL`, porque la implementacion existente fue construida sobre PostgreSQL. Esta configuracion queda marcada como heredada mientras se migra a Azure SQL Database.

Ejemplo de formato:

```env
# DATABASE_URL=postgres://USUARIO:<PASSWORD>@localhost:5432/poli_redi?sslmode=disable
```

La conexion se configura en:

```txt
backend/internal/database/database.go
```

Si `DATABASE_URL` no esta definida, el backend finaliza la ejecucion con error.

## Scripts SQL heredados

En la carpeta `database/` existen archivos SQL de la implementacion anterior. Se mantienen como referencia del modelo de datos, pero no representan la decision futura de base de datos:

- `schema.sql`: crea extensiones, funciones, tablas, indices, restricciones, triggers y vistas.
- `seed.sql`: inserta datos iniciales de usuarios, recursos, actividades, reservas, participantes, infracciones, bloqueos y notificaciones.
- `drop.sql`: elimina objetos de base de datos.
- `schema_0.1.sql`: version alternativa o anterior del esquema.

## Preparar base de datos local heredada

Pasos sugeridos:

1. Crear una base de datos PostgreSQL solo si se desea probar la implementacion heredada.
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
# DATABASE_URL=postgres://USUARIO:<PASSWORD>@localhost:5432/poli_redi?sslmode=disable
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
VITE_ENTRA_POST_LOGOUT_REDIRECT_URI=http://localhost:5173/login
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

1. Usar Azure SQL Database como base de datos objetivo. Si se prueba la implementacion heredada, levantar PostgreSQL.
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

- El backend actual depende de la configuracion heredada de PostgreSQL para iniciar; esto debe cambiar durante la migracion a Azure SQL Database.
- El endpoint `/api/health` tambien requiere que el backend haya conectado correctamente a la base de datos, porque la conexion se realiza antes de registrar y servir rutas.
- El frontend usa una URL de API fija: `http://localhost:3000/api`.
- La autenticacion esta integrada con Microsoft Entra ID mediante MSAL en frontend y validacion JWT en backend.
- Para facilitar futuros desarrollos, conviene crear archivos `.env.example` para backend y frontend.

## Siguiente documento sugerido

`docs/02-arquitectura.md`

Ese documento deberia explicar como se comunican frontend, backend, autenticacion y base de datos.

## Conexion con base de datos Azure heredada

Se habia creado una base de datos PostgreSQL en Azure, pero el proyecto ya no usara PostgreSQL. Esta seccion queda como registro historico y no debe tomarse como guia final de implementacion.

No se deben guardar credenciales reales dentro del repositorio. La URL de conexion debe quedar en un archivo `.env` local o en variables de entorno del servicio donde se despliegue el backend.

Formato recomendado:

```env
# DATABASE_URL=postgres://USUARIO:<PASSWORD>@HOST:5432/NOMBRE_DB?sslmode=require
```

En Azure normalmente es importante usar SSL, por eso se recomienda:

```txt
sslmode=require
```

Ejemplo con placeholders:

```env
# DATABASE_URL=postgres://USUARIO:<PASSWORD>@HOST:5432/NOMBRE_DB?sslmode=require
```

Checklist para validar la conexion:

- [ ] Confirmar host del servidor PostgreSQL en Azure.
- [ ] Confirmar nombre de la base de datos.
- [ ] Confirmar usuario administrador o usuario de aplicacion.
- [ ] Confirmar password.
- [ ] Confirmar que Azure permite conexiones desde la IP local de desarrollo.
- [ ] Confirmar que el puerto `5432` esta disponible.
- [ ] Confirmar uso de `sslmode=require`.
- [ ] Ejecutar `database/schema.sql` sobre la base Azure.
- [ ] Ejecutar `database/seed.sql` solo si se quieren datos de prueba.
- [ ] Levantar backend y probar `GET /api/health`.

Pendiente recomendado:

Crear un archivo `.env.example` para documentar variables sin exponer secretos reales.

Ejemplo:

```env
PORT=3000
# DATABASE_URL=postgres://USUARIO:<PASSWORD>@HOST:5432/NOMBRE_DB?sslmode=require
ENTRA_TENANT_ID=
ENTRA_API_CLIENT_ID=
ENTRA_ISSUER=
```

## Cambio de decision: no se usara PostgreSQL

Se definio que Poli-REDI ya no usara PostgreSQL. Por lo tanto, las instrucciones anteriores relacionadas con `DATABASE_URL`, `pgx`, `psql`, `schema.sql`, `seed.sql` y Azure PostgreSQL quedan como referencia historica de la implementacion existente.

Antes de continuar con nuevas funcionalidades, se debe actualizar la implementacion para Azure SQL Database y ajustar esta guia.

Checklist de migracion documental y tecnica:

- [x] Definir nueva tecnologia de base de datos: Azure SQL Database.
- [x] Definir si la base sera relacional, documental o administrada como servicio: relacional administrada en Azure.
- [ ] Actualizar variables de entorno requeridas.
- [ ] Reemplazar instrucciones de conexion local.
- [ ] Revisar `backend/internal/database/database.go`.
- [ ] Revisar repositorios en `backend/internal/repositories/`.
- [ ] Decidir si `database/schema.sql` y `database/seed.sql` se eliminan, migran o quedan como referencia.
- [ ] Crear nueva guia de carga de datos iniciales.
- [ ] Actualizar `README.md`.

Pregunta pendiente para documentacion:

```txt
Nueva base de datos elegida: Azure SQL Database
```

## Conexion objetivo con Azure SQL Database

La base de datos objetivo del proyecto sera Azure SQL Database. La implementacion actual del backend todavia no esta lista para esta conexion, porque usa PostgreSQL mediante `pgx`. Esta seccion documenta el destino esperado para la migracion.

Variables sugeridas para la futura conexion:

```env
DB_SERVER=servidor.database.windows.net
DB_PORT=1433
DB_NAME=poli_redi
DB_USER=usuario
DB_PASSWORD=
DB_ENCRYPT=true
```

Tambien se puede usar una cadena de conexion unica, si se decide simplificar la configuracion:

```env
# DATABASE_URL=sqlserver://USUARIO:<PASSWORD>@SERVIDOR.database.windows.net:1433?database=poli_redi&encrypt=true
```

Checklist para Azure SQL Database:

- [ ] Confirmar nombre del servidor SQL en Azure.
- [ ] Confirmar nombre de la base de datos.
- [ ] Confirmar usuario de aplicacion.
- [ ] Confirmar reglas de firewall para permitir acceso desde desarrollo local.
- [ ] Confirmar puerto `1433`.
- [ ] Confirmar cifrado de conexion.
- [ ] Migrar `schema.sql` a sintaxis T-SQL.
- [ ] Migrar `seed.sql` a sintaxis compatible con Azure SQL.
- [ ] Cambiar driver Go de PostgreSQL a SQL Server.
- [ ] Actualizar repositorios y consultas SQL.
- [ ] Probar `GET /api/health` contra Azure SQL Database.

Nota tecnica:

Azure SQL Database no soporta directamente varias caracteristicas usadas en el esquema PostgreSQL actual, como `EXCLUDE USING gist`, `tsrange`, `JSONB` y funciones PL/pgSQL. Esas reglas deben rediseñarse con T-SQL, indices, constraints, triggers o validaciones en la capa de servicio.

## Configuracion real de Azure SQL Database

La base de datos objetivo del proyecto es Azure SQL Database. Los datos de conexion confirmados son:

```env
DB_SERVER=poli-redi-server.database.windows.net
DB_PORT=1433
DB_NAME=poli-redi-database
DB_USER=poli-redi-admin
DB_PASSWORD=
DB_ENCRYPT=true
DB_TRUST_SERVER_CERTIFICATE=false
```

`DB_PASSWORD` debe completarse solo en el archivo `.env` local o en las variables de entorno del despliegue. No debe guardarse en archivos versionados.

El backend tambien acepta una cadena completa opcional:

```env
# AZURE_SQL_CONNECTION_STRING=server=poli-redi-server.database.windows.net;user id=poli-redi-admin;password=;port=1433;database=poli-redi-database;encrypt=true;trustservercertificate=false;
```

Archivos relevantes:

- `backend/.env.example`: plantilla segura sin password.
- `backend/internal/database/database.go`: conexion a Azure SQL mediante `github.com/microsoft/go-mssqldb`.
- `database/schema.sql`: DDL compatible con Azure SQL Database.
- `database/seed.sql`: datos iniciales compatibles con Azure SQL Database.
- `database/drop.sql`: limpieza compatible con Azure SQL Database.

Importante:

Si Azure SQL muestra un error como `Unknown object type 'EXTENSION'`, significa que se esta ejecutando un script antiguo de PostgreSQL. El script correcto no contiene `CREATE EXTENSION`, `EXCLUDE USING gist`, `tsrange` ni `plpgsql`.

Orden recomendado para una base nueva:

1. Ejecutar `database/schema.sql`.
2. Ejecutar `database/seed.sql` si se quieren datos iniciales.
3. Configurar `backend/.env` usando `backend/.env.example` como referencia.
4. Ejecutar el backend.
5. Probar `GET /api/health`.
