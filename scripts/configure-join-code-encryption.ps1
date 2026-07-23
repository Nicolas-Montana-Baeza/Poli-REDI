<#
.SYNOPSIS
Configura de forma segura el cifrado de códigos de invitación en backend/.env.

.DESCRIPTION
Sin parámetros reutiliza una configuración válida existente. Si las variables
no existen, genera una clave AES de 32 bytes y activa la versión 1.
Use -Rotate para agregar una versión nueva, conservando las anteriores.

.PARAMETER Rotate
Genera una clave nueva, agrega la siguiente versión y la deja activa.

.PARAMETER Repair
Reemplaza una configuración inválida por un llavero nuevo de versión 1. Conserva
el resto del archivo y crea un backup. No se puede combinar con -Rotate.

.EXAMPLE
.\scripts\configure-join-code-encryption.ps1

.EXAMPLE
.\scripts\configure-join-code-encryption.ps1 -Rotate

.EXAMPLE
.\scripts\configure-join-code-encryption.ps1 -Repair
#>
[CmdletBinding()]
param(
    [switch]$Rotate,
    [switch]$Repair
)

$ErrorActionPreference = 'Stop'
$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$backendDirectory = Join-Path $repoRoot 'backend'
$envPath = Join-Path $backendDirectory '.env'
$examplePath = Join-Path $backendDirectory '.env.example'

function Stop-Config([string]$Message) {
    throw "No se modificó backend/.env. $Message"
}

if ($Rotate -and $Repair) {
    Stop-Config '-Rotate y -Repair no se pueden combinar.'
}

