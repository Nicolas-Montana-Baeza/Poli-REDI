#!/usr/bin/env bash
set -Eeuo pipefail

readonly script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
readonly repository_root="$(cd -- "${script_dir}/../../.." && pwd)"
readonly quadlet_dir="${XDG_CONFIG_HOME:-${HOME}/.config}/containers/systemd"
readonly runtime_dir="${quadlet_dir}/poliredi-postgres-init"
readonly postgres_env="${quadlet_dir}/poliredi-postgres.env"
readonly backend_env="${XDG_CONFIG_HOME:-${HOME}/.config}/poli-redi/backend.env"
readonly service_name="poliredi-postgres.service"

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

install_files() {
    local owner_password app_password

    install -d -m 0700 "${quadlet_dir}" "$(dirname -- "${backend_env}")"
    install -d -m 0755 "${runtime_dir}"

    if [[ ! -f "${postgres_env}" ]]; then
        owner_password="$(generate_secret)"
        app_password="$(generate_secret)"
        umask 077
        printf '%s\n' \
            'POSTGRES_DB=poliredi' \
            'POSTGRES_USER=poliredi_owner' \
            "POSTGRES_PASSWORD=${owner_password}" \
            "POSTGRES_APP_PASSWORD=${app_password}" \
            'TZ=America/Santiago' \
            'PGTZ=America/Santiago' >"${postgres_env}"
        printf '%s\n' \
            "DATABASE_URL=postgres://poliredi_app:${app_password}@127.0.0.1:5432/poliredi?sslmode=disable" \
            'PORT=3000' \
            'APP_TIMEZONE=America/Santiago' \
            'CORS_ALLOWED_ORIGINS=http://localhost:5173' \
			'DEV_AUTH_ENABLED=true' \
			'MVP_SCOPE=mvp1' >"${backend_env}"
        chmod 0600 "${postgres_env}" "${backend_env}"
    fi

	if [[ -f "${backend_env}" ]] && ! grep -q '^MVP_SCOPE=' "${backend_env}"; then
		printf '%s\n' 'MVP_SCOPE=mvp1' >>"${backend_env}"
	fi

    install -m 0644 "${script_dir}/poliredi-postgres.container" "${quadlet_dir}/poliredi-postgres.container"
    install -m 0644 "${script_dir}/poliredi-postgres-data.volume" "${quadlet_dir}/poliredi-postgres-data.volume"
    install -m 0644 "${repository_root}/database/postgres/bootstrap/PG16_0000_local_role.sql" "${runtime_dir}/00_local_role.sql"
    install -m 0644 "${repository_root}/database/postgres/migrations/PG16_0001_mvp1_baseline.sql" "${runtime_dir}/10_mvp1_baseline.sql"
    install -m 0644 "${repository_root}/database/postgres/migrations/PG16_0002_mvp1_indexes.sql" "${runtime_dir}/20_mvp1_indexes.sql"
    install -m 0644 "${repository_root}/database/postgres/migrations/PG16_0003_mvp1_invariants.sql" "${runtime_dir}/30_mvp1_invariants.sql"
    install -m 0644 "${repository_root}/database/postgres/seed/PG16_seed_mvp1.sql" "${runtime_dir}/40_mvp1_seed.sql"

    systemctl --user daemon-reload
    printf 'Quadlet instalado. Variables de la API: %s\n' "${backend_env}"
}

start_service() {
    podman pull docker.io/library/postgres:16-alpine
    systemctl --user start "${service_name}"
    systemctl --user --no-pager --full status "${service_name}"
}

require_command podman
require_command systemctl

if [[ "$(podman info --format '{{.Host.CgroupsVersion}}')" != "v2" ]]; then
    printf 'Podman Quadlet rootless requiere cgroup v2.\n' >&2
    exit 1
fi

case "${1:-install}" in
    install)
        install_files
        start_service
        ;;
    start)
        systemctl --user start "${service_name}"
        ;;
    stop)
        systemctl --user stop "${service_name}"
        ;;
    status)
        systemctl --user --no-pager --full status "${service_name}"
        podman inspect --format '{{.State.Health.Status}}' poliredi-postgres-mvp1
        ;;
    logs)
        journalctl --user -u "${service_name}" -n 200 --no-pager
        ;;
    *)
        printf 'Uso: %s [install|start|stop|status|logs]\n' "$0" >&2
        exit 2
        ;;
esac
