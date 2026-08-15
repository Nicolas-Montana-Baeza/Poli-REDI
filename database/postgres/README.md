# PostgreSQL 16 local para MVP1

Este directorio define una base PostgreSQL nueva para el MVP1. No migra datos
ni ejecuta las migraciones T-SQL históricas. Reservas grupales, códigos,
talleres, programación institucional y notificaciones están fuera de esta fase.

## Requisitos

- Linux o WSL con Podman rootless y cgroup v2.
- Podman Quadlet y `systemctl --user` disponibles.
- Go para ejecutar posteriormente la API.

No se utiliza Docker Compose ni Docker Desktop. El prefijo `docker.io` del
Quadlet identifica únicamente el registro OCI que publica la imagen oficial de
PostgreSQL; el runtime local es Podman.

## Instalación y arranque

Desde la raíz del repositorio:

```bash
bash infra/local/quadlet/install.sh install
```

El instalador:

1. genera contraseñas locales aleatorias;
2. las guarda con permisos `0600` fuera del repositorio;
3. instala el Quadlet rootless y los SQL de inicialización;
4. crea el volumen `poliredi-postgres-mvp1-data`;
5. inicia PostgreSQL y espera su healthcheck.

PostgreSQL se publica solo en `127.0.0.1:55432`, para evitar colisiones con
instancias locales de MVP2 que puedan usar `5432`. El instalador deja ese mismo
puerto en el `DATABASE_URL` del backend.

Los archivos secretos quedan en:

- `~/.config/containers/systemd/poliredi-postgres.env`;
- `~/.config/poli-redi/backend.env`.

No copies esos archivos al repositorio. Para iniciar el backend desde Linux o
WSL, carga sus variables en la sesión:

```bash
set -a
source "${XDG_CONFIG_HOME:-$HOME/.config}/poli-redi/backend.env"
set +a
cd backend
go run ./cmd
```

El runtime `MVP_SCOPE=mvp1` usa PostgreSQL para autenticación local, usuarios,
recursos, actividades, política vigente, disponibilidad, reservas individuales,
detalle, historial propio y cancelación. Las rutas posteriores permanecen fuera
del router local; sus repositorios T-SQL se conservan temporalmente como deuda
aislada y no se ejecutan en este perfil.

## Operación local

```bash
bash infra/local/quadlet/install.sh status
bash infra/local/quadlet/install.sh logs
bash infra/local/quadlet/install.sh stop
bash infra/local/quadlet/install.sh start
```

Después de instalar, valida esquema, seed, solapes, `OPEN_USE`, bloqueos y
cancelación mediante una transacción que siempre termina en `ROLLBACK`:

```bash
bash infra/local/quadlet/verify-mvp1.sh
```

El seed es repetible y no crea reservas históricas. Puede reaplicarse así:

```bash
podman exec poliredi-postgres-mvp1 \
  psql -v ON_ERROR_STOP=1 -U poliredi_owner -d poliredi \
  -f /docker-entrypoint-initdb.d/40_mvp1_seed.sql
```

## Archivos aplicados a un volumen nuevo

1. `bootstrap/PG16_0000_local_role.sql`: rol `poliredi_app` no-superusuario.
2. `migrations/PG16_0001_mvp1_baseline.sql`: tablas mínimas del MVP1.
3. `migrations/PG16_0002_mvp1_indexes.sql`: índices y exclusiones de solape.
4. `migrations/PG16_0003_mvp1_invariants.sql`: triggers e invariantes.
5. `seed/PG16_seed_mvp1.sql`: usuarios, recursos, actividades y política local.

## Integridad relevante

- Los instantes usan `timestamptz`; las reglas institucionales se evalúan en
  `America/Santiago`.
- Los rangos son semiabiertos `[inicio, fin)`: una reserva puede comenzar justo
  cuando termina otra.
- Un recurso `RESERVABLE` no acepta reservas activas solapadas.
- Un recurso `OPEN_USE` acepta concurrencia entre personas distintas, pero una
  persona no puede solapar sus propias reservas.
- Reservas y bloqueos usan el mismo advisory lock transaccional por recurso.
- La API se conecta como `poliredi_app`, sin superusuario, creación de bases ni
  creación de roles.

Los scripts del entrypoint solo se aplican al crear un volumen vacío. Para
reiniciar la base debe detenerse el servicio y eliminarse explícitamente
`poliredi-postgres-mvp1-data`; esa operación destruye únicamente los datos
locales del MVP1 y no forma parte del flujo normal.