function Assert-NotReparsePoint([string]$Path, [string]$Label) {
    if (Test-Path -LiteralPath $Path) {
        $attributes = [System.IO.File]::GetAttributes($Path)
        if (($attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
            Stop-Config "$Label no puede ser un enlace simbólico ni otro punto de reanálisis."
        }
    }
}

if (-not (Test-Path -LiteralPath $backendDirectory -PathType Container)) {
    Stop-Config 'No se encontró el directorio backend.'
}
Assert-NotReparsePoint $backendDirectory 'El directorio backend'
Assert-NotReparsePoint $envPath 'backend/.env'

if (Get-Command git -ErrorAction SilentlyContinue) {
    & git -C $repoRoot check-ignore --quiet -- backend/.env
    if ($LASTEXITCODE -ne 0) {
        Stop-Config 'backend/.env no está ignorado por Git.'
    }
    & git -C $repoRoot check-ignore --quiet -- backend/.env.backup-verificacion
    if ($LASTEXITCODE -ne 0) {
        Stop-Config 'Los backups backend/.env.backup-* no están ignorados por Git.'
    }
}

$utf8 = New-Object System.Text.UTF8Encoding($false)
$sourceExisted = Test-Path -LiteralPath $envPath -PathType Leaf
if ($sourceExisted) {
    $content = [System.IO.File]::ReadAllText($envPath)
} elseif (Test-Path -LiteralPath $examplePath -PathType Leaf) {
    $content = [System.IO.File]::ReadAllText($examplePath)
} else {
    $content = "# Configuración local de Poli-REDI`r`n"
}

$newline = if ($content.Contains("`r`n")) { "`r`n" } else { "`n" }
$hadTrailingNewline = $content -match "(\r\n|\n)$"
$lineList = New-Object 'System.Collections.Generic.List[string]'
if ($content.Length -gt 0) {
    foreach ($line in [System.Text.RegularExpressions.Regex]::Split($content, "\r?\n")) {
        [void]$lineList.Add($line)
    }
    if ($hadTrailingNewline -and $lineList.Count -gt 0) {
        $lineList.RemoveAt($lineList.Count - 1)
    }
}
$lines = $lineList
$versionIndexes = @()
$keysIndexes = @()
for ($index = 0; $index -lt $lines.Count; $index++) {
    if ($lines[$index] -match '^\s*JOIN_CODE_KEY_VERSION\s*=') { $versionIndexes += $index }
    if ($lines[$index] -match '^\s*JOIN_CODE_ENCRYPTION_KEYS\s*=') { $keysIndexes += $index }
}
if ($versionIndexes.Count -gt 1 -or $keysIndexes.Count -gt 1) {
    Stop-Config 'Las variables JOIN_CODE_* están duplicadas.'
}

function Get-SettingValue([int[]]$Indexes) {
    if ($Indexes.Count -eq 0) { return $null }
    return (($lines[$Indexes[0]] -split '=', 2)[1]).Trim()
}

$versionText = Get-SettingValue $versionIndexes
$keysText = Get-SettingValue $keysIndexes
$hasVersion = $null -ne $versionText -and $versionText -ne ''
$hasKeys = $null -ne $keysText -and $keysText -ne ''
if (-not $Repair -and ($hasVersion -xor $hasKeys)) {
    Stop-Config 'La configuración está incompleta: ambas variables deben estar presentes.'
}

$keyring = [ordered]@{}
$activeVersion = 0
if (-not $Repair -and $hasVersion -and $hasKeys) {
    if ($versionText -notmatch '^[1-9][0-9]*$') {
        Stop-Config 'JOIN_CODE_KEY_VERSION debe ser un entero positivo sin comillas, espacios ni prefijo.'
    }
    $activeVersion = [int]$versionText
    foreach ($entry in ($keysText -split ',')) {
        if ($entry -notmatch '^([1-9][0-9]*):([A-Za-z0-9+/]+={0,2})$') {
            Stop-Config 'JOIN_CODE_ENCRYPTION_KEYS debe usar version:base64, separado por comas.'
        }
        $entryVersion = [int]$Matches[1]
        if ($keyring.Contains($entryVersion)) {
            Stop-Config "La versión $entryVersion está duplicada en el llavero."
        }
        try { $decoded = [Convert]::FromBase64String($Matches[2]) } catch {
            Stop-Config "La versión $entryVersion no contiene Base64 válido."
        }
        if ($decoded.Length -ne 32) {
            Stop-Config "La versión $entryVersion no decodifica exactamente 32 bytes."
        }
        $keyring.Add($entryVersion, $Matches[2])
    }
    if (-not $keyring.Contains($activeVersion)) {
        Stop-Config 'JOIN_CODE_KEY_VERSION no existe en JOIN_CODE_ENCRYPTION_KEYS.'
    }
}

if (-not $Repair -and $hasVersion -and -not $Rotate) {
    Write-Host "La configuración de cifrado existente es válida. Versión activa: $activeVersion. No se realizaron cambios."
    exit 0
}

$newVersion = if ($Repair -or $keyring.Count -eq 0) { 1 } else { ([int[]]$keyring.Keys | Measure-Object -Maximum).Maximum + 1 }
$bytes = New-Object byte[] 32
$rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
try { $rng.GetBytes($bytes) } finally { $rng.Dispose() }
$keyring.Add($newVersion, [Convert]::ToBase64String($bytes))
$activeVersion = $newVersion
$keyringText = (($keyring.GetEnumerator() | ForEach-Object { "$($_.Key):$($_.Value)" }) -join ',')

function Set-Setting([string]$Name, [string]$Value, [int[]]$Indexes) {
    if ($Indexes.Count -eq 1) {
        $lines[$Indexes[0]] = "$Name=$Value"
    } else {
        [void]$lines.Add("$Name=$Value")
    }
}
Set-Setting 'JOIN_CODE_ENCRYPTION_KEYS' $keyringText $keysIndexes
Set-Setting 'JOIN_CODE_KEY_VERSION' ([string]$activeVersion) $versionIndexes
$newContent = [string]::Join($newline, $lines)
if ($hadTrailingNewline -or $content.Length -eq 0) {
    $newContent += $newline
}

$backupPath = $null
if ($sourceExisted) {
    $stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
    $backupPath = "$envPath.backup-$stamp-$([Guid]::NewGuid().ToString('N').Substring(0,8))"
}
$temporaryPath = "$envPath.tmp-$([Guid]::NewGuid().ToString('N'))"
try {
    [System.IO.File]::WriteAllText($temporaryPath, $newContent, $utf8)
    Assert-NotReparsePoint $backendDirectory 'El directorio backend'
    Assert-NotReparsePoint $envPath 'backend/.env'
    if ([System.IO.File]::Exists($envPath)) {
        [System.IO.File]::Replace($temporaryPath, $envPath, $backupPath, $true)
    } else {
        [System.IO.File]::Move($temporaryPath, $envPath)
    }
} finally {
    [System.IO.File]::Delete($temporaryPath)
}
if ($backupPath) {
    Write-Host "Backup recuperable creado: $([System.IO.Path]::GetFileName($backupPath))"
}
Write-Host "Configuración guardada de forma atómica. Versión activa: $activeVersion. No se mostraron claves."
