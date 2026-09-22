# Poli-REDI - Matriz de vigencia documental

Fecha de corte: 2026-09-22

## Propósito

Esta matriz indica qué documentos describen el producto actual y cuáles se
conservan como evidencia histórica. Una referencia a Azure, SQL Server, Docker
o Quadlet no es necesariamente un error: lo es cuando aparece como instrucción
vigente fuera del contexto histórico o de desarrollo aislado.

## Fuentes operativas vigentes

| Documento | Uso actual |
| --- | --- |
| `README.md` | Punto de entrada, stack, ejecución y enlaces principales |
| `docs/00-resumen-proyecto.md` | Estado ejecutivo y brechas vigentes |
| `docs/01-instalacion-y-ejecucion.md` | Desarrollo local y verificación aislada |
| `docs/02-arquitectura.md` | Arquitectura lógica y de despliegue vigente |
| `docs/03-base-de-datos.md` | Modelo PostgreSQL y genealogía de persistencia |
| `docs/04-backend.md` | Contratos y estructura del backend vigente |
| `docs/05-frontend.md` | Arquitectura y experiencia del frontend vigente |
| `docs/06-flujo-reservas.md` | Flujos funcionales actuales |
| `docs/08-requisitos-historias-casos-uso.md` | Requisitos y casos de uso vigentes |
| `docs/09-mvps-roadmap.md` | Alcance incremental y pendientes |
| `docs/10-guia-redeploy.md` | Despliegue y operación actual |
| `docs/13-estado-actual-producto.md` | Evaluación de producto y contradicciones |
| `docs/14-evolucion-y-trazabilidad-requisitos.md` | Evolución de decisiones y requisitos |
| `docs/15-checklist-cierre-mvp2.md` | Evidencia y cierre técnico de MVP2 |
| `database/README.md` | Fuente de verdad de persistencia vigente |
| `database/postgres/README.md` | Desarrollo y verificación PostgreSQL |

La configuración de producción local vive en el repositorio separado
`poliredi-infra`, especialmente `compose/OPERACION.md`.

## Documentos históricos o mixtos

| Documento | Regla de lectura |
| --- | --- |
| `docs/00-revision-inicial.md` | Estado histórico; no usar para instalar ni definir arquitectura actual |
| `docs/07-backlog.md` | Registro acumulado; entradas Azure/SQL Server describen etapas cerradas |
| `docs/12-checklist-demo-mvp1.md` | Evidencia de la demo Azure de julio; no prueba el stack actual |
| Scripts T-SQL en `database/` | Legado SQL Server; no ejecutar contra PostgreSQL |

## Tecnologías vigentes

- Frontend Vue 3/Vite compilado como archivos estáticos.
- Backend Go/Fiber con `pgx`.
- PostgreSQL 16 y migraciones `PG16_*`.
- Podman Compose para administrar PostgreSQL, backend y Caddy.
- systemd de usuario para iniciar el stack en Debian.
- Caddy como servidor estático y proxy `/api/*`.
- Tailscale Funnel para publicación HTTPS.
- Microsoft Entra ID para autenticación real.

## Tecnologías históricas

- Azure SQL Database y SQL Server fueron una etapa intermedia de persistencia.
- Azure Static Web Apps y Azure App Service alojaron una demo anterior.
- Docker aparece en esa historia y en nombres de registros de imágenes; el
  runtime actual utiliza Podman. Las referencias `docker.io/...` son nombres de
  registro OCI y no implican usar Docker Engine.
- Quadlet se conserva para crear bases aisladas de desarrollo y ejecutar
  verificaciones; no administra el stack desplegado actual.

## Regla de mantenimiento

Todo cambio de arquitectura debe actualizar primero README, resumen, instalación,
arquitectura, despliegue y esta matriz. La trazabilidad histórica no se reescribe:
se etiqueta con fecha, estado y relación con la arquitectura vigente.
