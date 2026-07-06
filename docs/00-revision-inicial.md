# Poli-REDI - Revision inicial del proyecto

## Objetivo del documento

Este documento registra el estado inicial del proyecto Poli-REDI antes de continuar con nuevas funcionalidades. La idea es partir desde una revision ordenada: entender que existe, que falta, como esta organizado el codigo y que tareas deberian priorizarse.

Nota de mantenimiento: este documento es historico. El estado vigente del proyecto ya usa Azure SQL Database, `github.com/microsoft/go-mssqldb`, frontend Vue/Vite, backend Go/Fiber y demo online inicial en Azure. Para instrucciones actuales usar `README.md`, `docs/01-instalacion-y-ejecucion.md`, `docs/02-arquitectura.md` y `docs/03-base-de-datos.md`.

## Descripcion general

Poli-REDI es un sistema web para la gestion de reservas deportivas. El proyecto contempla una aplicacion frontend, una API backend y una base de datos relacional para administrar usuarios, recursos reservables, reservas, disponibilidad, infracciones, notificaciones y reportes.

## Stack tecnico identificado

### Frontend

- Vue 3
- Vite
- Pinia
- Vue Router
- Axios
- MSAL para autenticacion con Microsoft/Azure

### Backend

- Go
- Fiber
- Capa de base de datos actualmente implementada con pgx, pendiente de migracion

### Base de datos

- Azure SQL Database
- Implementacion heredada basada en scripts SQL en la carpeta `database/`
- Restricciones para evitar solapamiento de reservas
- Triggers para auditoria, timestamps, infracciones y notificaciones

## Estructura principal del repositorio

```txt
Poli-REDI/
  backend/
  database/
  files/
  frontend/
  README.md
  package.json
```

## Carpetas identificadas

### `backend/`

Contiene la API del sistema desarrollada en Go. Se observa una estructura por capas:

- `cmd/`: punto de entrada de la aplicacion.
- `internal/database/`: configuracion de conexion a base de datos.
- `internal/handlers/`: controladores HTTP.
- `internal/middleware/`: middleware de autenticacion.
- `internal/models/`: modelos de dominio.
- `internal/repositories/`: acceso a datos.
- `internal/routes/`: registro de rutas.
- `internal/services/`: logica de negocio.

### `frontend/`

Contiene la aplicacion web desarrollada con Vue 3 y Vite. Se observa una organizacion por vistas, componentes, servicios, stores y utilidades:

- `src/views/`: pantallas principales del sistema.
- `src/components/`: componentes reutilizables.
- `src/services/`: comunicacion con la API.
- `src/stores/`: estado global con Pinia.
- `src/router/`: rutas del frontend.
- `src/auth/`: configuracion y servicios de autenticacion.
- `src/utils/`: funciones auxiliares.

### `database/`

Contiene scripts SQL de la implementacion anterior. Estos archivos sirven como referencia del modelo de datos, pero deben revisarse si el proyecto deja de usar PostgreSQL:

- `schema.sql`
- `schema_0.1.sql`
- `seed.sql`
- `drop.sql`

### `files/`

Contiene archivos CSV de apoyo o datos de carga:

- `azure.csv`
- `data.csv`
- `users.csv`

## Funcionalidades identificadas en el codigo

### Backend

Rutas identificadas:

- `GET /api/health`
- `GET /api/me`
- `GET /api/resources`
- `GET /api/reservations`
- `POST /api/reservations`
- `PATCH /api/reservations/cancel`

Funcionalidades presentes:

- Consulta de salud de la API.
- Obtencion de usuario autenticado.
- Listado de recursos.
- Listado de reservas.
- Creacion de reservas.
- Cancelacion de reservas.
- Middleware de autenticacion.

### Base de datos

Tablas identificadas:

- `users`
- `resources`
- `activities`
- `reservations`
- `participants`
- `violations`
- `priority_reservations`
- `availability_blocks`
- `notifications`
- `logs`

Vistas identificadas:

- `vw_resource_usage`
- `vw_peak_hours`
- `vw_user_violations`

Reglas importantes:

- Evita solapamiento de reservas por recurso.
- Evita que un usuario tenga reservas simultaneas.
- Registra cambios relevantes mediante logs.
- Genera notificaciones al crear infracciones.

