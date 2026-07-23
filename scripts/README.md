# Scripts de configuración

## Cifrado de códigos de invitación

Desde la raíz del repositorio, en Windows PowerShell 5.1 o PowerShell moderno:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\configure-join-code-encryption.ps1
```

La primera ejecución crea `backend/.env` desde `.env.example` cuando sea
necesario, genera una clave AES de 32 bytes y activa la versión 1. Si ya existe
una configuración válida, no la cambia.

La rotación es siempre explícita:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\configure-join-code-encryption.ps1 -Rotate
```

La rotación conserva las claves anteriores para descifrar códigos existentes,
crea un backup recuperable y escribe el archivo de forma atómica. El script no
imprime claves. Ante duplicados o formatos inválidos, aborta sin modificar el
archivo.

Si los valores actuales son inválidos, la reparación debe solicitarse de forma
explícita:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\configure-join-code-encryption.ps1 -Repair
```

`-Repair` reemplaza o inserta únicamente `JOIN_CODE_ENCRYPTION_KEYS` y
`JOIN_CODE_KEY_VERSION`, genera un llavero nuevo de versión 1 y crea un backup
atómico con el contenido original. Conserva las demás líneas y comentarios, no
acepta variables `JOIN_CODE_*` duplicadas y no puede combinarse con `-Rotate`.
Sin `-Repair`, una configuración inválida continúa abortando sin cambios.

Los backups `backend/.env.backup-*` están ignorados de forma explícita por Git.
El script también rechaza `backend` o `.env` cuando son enlaces simbólicos o
puntos de reanálisis. Su modelo de seguridad supone un repositorio local
confiable: no pretende resistir a otro proceso con privilegios que reemplace el
directorio durante la operación.

Prueba aislada, sin tocar el `.env` real:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\test-configure-join-code-encryption.ps1
```
