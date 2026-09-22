# Poli-REDI - Despliegue y operación vigente

Fecha de corte: 2026-09-22

## Propósito

Esta guía describe el despliegue vigente de Poli-REDI en Debian sobre WSL:

- PostgreSQL 16;
- backend Go/Fiber;
- frontend Vue/Vite servido por Caddy;
- tres contenedores administrados como un solo stack con Podman Compose;
- arranque mediante un servicio systemd de usuario;
- publicación HTTPS mediante Tailscale Funnel.

La antigua demo de Azure Static Web Apps, Azure App Service y Azure SQL es un
antecedente histórico. No debe utilizarse como receta de despliegue actual.

## Arquitectura desplegada

```text
Internet
  -> Tailscale Funnel :443
  -> https+insecure://localhost:8443
  -> Caddy
       -> /api/* -> backend:3000
       -> /*     -> frontend estático
  -> backend
  -> PostgreSQL 16
```

El stack está versionado en el repositorio de infraestructura:

```text
~/projects/poliredi-infra/
├── compose/
│   ├── compose.yaml
│   ├── Caddyfile
│   ├── README.md
│   └── OPERACION.md
└── systemd/
    └── poliredi-stack.service
```

El repositorio de la aplicación permanece en `~/projects/poliredi`.

## Componentes y persistencia

| Servicio | Imagen o artefacto | Red y publicación |
| --- | --- | --- |
| PostgreSQL | `postgres:16-alpine` | Red interna; `127.0.0.1:55432` para administración local |
| Backend | `localhost/poliredi-backend:mvp2` | Redes de base y frontend; sin puerto público |
| Web | `caddy:2.11.4-alpine` + `frontend/dist` | `127.0.0.1:8443` |

PostgreSQL utiliza el volumen externo `poliredi-postgres-mvp1-data`. El stack
adopta ese volumen; no crea otra base ni ejecuta migraciones automáticamente.
Nunca deben iniciarse dos servidores PostgreSQL sobre el mismo volumen.

## Requisitos del host

- Debian en WSL con systemd.
- Podman rootless 5.4.2 o compatible.
- `podman-compose` 1.3.0 o compatible.
- Tailscale conectado y Funnel autorizado.
- Imágenes backend y frontend compiladas y validadas.

Fijar el proveedor Linux para no invocar Docker Compose de Windows:

```bash
export PODMAN_COMPOSE_PROVIDER=/usr/bin/podman-compose
podman compose version
```

## Configuración privada

Los archivos privados permanecen fuera de Git:

```text
~/.config/poli-redi/compose.env
~/.config/poli-redi/backend.env
~/.config/containers/systemd/poliredi-postgres.env
```

`compose.env` define las rutas absolutas y la imagen backend. `backend.env`
contiene Entra ID, CORS y configuración de la API. Compose fuerza
`DATABASE_URL` vacía y usa `PGHOST=postgres` con el resto de variables `PG*`.

La contraseña del rol `poliredi_app` se inyecta desde el secreto externo Podman
`poliredi-postgres-app-password` como `PGPASSWORD`. La configuración `type: env`
fue validada con podman-compose 1.3.0 y debe revisarse si se cambia de proveedor.
No regenerar claves de códigos de invitación: podría invalidar códigos vigentes.

El frontend público se compila con:

```env
VITE_API_BASE_URL=/api
VITE_MVP_SCOPE=mvp2
VITE_DEV_AUTH_ENABLED=false
VITE_ENTRA_REDIRECT_URI=https://desktop-epot7cf.tail16d8fb.ts.net/auth/callback
VITE_ENTRA_POST_LOGOUT_REDIRECT_URI=https://desktop-epot7cf.tail16d8fb.ts.net/login
```

Las demás variables de Entra deben coincidir con el registro de aplicación.

## Validar la configuración

```bash
export PODMAN_COMPOSE_PROVIDER=/usr/bin/podman-compose
cd ~/projects/poliredi-infra/compose
podman compose --env-file ~/.config/poli-redi/compose.env config >/dev/null

podman run --rm --network none \
  -v "$PWD/Caddyfile:/etc/caddy/Caddyfile:ro" \
  docker.io/library/caddy:2.11.4-alpine \
  caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile
```

## Construir una nueva versión

Backend:

