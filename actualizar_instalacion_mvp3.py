#!/usr/bin/env python3
"""Sincroniza docs/01-instalacion-y-ejecucion.md con MVP3.

Solo modifica el Markdown indicado, crea un respaldo local y realiza
validaciones. No invoca Podman, Quadlet, PostgreSQL ni migraciones.

Ejecucion desde la raiz del repositorio:
    python3 /ruta/al/archivo/actualizar_instalacion_mvp3.py
"""

from __future__ import annotations

from datetime import datetime
from pathlib import Path
import os
import shutil
import subprocess
import sys
import tempfile


DOCUMENT = Path("docs/01-instalacion-y-ejecucion.md")
EXPECTED_BRANCH = "feature/mvp3"
EXPECTED_COMMIT = "2f8bd01"


def git(*arguments: str) -> str:
    return subprocess.check_output(["git", *arguments], text=True).strip()


def ensure(condition: bool, message: str) -> None:
    if not condition:
        raise RuntimeError("ABORTADO: " + message)


def main() -> None:
    ensure(Path(".git").exists(), "ejecuta este archivo desde la raiz de Poli-REDI")
    ensure(DOCUMENT.is_file(), f"no se encuentra {DOCUMENT}")
    ensure(git("branch", "--show-current") == EXPECTED_BRANCH, "rama inesperada")
    ensure(git("rev-parse", "--short", "HEAD") == EXPECTED_COMMIT, "commit base inesperado")
    ensure(
        subprocess.run(["git", "diff", "--quiet", "--", str(DOCUMENT)], check=False).returncode == 0,
        "la guia tiene modificaciones sin guardar en Git",
    )
    ensure(
        subprocess.run(["git", "diff", "--cached", "--quiet"], check=False).returncode == 0,
        "existen cambios preparados (staged)",
    )

    raw = DOCUMENT.read_bytes()
    encoding = "utf-8-sig" if raw.startswith(b"\xef\xbb\xbf") else "utf-8"
    original = raw.decode(encoding)
    line_ending = "\r\n" if "\r\n" in original else "\n"
    current = original.replace("\r\n", "\n")
    modifications: list[str] = []

    def replace_one(old: str, new: str, label: str) -> None:
        nonlocal current
        ensure(current.count(old) == 1, f"texto original no coincide en: {label}")
        current = current.replace(old, new, 1)
        modifications.append(label)

    def replace_section(start: str, end: str, replacement: str, label: str) -> None:
        nonlocal current
        ensure(current.count(start) == 1 and current.count(end) == 1,
               f"encabezados inesperados en: {label}")
        left = current.index(start)
        right = current.index(end, left + len(start))
        current = current[:left] + replacement.strip() + "\n\n" + current[right:]
        modifications.append(label)

    # 1. Fecha y 2. distincion de operaciones.
    replace_one(
        "Fecha de corte: 2026-09-22",
        "Fecha de corte documental: 2026-10-27\n"
        "Referencia: cierre tecnico MVP3 del 2026-10-24.",
        "fecha de corte",
    )
    replace_one(
        "- Microsoft Entra ID o autenticacion local controlada.",
        "- Microsoft Entra ID o autenticacion local controlada.\n\n"
        "Esta guia distingue tres operaciones independientes: preparar una base nueva, "
        "migrar una base existente y actualizar el backend/frontend. Desplegar codigo "
        "no aplica automaticamente migraciones PostgreSQL.",
        "alcance operativo",
    )

    # 3. El Quadlet NO es independiente del stack Compose existente.
    replace_section(
        "## Preparar PostgreSQL aislado para desarrollo",
        "## Estado de migraciones",
        """## Quadlet de desarrollo y precaucion con el volumen compartido

El repositorio conserva `infra/local/quadlet/install.sh` para crear una instancia
PostgreSQL de desarrollo. Sin embargo, la configuracion Quadlet y el stack
Podman Compose documentado referencian el **mismo volumen**:
`poliredi-postgres-mvp1-data` y el mismo puerto host `127.0.0.1:55432`.

Por tanto, **no es una instancia aislada del stack integrado**. No ejecutar
`bash infra/local/quadlet/install.sh install` mientras el stack utilice ese
volumen. Nunca iniciar dos servidores PostgreSQL sobre un mismo directorio de
datos ni borrar/recrear ese volumen como procedimiento de actualizacion.

En un directorio de datos inicialmente vacio, el instalador Quadlet copia el
bootstrap local, las migraciones `PG16_0001` a `PG16_0010` y el seed MVP1.
No contempla las cinco migraciones posteriores de MVP3.

Variables privadas del mecanismo Quadlet:

```txt
~/.config/containers/systemd/poliredi-postgres.env
~/.config/poli-redi/backend.env
```

Para pruebas reproducibles sin afectar la instancia habitual existe un
verificador **efimero de MVP2**. La provision completa de MVP3 con aislamiento,
respaldo y comprobaciones se mantiene pendiente en `INFRA-007`.
""",
        "seguridad del volumen compartido",
    )

    # 4. Inventario completo de archivos de migracion.
    replace_one(
        "PG16_0010_mvp2_group_resource_rules.sql\n```",
        """PG16_0010_mvp2_group_resource_rules.sql
PG16_0011_mvp3_reservation_venue_guard.sql
PG16_0012_mvp3_availability_block_overlap.sql
PG16_0013_mvp3_occurrence_revisions.sql
PG16_0014_mvp3_user_access_audit.sql
PG16_0015_mvp3_user_role_audit.sql
```""",
        "inventario de 15 migraciones",
    )

    # 5. Comportamiento verdadero del bootstrap y configuracion MVP.
    replace_one(
        """El instalador automatiza actualmente el bootstrap, `0001` a `0003`, el seed
MVP1 y `0004` a `0010` para todo volumen nuevo.

Por eso el instalador genera:

```env
MVP_SCOPE=mvp1
```

como valor seguro.

Un volumen existente no vuelve a ejecutar los archivos de inicializacion. En
ese caso las migraciones nuevas deben aplicarse explicitamente antes de usar
`MVP_SCOPE=mvp2`.""",
        """El instalador Quadlet automatiza la cadena `0001` a `0010` en un volumen
nuevo y genera `MVP_SCOPE=mvp1` como valor conservador.

Los scripts de inicializacion de PostgreSQL se ejecutan al inicializar un
directorio de datos vacio; no se repiten al reiniciar un volumen existente.

El backend reconoce los alcances `mvp1`, `mvp2`, `mvp3` y `full` mediante
`MVP_SCOPE`. El frontend selecciona la superficie funcional en compilacion con
`VITE_MVP_SCOPE`.

**Cambiar el scope no ejecuta migraciones.** Para habilitar MVP3 debe
comprobarse que `PG16_0011` a `PG16_0015` esten aplicadas, que sus
precondiciones y privilegios sean correctos y que el runtime sea compatible.
""".strip(),
        "bootstrap y scopes",
    )

    # 6. Distinguir operaciones, sin comandos que alteren volumenes.
    replace_one(
        "## Verificacion limpia MVP2",
        """## Diferenciar instalacion, migracion y despliegue

### Preparar una base PostgreSQL nueva

Requiere bootstrap, migraciones en orden, seed adecuado y validacion del
esquema. La automatizacion actual de Quadlet llega a MVP2; la de MVP3 esta
pendiente (`INFRA-007`).

### Actualizar una base PostgreSQL existente

Requiere inventariar el esquema aplicado, crear y verificar un respaldo,
revisar incompatibilidades y ejecutar migraciones incrementales mediante un
procedimiento controlado. No reutilizar scripts de inicializacion de volumen
vacio ni suponer que un archivo SQL versionado ya fue aplicado.

### Desplegar backend y frontend

Cambiar imagenes de backend, variables de entorno o archivos compilados del
frontend **no reinicializa PostgreSQL**. El despliegue requiere compatibilidad
con el esquema presente y comprobaciones de salud posteriores.

Consultar `docs/10-guia-redeploy.md` antes de cualquier intervencion operativa.

## Verificacion limpia MVP2""",
        "procedimientos separados",
    )

    # 7. El verificador actual NO prueba las migraciones de MVP3.
    replace_one(
        "El entorno efimero se elimina al terminar y no modifica\nla instancia habitual.",
        "El entorno efimero se elimina al terminar y no modifica\n"
        "la instancia habitual. Valida las migraciones `0001` a `0010` de MVP2, "
        "no la cadena completa MVP3.",
        "alcance de verificacion MVP2",
    )

    # 8. Alcance frontend conservador y 9. npm reproducible.
    replace_one(
        "VITE_APP_TIMEZONE=America/Santiago\n",
        "VITE_APP_TIMEZONE=America/Santiago\nVITE_MVP_SCOPE=mvp1\n",
        "scope frontend",
    )
    replace_one("npm install\nnpm run dev", "npm ci\nnpm run dev", "dependencias frontend")

    ensure(len(modifications) == 9, "se esperaban 9 modificaciones")
    ensure(current.count("```") % 2 == 0, "bloques de codigo Markdown desequilibrados")
    ensure(current.count("## Verificacion limpia MVP2") == 1, "seccion MVP2 duplicada")
    ensure(current.count("## Estado de migraciones") == 1, "seccion de migraciones duplicada")
    ensure(current != original.replace("\r\n", "\n"), "no hay cambios")

    inventory_header = "Migraciones PostgreSQL presentes:\n\n```txt\n"
    ensure(current.count(inventory_header) == 1, "inventario de migraciones ausente")
    inventory = current.split(inventory_header, 1)[1].split("\n```", 1)[0]
    documented = sorted(line.strip() for line in inventory.splitlines() if line.strip())
    migration_dir = Path("database/postgres/migrations")
    ensure(migration_dir.is_dir(), "no existe el directorio de migraciones")
    actual = sorted(path.name for path in migration_dir.glob("PG16_*.sql"))
    ensure(len(actual) == 15, f"se esperaban 15 migraciones reales y hay {len(actual)}")
    ensure(documented == actual, "el inventario documentado no coincide con el repositorio")

    for other in ("docs/10-guia-redeploy.md", "docs/mvp3/cierre-mvp3-prototipo.md"):
        ensure(Path(other).is_file(), f"falta referencia local {other}")

    backup_dir = (Path.home() / "backups/poliredi/documentacion"
                  / ("instalacion-mvp3-" + datetime.now().strftime("%Y%m%d-%H%M%S-%f")))
    backup_dir.mkdir(parents=True, exist_ok=False)
    backup_path = backup_dir / DOCUMENT.name
    shutil.copy2(DOCUMENT, backup_path)

    target_text = current.replace("\n", line_ending)
    encoded = target_text.encode(encoding)
    temp_file = None
    try:
        with tempfile.NamedTemporaryFile(mode="wb", dir=DOCUMENT.parent,
                                         prefix=".instalacion-mvp3-", delete=False) as tmp:
            temp_file = Path(tmp.name)
            tmp.write(encoded)
        shutil.copymode(DOCUMENT, temp_file)
        os.replace(temp_file, DOCUMENT)
    finally:
        if temp_file is not None:
            temp_file.unlink(missing_ok=True)

    print("GUIA DE INSTALACION ACTUALIZADA")
    print("Correcciones:", len(modifications))
    for label in modifications:
        print(" -", label)
    print("Migraciones documentadas:", len(documented))
    print("Respaldo:", backup_path)
    print("No se ejecuto SQL, Podman ni Quadlet.")
    print("\nVerificacion Git:")
    subprocess.run(["git", "diff", "--check"], check=True)
    subprocess.run(["git", "diff", "--stat"], check=True)
    subprocess.run(["git", "status", "--short"], check=True)


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, OSError, subprocess.CalledProcessError) as exc:
        print(str(exc), file=sys.stderr)
        sys.exit(1)
