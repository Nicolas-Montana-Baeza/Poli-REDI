#!/usr/bin/env bash
set -Eeuo pipefail

readonly script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
readonly repository_root="$(cd -- "${script_dir}/../../.." && pwd)"
readonly container_name="poliredi-mvp2-verify-${UID}-$$"
readonly volume_name="poliredi-mvp2-verify-${UID}-$$"
readonly postgres_image="docker.io/library/postgres:16-alpine"

runtime_dir="$(mktemp -d "${TMPDIR:-/tmp}/poliredi-mvp2-init.XXXXXX")"
readonly frontend_runtime_dir="${runtime_dir}/frontend"
owner_password=""
app_password=""

# mktemp crea el directorio con modo 0700. El proceso postgres del contenedor
# necesita atravesarlo para leer los scripts montados como solo lectura. Los
# archivos no contienen secretos y se mantienen en 0644; las contrasenas se
# entregan exclusivamente mediante variables de entorno.
chmod 0755 "${runtime_dir}"

require_command() {
    if ! command -v "$1" >/dev/null 2>&1; then
        printf 'Falta el comando requerido: %s\n' "$1" >&2
        exit 1
    fi
}

generate_secret() {
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -hex 24
        return
    fi

    od -An -N24 -tx1 /dev/urandom | tr -d ' \n'
}

container_exists() {
    podman container exists "${container_name}" 2>/dev/null
}

cleanup() {
    local exit_code=$?

    trap - EXIT INT TERM

    if ((exit_code != 0)) && container_exists; then
        printf '\n=== Logs PostgreSQL efimero ===\n' >&2
        podman logs "${container_name}" >&2 || true
    fi

    podman rm --force "${container_name}" >/dev/null 2>&1 || true
    podman volume rm "${volume_name}" >/dev/null 2>&1 || true

    if [[ -n "${runtime_dir}" && -d "${runtime_dir}" ]]; then
        rm -rf -- "${runtime_dir}"
    fi

    exit "${exit_code}"
}

install_init_file() {
    local source_path=$1
    local target_name=$2

    install -m 0644 \
        "${repository_root}/${source_path}" \
        "${runtime_dir}/${target_name}"
}

wait_until_healthy() {
    local attempt health state

    for attempt in $(seq 1 90); do
        health="$(
            podman inspect \
                --format '{{if .State.Health}}{{.State.Health.Status}}{{end}}' \
                "${container_name}" 2>/dev/null || true
        )"

        if [[ "${health}" == "healthy" ]]; then
            return
        fi

        state="$(
            podman inspect \
                --format '{{.State.Status}}' \
                "${container_name}" 2>/dev/null || true
        )"

        if [[ "${state}" == "exited" || "${state}" == "stopped" ]]; then
            printf 'PostgreSQL efimero termino antes de quedar saludable.\n' >&2
            return 1
        fi

        sleep 2
    done

    printf 'PostgreSQL efimero no quedo saludable dentro del plazo.\n' >&2
    return 1
}

wait_until_initialized() {
    local attempt state initialization_log query_result

    for attempt in $(seq 1 90); do
        initialization_log="$(podman logs "${container_name}" 2>&1 || true)"

        if [[ "${initialization_log}" == *'PostgreSQL init process complete; ready for start up.'* ]]; then
            query_result="$(
                podman exec "${container_name}" \
                    psql \
                        --username=poliredi_owner \
                        --dbname=poliredi \
                        --no-psqlrc \
                        --tuples-only \
                        --no-align \
                        --command='SELECT 1' \
                    2>/dev/null || true
            )"

            if [[ "${query_result}" == "1" ]]; then
                return
            fi
        fi

        state="$(
            podman inspect \
                --format '{{.State.Status}}' \
                "${container_name}" 2>/dev/null || true
        )"

        if [[ "${state}" == "exited" || "${state}" == "stopped" ]]; then
            printf 'PostgreSQL efimero termino durante la inicializacion.\n' >&2
            return 1
        fi

        sleep 2
    done

    printf 'PostgreSQL efimero no completo las migraciones dentro del plazo.\n' >&2
    return 1
}

trap cleanup EXIT INT TERM

for command_name in podman go node npm install mktemp mkdir seq tar; do
    require_command "${command_name}"
done

if [[ "$(podman info --format '{{.Host.CgroupsVersion}}')" != "v2" ]]; then
    printf 'La verificacion requiere Podman rootless con cgroup v2.\n' >&2
    exit 1
fi

owner_password="$(generate_secret)"
app_password="$(generate_secret)"

install_init_file \
    'database/postgres/bootstrap/PG16_0000_local_role.sql' \
    '00_local_role.sql'
install_init_file \
    'database/postgres/migrations/PG16_0001_mvp1_baseline.sql' \
    '10_mvp1_baseline.sql'
