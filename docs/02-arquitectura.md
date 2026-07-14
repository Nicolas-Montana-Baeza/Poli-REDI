# Poli-REDI - Arquitectura

## Objetivo

Este documento resume la arquitectura vigente de Poli-REDI y el flujo principal entre frontend, backend, autenticacion y base de datos.

## Vista general

```mermaid
flowchart LR
  User["Usuario"] --> Frontend["Frontend Vue/Vite"]
  Frontend --> Entra["Microsoft Entra ID"]
  Frontend --> API["Backend Go/Fiber"]
  API --> EntraKeys["JWKS / validacion JWT"]
  API --> DB["Azure SQL Database"]
  DB --> API
  API --> Frontend
```

## Frontend

El frontend vive en `frontend/` y usa:

- Vue 3.
- Vite.
- Pinia.
- Vue Router.
- Axios.
- MSAL Browser.

Responsabilidades principales:

- Autenticar al usuario con Microsoft Entra ID o modo local de desarrollo.
- Proteger rutas publicas, autenticadas y administrativas.
- Consultar recursos, actividades, reservas, talleres, notificaciones y usuario actual.
- Mostrar disponibilidad, formularios de reserva, historial, talleres deportivos y panel admin base.
- Mostrar imagenes configurables de recursos y permitir su actualizacion a administradores.
- Enviar reservas sin confiar en IDs de usuario definidos en cliente.
- Consumir disponibilidad desde un endpoint sanitizado y combinar reservas, talleres recurrentes y actividades institucionales para bloquear horarios ocupados.

## Backend

El backend vive en `backend/` y usa Go/Fiber.

Estructura principal:

- `cmd/`: arranque de la API.
- `internal/routes/`: registro de rutas.
- `internal/middleware/`: autenticacion, modo local y usuario actual.
- `internal/handlers/`: entrada HTTP.
- `internal/services/`: reglas de negocio.
- `internal/repositories/`: consultas Azure SQL.
- `internal/models/`: modelos de datos.
- `internal/validators/`: validadores reutilizables.

Responsabilidades principales:

- Validar tokens Bearer de Microsoft Entra ID.
- Resolver o crear usuario local autenticado.
- Aplicar permisos de usuario normal y administrador.
- Crear y cancelar reservas usando el usuario autenticado.
- Listar talleres activos e inscribir al usuario autenticado cuando tenga RUT.
- Actualizar imagenes de recursos mediante ruta administrativa protegida.
- Exponer disponibilidad sanitizada para usuarios normales y detalle completo de reservas solo a administradores.
- Validar que reservas no se crucen con talleres activos asociados al mismo recurso.
- Validar RUT obligatorio para usuarios normales.
- Exponer datos desde Azure SQL Database.

## Base de datos

La base vigente es Azure SQL Database.

Scripts principales:

- `database/schema.sql`.
- `database/seed.sql`.
- `database/drop.sql`.

La base aplica reglas criticas mediante constraints, indices, triggers y vistas:

- Usuarios, recursos, actividades, reservas, talleres e inscripciones.
- Imagen opcional por recurso mediante `resources.image_url`.
- Validacion basica de RUT.
- Conflictos de reserva por recurso y usuario.
- Recursos de uso libre (`OPEN_USE`) que permiten concurrencia y se visualizan como intensidad de uso.
- Control de cupos e inscripcion unica por usuario en talleres.
- Bloqueos y actividades programadas para iteraciones administrativas.
- Notificaciones y auditoria.

## Autenticacion

Flujo real:

1. El usuario inicia sesion con Microsoft Entra ID.
2. El frontend obtiene token para la API.
3. El backend valida issuer, audience, firma y claims.
4. El backend busca o crea usuario local.
5. Las rutas protegidas usan el usuario local resuelto.

Flujo local:

1. `DEV_AUTH_ENABLED=true` habilita accesos locales.
2. El frontend envia headers `X-Dev-Auth-*`.
3. El backend crea o resuelve un usuario local de prueba.
4. Este modo no debe activarse en ambientes publicos.

## Flujo de reserva

```mermaid
sequenceDiagram
  actor Usuario
  participant UI as Frontend
  participant API as Backend
  participant DB as Azure SQL

  Usuario->>UI: Selecciona horario
  UI->>API: POST /api/reservations
  API->>API: Valida usuario autenticado y RUT
  API->>DB: Inserta reserva
  DB->>DB: Valida conflictos
  DB-->>API: Resultado
  API-->>UI: Reserva creada o error
  UI-->>Usuario: Actualiza disponibilidad
```

## Despliegue

La demo online inicial usa:

- Frontend: Azure Static Web Apps.
- Backend: Azure App Service con Docker.
- Base de datos: Azure SQL Database.
- Variables frontend `VITE_*` desde GitHub Actions.
- Variables backend en App Service.

## Riesgos y mejoras recomendadas

- Extender el endpoint de disponibilidad para aceptar fecha o rango y sumar bloqueos; las actividades programadas activas ya forman parte del contrato.
- Definir un contrato temporal unico entre `DATETIME2`, API y frontend mediante `APP_TIMEZONE` (`RES-009`).
- Hacer que estado inicial y transiciones de reserva pertenezcan exclusivamente al servidor (`RES-010`).
- Aplicar jornada y duraciones permitidas en backend (`RES-011`).
- Agregar pruebas backend reales para reglas criticas; `go test ./...` actualmente no descubre archivos de test (`QA-001`).
- Agregar regresion frontend automatizada para router, formulario, recursos no reservables y helpers temporales (`QA-002`).
- Evitar detalles internos en respuestas HTTP y conservarlos solo en logs sanitizados (`SEC-005`).
- Mantener `RequireAdmin` para nuevas rutas administrativas y `DEV_AUTH_ENABLED=false` en despliegues publicos; ambas bases ya estan implementadas/documentadas.
