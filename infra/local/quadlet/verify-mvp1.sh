#!/usr/bin/env bash
set -Eeuo pipefail

readonly script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
readonly repository_root="$(cd -- "${script_dir}/../../.." && pwd)"
readonly container_name="poliredi-postgres-mvp1"

if ! command -v podman >/dev/null 2>&1; then
    printf 'Podman no esta disponible. Ejecuta este script dentro de Linux o WSL.\n' >&2
    exit 1
fi

if ! podman container exists "${container_name}"; then
    printf 'PostgreSQL no esta iniciado: el contenedor %s no existe. Ejecuta install.sh install y revisa los logs.\n' "${container_name}" >&2
    exit 1
fi

health_status="$(podman inspect --format '{{.State.Health.Status}}' "${container_name}" 2>/dev/null || true)"

if [[ "${health_status}" != "healthy" ]]; then
    printf 'PostgreSQL no esta saludable. Ejecuta install.sh install y revisa los logs.\n' >&2
    exit 1
fi

podman exec -i "${container_name}" \
    psql -v ON_ERROR_STOP=1 -U poliredi_owner -d poliredi \
    <"${repository_root}/database/postgres/check/PG16_verify_mvp1.sql"

printf 'Verificacion transaccional terminada; el script hizo ROLLBACK de sus datos de prueba.\n'