### Frontend

Vistas identificadas:

- `DashboardView.vue`
- `ResourcesView.vue`
- `ReservationsView.vue`
- `ReservationDetailView.vue`
- `AvailabilityView.vue`
- `AdminView.vue`
- `UsersView.vue`
- `ReportsView.vue`
- `HistoryView.vue`
- `SettingsView.vue`
- `AuthCallbackView.vue`
- `NotFoundView.vue`

Componentes relevantes:

- Componentes de layout: sidebar, header, menu de usuario y notificaciones.
- Componentes de dashboard: tarjetas, carrusel, panel de reservas y acciones rapidas.
- Componentes de disponibilidad: calendario, grilla horaria, timeline y modal de detalle.
- Componentes de formularios: selector de recurso, formulario de reserva y selector de fecha/hora.
- Componentes de administracion: metricas, alertas y resumen.

## Observaciones iniciales

- El proyecto ya tiene una base tecnica importante; no parece estar en etapa cero de codigo.
- La base de datos actual esta modelada en scripts PostgreSQL, pero existe una decision de cambiar la tecnologia de base de datos.
- El backend ya expone rutas centrales para recursos y reservas.
- El frontend ya tiene varias vistas y componentes creados.
- Antes de implementar nuevas funcionalidades conviene revisar si las pantallas estan conectadas realmente con la API.
- Algunas piezas parecen estar parcialmente implementadas, por ejemplo la vista `ReservationsView.vue` muestra contenido pendiente.

## Riesgos o puntos a revisar

- Migrar la implementacion de base de datos desde PostgreSQL/pgx hacia Azure SQL Database.
- Confirmar que el frontend compila correctamente.
- Confirmar que la autenticacion funciona en entorno local.
- Verificar si las rutas protegidas pueden probarse sin credenciales reales.
- Revisar si los servicios frontend coinciden con las rutas actuales del backend.
- Revisar consistencia entre modelos Go, respuestas JSON y consumo desde Vue.
- Revisar manejo de errores, especialmente conflictos de reserva.
- Revisar si existen tests o si se deben crear casos minimos de prueba.

## Estado inicial recomendado

El proyecto deberia avanzar primero por revision y documentacion, no por nuevas features. El orden sugerido es:

1. Documentar instalacion y ejecucion local.
2. Verificar backend.
3. Verificar frontend.
4. Revisar flujo principal de reservas.
5. Documentar arquitectura.
6. Crear backlog real basado en hallazgos.

## Proximo documento sugerido

`docs/01-instalacion-y-ejecucion.md`

Este documento deberia explicar como ejecutar el backend, el frontend y la base de datos en ambiente local.

## Decision posterior sobre base de datos

Se definio que Poli-REDI ya no usara PostgreSQL. La base de datos objetivo sera Azure SQL Database. Por ahora, el codigo y la documentacion anterior todavia muestran rastros de PostgreSQL, `pgx`, `DATABASE_URL` y scripts SQL. Eso debe considerarse estado heredado.

Trabajo pendiente:

- Usar Azure SQL Database como nueva tecnologia de base de datos.
- Revisar si se mantiene Go/Fiber como backend.
- Migrar la capa `internal/database/`.
- Migrar los repositorios en `internal/repositories/`.
- Adaptar o reemplazar los scripts de `database/`.
- Actualizar variables de entorno y guia de instalacion.
- Actualizar el backlog tecnico con tareas de migracion.


## Decision de base de datos objetivo

La base de datos objetivo para Poli-REDI sera Azure SQL Database. Esto implica migrar la implementacion actual, que fue construida con PostgreSQL y `pgx`, hacia una conexion compatible con SQL Server/Azure SQL.

Impacto esperado:

- Cambiar driver de base de datos en el backend.
- Revisar sintaxis SQL de repositorios.
- Migrar scripts de `database/` desde PostgreSQL a T-SQL.
- Reemplazar restricciones o funciones propias de PostgreSQL, como `EXCLUDE USING gist`, `tsrange`, `JSONB` y triggers PL/pgSQL.
- Actualizar variables de entorno y documentacion de instalacion.
