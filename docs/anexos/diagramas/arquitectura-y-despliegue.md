# Anexo — Arquitectura interna y despliegue

**Audiencia:** Arquitecto, desarrollo, QA y DevOps

**Propósito:** representar componentes internos y la topología prevista u observable de despliegue

**Estado:** ANEXO TÉCNICO CONTRASTADO CON EL CÓDIGO LOCAL

**Corte:** 2026-08-11

**Fuente:** código local, configuración de despliegue y arquitectura canónica

**No demuestra:** disponibilidad online actual ni despliegue integrado de frontend, API, identidad y base

## Resumen

- La API combina servicios de dominio con accesos directos a repositorios.
- `RequireAuth` valida identidad y sincroniza el usuario local mediante el
  repositorio; no existe un componente `AdminGuard`.
- El repositorio contiene automatización de Azure Static Web Apps para el
  frontend, no un pipeline equivalente para el backend.
- App Service y Azure SQL describen la topología objetivo/documentada; su
  operación actual no fue verificada en este corte.

Volver a [Arquitectura y contratos](../../02-arquitectura-y-contratos.md).

## 1. Componentes internos reales

**Objetivo:** mostrar la estructura real sin imponer una capa Service donde el
código no la utiliza.

```mermaid
flowchart LR
    subgraph BROWSER["Navegador"]
        ROUTER["Vue Router<br/>beforeEach + meta"]
        VIEWS["Views y componentes"]
        STORES["Pinia Stores"]
        FESVC["Services Axios"]
        ROUTER --> VIEWS --> STORES --> FESVC
    end

    ENTRA["Microsoft Entra ID<br/>OIDC + JWKS"]

    subgraph BACKEND["API Go / Fiber"]
        ROUTES["routes.RegisterRoutes"]
        AUTH["RequireAuth / RequireAdmin"]
        HANDLERS["Handlers"]
        SERVICES["Services<br/>reservas, talleres, políticas"]
        REPOS["Repositories"]
        RULES["reservationrules<br/>businessclock"]
        SECRET["joinsecret AES"]
        WORKER["Worker expiración<br/>cada 30 s"]

        ROUTES --> AUTH --> HANDLERS
        HANDLERS --> SERVICES
        SERVICES --> RULES
        SERVICES --> REPOS
        AUTH -.-> REPOS
        REPOS --> SECRET
        WORKER --> REPOS

        DIRECT["Accesos directos reales:<br/>participantes, RUT, recursos,<br/>usuarios, actividades y notificaciones"]
        HANDLERS -.-> DIRECT -.-> REPOS
    end

    DB[("Azure SQL Database<br/>constraints, índices, triggers y vistas")]

    ROUTER --> ENTRA
    FESVC -->|"REST JSON / Bearer"| ROUTES
    AUTH -->|"Validar JWT / JWKS"| ENTRA
    REPOS -->|"go-mssqldb / T-SQL"| DB
```

Decisiones y divergencias:

- Participantes, perfil/RUT, recursos, usuarios, actividades y notificaciones
  llaman al repositorio directamente en todo o parte de sus operaciones.
- Reservas, talleres y políticas poseen servicios de dominio explícitos.
- El middleware consulta/sincroniza al usuario local; la autorización no se
  deduce de la interfaz ni de un guard inexistente.
- `joinsecret` cifra el código recuperable; el hash de búsqueda permanece en
  la reserva y no sustituye el secreto owner-only.

## 2. Despliegue documentado

**Objetivo:** separar el pipeline frontend observable del procedimiento
backend que no está automatizado en este repositorio.

```mermaid
flowchart TB
    DEV["Repositorio GitHub"]
    GHA["GitHub Actions<br/>Static Web Apps workflow"]
    SWA["Azure Static Web Apps<br/>Frontend Vue compilado"]
    OPS["Proceso backend externo/manual<br/>no definido en .github"]
    IMAGE["Imagen Docker backend<br/>tag inmutable recomendado"]
    APP["Azure App Service<br/>contenedor Go/Fiber"]
    SQL[("Azure SQL Database")]
    ENTRA["Microsoft Entra ID"]
    USER["Navegador"]

    DEV -->|"push main / PR"| GHA
    GHA -->|"npm build + deploy"| SWA

    DEV -->|"backend/Dockerfile"| OPS
    OPS -->|"build y publicación"| IMAGE
    IMAGE -->|"actualización de contenedor"| APP

    USER -->|"HTTPS SPA"| SWA
    USER -->|"OAuth 2.0 / OIDC"| ENTRA
    SWA -->|"REST HTTPS<br/>VITE_API_BASE_URL"| APP
    APP -->|"JWKS / issuer / audience"| ENTRA
    APP -->|"TLS / T-SQL"| SQL

    CONFIG["Configuración protegida:<br/>CORS, DB, Entra, APP_TIMEZONE,<br/>JOIN_CODE_ENCRYPTION_KEYS"]
    CONFIG -.-> APP

    NOTE["Estado online actual:<br/>no verificado en este corte"]
    NOTE -.-> SWA
    NOTE -.-> APP
```

Decisiones y divergencias:

- `.github/workflows/azure-static-web-apps-purple-ground-0205c9f10.yml`
  automatiza build y despliegue del frontend.
- No existe un workflow equivalente para construir, publicar y actualizar la
  imagen backend. El diagrama no inventa un registro de contenedores.
- `DEV_AUTH_ENABLED` y `VITE_DEV_AUTH_ENABLED` deben permanecer desactivados en
  ambientes públicos.
- CORS, redirect URIs, health, commit, versiones y conexión a Azure SQL deben
  verificarse juntos antes de afirmar operación online.

## Documentos relacionados

- [Arquitectura y contratos](../../02-arquitectura-y-contratos.md)
- [Base de datos y migraciones](../../04-base-de-datos-y-migraciones.md)
- [Reglas y flujos](../../06-reglas-y-flujos.md)
- [Modelo entidad-relación](modelo-entidad-relacion.md)

## Fuentes

- [`backend/internal/routes/routes.go`](../../../backend/internal/routes/routes.go)
- [`backend/internal/middleware/auth_middleware.go`](../../../backend/internal/middleware/auth_middleware.go)
- [`backend/internal/handlers`](../../../backend/internal/handlers)
- [`backend/internal/services`](../../../backend/internal/services)
- [`backend/internal/repositories`](../../../backend/internal/repositories)
- [`backend/Dockerfile`](../../../backend/Dockerfile)
- [`frontend/src/router/index.js`](../../../frontend/src/router/index.js)
- [`.github/workflows/azure-static-web-apps-purple-ground-0205c9f10.yml`](../../../.github/workflows/azure-static-web-apps-purple-ground-0205c9f10.yml)
