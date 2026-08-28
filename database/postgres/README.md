# PostgreSQL 16 para Poli-REDI

Este directorio contiene la linea de persistencia PostgreSQL vigente.

## Evolucion

La baseline inicial corresponde a MVP1:

```txt
PG16_0001_mvp1_baseline.sql
PG16_0002_mvp1_indexes.sql
PG16_0003_mvp1_invariants.sql
```

MVP2 agrega:

```txt
PG16_0004_mvp2_group_participants.sql
PG16_0005_mvp2_institutional_scheduling.sql
PG16_0006_mvp2_institutional_availability.sql
PG16_0007_mvp2_schedule_exceptions.sql
PG16_0008_mvp2_schedule_exception_availability.sql
PG16_0009_full_notifications.sql
PG16_0010_mvp2_group_resource_rules.sql
```

Los scripts T-SQL historicos ubicados directamente bajo `database/` no pertenecen a esta linea de migraciones.

## Requisitos locales

- Linux o WSL.
- Podman rootless.
- cgroup v2.
- systemd de usuario.
- Podman Quadlet.

## Instalacion local

Desde la raiz:

```bash
bash infra/local/quadlet/install.sh install
```

El instalador genera credenciales aleatorias fuera del repositorio y levanta PostgreSQL 16.

PostgreSQL se publica en:

```txt
127.0.0.1:55432
```

Los secretos quedan en:

```txt
~/.config/containers/systemd/poliredi-postgres.env
~/.config/poli-redi/backend.env
```

## Provisionamiento automatico actual

El instalador Quadlet aplica automaticamente:

1. `bootstrap/PG16_0000_local_role.sql`
2. `PG16_0001` a `PG16_0003`;
3. `seed/PG16_seed_mvp1.sql`;
4. `PG16_0004` a `PG16_0010`.

El `MVP_SCOPE` controla las rutas que expone la API; no reduce el esquema
fisico inicializado en PostgreSQL.

## Aplicar PG16_0010 a un volumen existente

Los scripts de inicializacion solo se ejecutan al crear un volumen nuevo. En
un volumen local que ya contiene `PG16_0004` a `PG16_0009`, ejecutar desde la
raiz del repositorio:

```bash
podman exec -i poliredi-postgres-mvp1 \
  sh -c 'psql --username="$POSTGRES_USER" --dbname="$POSTGRES_DB" -v ON_ERROR_STOP=1' \
  < database/postgres/migrations/PG16_0010_mvp2_group_resource_rules.sql
```

La migracion usa una transaccion y falla completa ante cualquier precondicion
incumplida.

## Ejecutar backend

Con las variables generadas por Quadlet:

```bash
set -a
source "${XDG_CONFIG_HOME:-$HOME/.config}/poli-redi/backend.env"
set +a

cd backend
go run ./cmd
```

## Operacion

```bash
bash infra/local/quadlet/install.sh status
bash infra/local/quadlet/install.sh logs
bash infra/local/quadlet/install.sh stop
bash infra/local/quadlet/install.sh start
```

## Verificacion MVP1

```bash
bash infra/local/quadlet/verify-mvp1.sh
```

## Integridad relevante

- Fechas e instantes usan tipos PostgreSQL adecuados.
- Las reglas institucionales se interpretan en `America/Santiago`.
- Los rangos de agenda son semiabiertos.
- Los recursos `RESERVABLE` protegen solapes incompatibles.
- `OPEN_USE` admite concurrencia bajo sus reglas.
- La API utiliza un usuario de aplicacion sin privilegios administrativos globales.

Consultar `database/README.md` para la clasificacion completa de la persistencia vigente y los componentes legacy.


## Bootstrap local y migraciones vigentes

Una base local nueva creada por `infra/local/quadlet/install.sh` inicializa en
orden:

1. rol local;
2. `PG16_0001` a `PG16_0003`;
3. seed MVP1;
4. `PG16_0004` a `PG16_0010`.

El `MVP_SCOPE` controla la superficie HTTP expuesta, no la version fisica del
esquema. Mantener el esquema completo permite cambiar de `mvp1` a `mvp2` o
`full` sin reintroducir una base parcial.

Los volumenes PostgreSQL ya existentes no vuelven a ejecutar el directorio de
inicializacion de la imagen; cualquier migracion nueva debe aplicarse de forma
explicita al volumen existente.
