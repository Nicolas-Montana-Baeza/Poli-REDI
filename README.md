# Poli-REDI

Sistema web para centralizar la disponibilidad, las reservas y las actividades deportivas institucionales.

## Lectura rápida

- [Estado actual](docs/01-estado-actual.md)
- [Plan de entrega 2026](docs/08-plan-entrega-2026.md)
- [Checklist de cierre](docs/09-checklist-cierre.md)
- [Gobierno documental](docs/00-gobierno-documental.md)

El estado cambia con cada incremento. Este README no reemplaza el dictamen fechado de `docs/01-estado-actual.md`.

## Capacidades del prototipo

- autenticación institucional con Microsoft Entra ID y modo local de prueba;
- consulta de recursos y disponibilidad;
- reservas particulares y solicitudes grupales;
- participantes, códigos de invitación y progreso de confirmación;
- talleres e historial personal;
- consulta administrativa básica e indicadores;
- evolución planificada hacia programación institucional, soporte y despliegue.

El alcance aprobado y sus exclusiones se mantienen en [requisitos y trazabilidad](docs/03-requisitos-y-trazabilidad.md).

## Tecnología

| Capa | Tecnología |
|---|---|
| Frontend | Vue 3, Vite, Pinia, Vue Router, Axios y MSAL Browser |
| Backend | Go, Fiber y `go-mssqldb` |
| Datos | Azure SQL Database / SQL Server local |
| Identidad | Microsoft Entra ID |
| Despliegue | Azure Static Web Apps y Azure App Service |

## Inicio local

1. Preparar las variables locales sin reutilizar secretos de Azure.
2. Levantar SQL Server o conectarse a una base de desarrollo autorizada.
3. Aplicar el esquema o las migraciones según el estado de la base.
4. Iniciar backend y frontend.
5. Ejecutar las verificaciones aplicables antes de probar manualmente.

Los comandos, variables, migraciones y procedimientos de recuperación están en [instalación y despliegue](docs/05-instalacion-despliegue-recuperacion.md). Las normas de colaboración están en [CONTRIBUTING.md](CONTRIBUTING.md).

## Documentación por audiencia

### Producto y tesis

1. [Estado actual](docs/01-estado-actual.md)
2. [Requisitos y trazabilidad](docs/03-requisitos-y-trazabilidad.md)
3. [Plan de entrega](docs/08-plan-entrega-2026.md)

### Desarrollo y arquitectura

1. [Arquitectura y contratos](docs/02-arquitectura-y-contratos.md)
2. [Reglas y flujos](docs/06-reglas-y-flujos.md)
3. [Base de datos y migraciones](docs/04-base-de-datos-y-migraciones.md)
4. [Calidad y evidencia](docs/07-calidad-y-evidencia.md)

### Operación y cierre

1. [Instalación, despliegue y recuperación](docs/05-instalacion-despliegue-recuperacion.md)
2. [Plan de entrega](docs/08-plan-entrega-2026.md)
3. [Checklist de cierre](docs/09-checklist-cierre.md)

## Seguridad

- No versionar `.env`, tokens, claves, RUT ni cadenas de conexión.
- `DEV_AUTH_ENABLED` debe estar desactivado fuera del desarrollo local.
- La autorización efectiva se valida en backend.
- Las respuestas y evidencias deben minimizar datos por audiencia.

Ante una discrepancia documental, aplicar la precedencia definida en [gobierno documental](docs/00-gobierno-documental.md).