install_init_file \
    'database/postgres/migrations/PG16_0002_mvp1_indexes.sql' \
    '20_mvp1_indexes.sql'
install_init_file \
    'database/postgres/migrations/PG16_0003_mvp1_invariants.sql' \
    '30_mvp1_invariants.sql'
install_init_file \
    'database/postgres/seed/PG16_seed_mvp1.sql' \
    '40_mvp1_seed.sql'
install_init_file \
    'database/postgres/migrations/PG16_0004_mvp2_group_participants.sql' \
    '50_mvp2_group_participants.sql'
install_init_file \
    'database/postgres/migrations/PG16_0005_mvp2_institutional_scheduling.sql' \
    '60_mvp2_institutional_scheduling.sql'
install_init_file \
    'database/postgres/migrations/PG16_0006_mvp2_institutional_availability.sql' \
    '70_mvp2_institutional_availability.sql'
install_init_file \
    'database/postgres/migrations/PG16_0007_mvp2_schedule_exceptions.sql' \
    '80_mvp2_schedule_exceptions.sql'
install_init_file \
    'database/postgres/migrations/PG16_0008_mvp2_schedule_exception_availability.sql' \
    '90_mvp2_schedule_exception_availability.sql'
install_init_file \
    'database/postgres/migrations/PG16_0009_full_notifications.sql' \
    '95_full_notifications.sql'
install_init_file \
    'database/postgres/migrations/PG16_0010_mvp2_group_resource_rules.sql' \
    '97_mvp2_group_resource_rules.sql'

printf '=== Creando PostgreSQL 16 efimero ===\n'
podman volume create "${volume_name}" >/dev/null

podman run --detach \
    --name "${container_name}" \
    --volume "${volume_name}:/var/lib/postgresql/data" \
    --volume "${runtime_dir}:/docker-entrypoint-initdb.d:ro,Z" \
    --publish '127.0.0.1::5432' \
    --env 'POSTGRES_DB=poliredi' \
    --env 'POSTGRES_USER=poliredi_owner' \
    --env "POSTGRES_PASSWORD=${owner_password}" \
    --env "POSTGRES_APP_PASSWORD=${app_password}" \
    --env 'TZ=America/Santiago' \
    --env 'PGTZ=America/Santiago' \
    --health-cmd \
        'pg_isready --username=poliredi_owner --dbname=poliredi' \
    --health-interval 2s \
    --health-timeout 5s \
    --health-retries 60 \
    --health-start-period 5s \
    "${postgres_image}" >/dev/null

wait_until_healthy
wait_until_initialized

port_mapping="$(podman port "${container_name}" 5432/tcp | tail -n 1)"
host_port="${port_mapping##*:}"

if [[ ! "${host_port}" =~ ^[0-9]+$ ]]; then
    printf 'No se pudo determinar el puerto efimero de PostgreSQL.\n' >&2
    exit 1
fi

readonly database_url="postgres://poliredi_app:${app_password}@127.0.0.1:${host_port}/poliredi?sslmode=disable"

printf '=== Verificando esquema y reglas MVP2 ===\n'
podman exec -i "${container_name}" \
    psql \
        --username=poliredi_owner \
        --dbname=poliredi \
        --no-psqlrc \
        --set=ON_ERROR_STOP=1 \
        --file=- \
    <"${repository_root}/database/postgres/check/PG16_verify_mvp2.sql"

printf '=== Ejecutando backend e integraciones contra la base efimera ===\n'
(
    cd "${repository_root}/backend"

    DATABASE_URL="${database_url}" \
    APP_TIMEZONE='America/Santiago' \
    MVP_SCOPE='mvp2' \
    POLIREDI_INTEGRATION='1' \
        go test ./... -p 1 -count=1

    DATABASE_URL="${database_url}" \
    APP_TIMEZONE='America/Santiago' \
    MVP_SCOPE='mvp2' \
        go vet ./...
)

printf '=== Ejecutando frontend MVP2 ===\n'
mkdir -p "${frontend_runtime_dir}"
(
    cd "${repository_root}/frontend"
    tar \
        --exclude='./node_modules' \
        --exclude='./dist' \
        --exclude='./.env' \
        --exclude='./.env.*' \
        -cf - \
        .
) | tar -xf - -C "${frontend_runtime_dir}"

(
    cd "${frontend_runtime_dir}"

    npm ci
    npm test

    VITE_MVP_SCOPE='mvp2' \
    VITE_API_BASE_URL='http://localhost:3000/api' \
        npm run build
)

printf '\nCIERRE AUTOMATIZADO MVP2: PASS\n'
printf 'La base, el contenedor y el volumen efimeros se eliminaran ahora.\n'