```bash
cd ~/projects/poliredi/backend
podman build -t localhost/poliredi-backend:mvp2 -f Containerfile .
```

Frontend:

```bash
cd ~/projects/poliredi/frontend
npm ci
npm test
npm run build
```

El frontend es un bind mount de `frontend/dist`; no requiere reconstruir Caddy.
Después de reemplazar la imagen backend o el build frontend, recrear el servicio
afectado y verificar salud. Evitar tags ambiguos para entregas reproducibles.

## Operación

Estado general:

```bash
systemctl --user status poliredi-stack.service --no-pager
podman ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'
curl --fail --insecure https://localhost:8443/api/health
```

Inicio y parada del stack:

```bash
systemctl --user start poliredi-stack.service
systemctl --user stop poliredi-stack.service
```

Registros:

```bash
export PODMAN_COMPOSE_PROVIDER=/usr/bin/podman-compose
cd ~/projects/poliredi-infra/compose
podman compose --env-file ~/.config/poli-redi/compose.env logs --tail=100
```

La unidad es `oneshot` con `RemainAfterExit=yes`; `active (exited)` confirma que
Compose terminó, no que los contenedores estén saludables. Siempre revisar
`podman ps` y `/api/health`.

Durante el arranque Caddy puede responder `502` mientras PostgreSQL alcanza
`healthy` y arranca la API. El estado debe recuperarse sin intervención.

## Publicación con Funnel

Estado esperado:

```bash
tailscale funnel status
```

Destino:

```text
https://desktop-epot7cf.tail16d8fb.ts.net
  -> https+insecure://localhost:8443
```

Validación local con el host público:

```bash
curl --insecure \
  -H 'Host: desktop-epot7cf.tail16d8fb.ts.net' \
  https://localhost:8443/api/health
```

También debe probarse desde otra red. En el corte, el navegador público funcionó,
pero `curl` desde Debian al dominio público agotó el tiempo TLS; esta diferencia
permanece pendiente de investigación.

## Arranque automático

La unidad instalada es:

```text
~/.config/systemd/user/poliredi-stack.service
```

Debe estar habilitada y el usuario debe conservar linger:

```bash
systemctl --user is-enabled poliredi-stack.service
loginctl show-user "$USER" -p Linger
```

La prueba del corte confirmó que, después de terminar y volver a abrir Debian,
systemd inició el stack sin ejecutar Compose manualmente. Esto no inicia WSL por
sí solo al encender Windows.

## Respaldo y recuperación

Antes de migraciones o cambios de persistencia, crear un dump en formato custom
y comprobar su catálogo con `pg_restore --list`. No guardar respaldos ni secretos
en Git.

Respaldo anterior al corte Compose:

```text
~/projects/poliredi-backups/pre-compose-20260922-172926
```

Quadlets retirados durante el corte:

```text
~/projects/poliredi-backups/corte-compose-20260922-173956/quadlet
```

Los Quadlets anteriores no constituyen un rollback listo: la API buscaba
`poliredi-postgres` sin que la base MVP1 estuviera disponible en su red. Restaurar
solo los archivos reproduce el fallo. Para volver a Quadlet se debe corregir la
red o alias, reutilizar PostgreSQL 16 y el secreto validado, detener Compose con
`down` sin `-v` y confirmar que el volumen está libre antes de iniciar la base.

Nunca ejecutar `down -v` sobre este stack ni borrar
`poliredi-postgres-mvp1-data` como parte de un redeploy.

## Validación mínima posterior a un cambio

1. Los tres contenedores aparecen `healthy`.
2. `/api/health` responde por Caddy.
3. El login Microsoft funciona.
4. `/api/me` responde para una sesión válida.
5. Se conservan recursos, reservas y participantes existentes.
6. Disponibilidad, creación, cancelación y códigos grupales funcionan.
7. Funnel abre la aplicación desde un navegador externo.
8. El stack vuelve a iniciar después de reiniciar Debian.

## Antecedentes históricos

Azure Static Web Apps, Azure App Service, Azure SQL, SQL Server y Docker aparecen
en documentos de evolución y evidencia de julio de 2026. Esas referencias son
válidas solo como historia del proyecto. PostgreSQL 16, Podman Compose, Caddy,
systemd y Tailscale Funnel forman la arquitectura desplegada vigente.
