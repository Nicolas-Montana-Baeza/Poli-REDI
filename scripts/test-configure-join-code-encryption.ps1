$ErrorActionPreference = 'Stop'
$sourceScript = Join-Path $PSScriptRoot 'configure-join-code-encryption.ps1'
$testRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("poli-redi-env-test-" + [Guid]::NewGuid().ToString('N'))
$utf8 = New-Object System.Text.UTF8Encoding($false)

function Write-Exact([string]$Path, [string]$Content) {
    [System.IO.File]::WriteAllText($Path, $Content, $utf8)
}

function Invoke-Configurator([switch]$Rotate, [switch]$Repair, [switch]$ExpectFailure) {
    $arguments = @('-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', $script)
    if ($Rotate) { $arguments += '-Rotate' }
    if ($Repair) { $arguments += '-Repair' }
    $oldPreference = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try {
        $result = & powershell.exe @arguments 2>&1
        $exitCode = $LASTEXITCODE
    } finally {
        $ErrorActionPreference = $oldPreference
    }
    if ($ExpectFailure) {
        if ($exitCode -eq 0) { throw "se esperaba un error, pero terminó correctamente: $result" }
    } elseif ($exitCode -ne 0) {
        throw "el configurador falló: $result"
    }
    return @($result)
}

function Assert-GeneratedKey([string]$Content) {
    $matches = [regex]::Matches($Content, '(?m)^JOIN_CODE_ENCRYPTION_KEYS=(.+)$')
    if ($matches.Count -ne 1) { throw 'JOIN_CODE_ENCRYPTION_KEYS no aparece exactamente una vez' }
    foreach ($entry in ($matches[0].Groups[1].Value.Trim() -split ',')) {
        $parts = $entry -split ':', 2
        if ($parts.Count -ne 2) { throw 'entrada de llavero inválida' }
        try { $decoded = [Convert]::FromBase64String($parts[1]) } catch { throw 'la clave generada no es Base64 válido' }
        if ($decoded.Length -ne 32) { throw 'la clave generada no decodifica 32 bytes' }
    }
}

function Assert-UniqueSettings([string]$Content) {
    if ([regex]::Matches($Content, '(?m)^JOIN_CODE_ENCRYPTION_KEYS=').Count -ne 1) {
        throw 'JOIN_CODE_ENCRYPTION_KEYS está ausente o duplicada'
    }
    if ([regex]::Matches($Content, '(?m)^JOIN_CODE_KEY_VERSION=').Count -ne 1) {
        throw 'JOIN_CODE_KEY_VERSION está ausente o duplicada'
    }
}

function Remove-JoinSettings([string]$Content) {
    return [regex]::Replace($Content, '(?m)^JOIN_CODE_(?:ENCRYPTION_KEYS|KEY_VERSION)=.*(?:\r?\n|$)', '')
}

function Assert-Repair([string]$Original) {
    Write-Exact $envPath $Original
    $backupCount = @(Get-ChildItem (Join-Path $testRoot 'backend') -Filter '.env.backup-*').Count
    $output = Invoke-Configurator -Repair
    $repaired = [System.IO.File]::ReadAllText($envPath)
    Assert-UniqueSettings $repaired
    Assert-GeneratedKey $repaired
    if ($repaired -notmatch '(?m)^JOIN_CODE_KEY_VERSION=1$') { throw 'la reparación no activó la versión 1' }
    if ((Remove-JoinSettings $repaired) -ne (Remove-JoinSettings $Original)) {
        throw 'la reparación modificó contenido ajeno a JOIN_CODE_*'
    }
    $backups = @(Get-ChildItem (Join-Path $testRoot 'backend') -Filter '.env.backup-*' | Sort-Object LastWriteTimeUtc)
    if ($backups.Count -ne ($backupCount + 1)) { throw 'la reparación no creó exactamente un backup' }
    if ([System.IO.File]::ReadAllText($backups[-1].FullName) -ne $Original) { throw 'el backup no contiene el original byte-equivalente' }
    if (($output -join "`n") -match '[A-Za-z0-9+/]{40,}={0,2}') { throw 'la reparación filtró una clave' }
}

try {
    New-Item -ItemType Directory -Path (Join-Path $testRoot 'scripts'),(Join-Path $testRoot 'backend') | Out-Null
    Copy-Item $sourceScript (Join-Path $testRoot 'scripts\configure-join-code-encryption.ps1')
    Write-Exact (Join-Path $testRoot '.gitignore') "backend/.env`nbackend/.env.backup-*`n!backend/.env.example`n"
    Write-Exact (Join-Path $testRoot 'backend\.env.example') "PORT=3000`n# comentario`nJOIN_CODE_ENCRYPTION_KEYS=`nJOIN_CODE_KEY_VERSION=`nOTHER=value`n"
    & git -C $testRoot init --quiet
    $script = Join-Path $testRoot 'scripts\configure-join-code-encryption.ps1'
    $envPath = Join-Path $testRoot 'backend\.env'

    & git -C $testRoot check-ignore --quiet -- backend/.env
    if ($LASTEXITCODE -ne 0) { throw 'backend/.env no está ignorado en la prueba real de Git' }
    & git -C $testRoot check-ignore --quiet -- backend/.env.backup-prueba
    if ($LASTEXITCODE -ne 0) { throw 'backend/.env.backup-* no está ignorado en la prueba real de Git' }
    & git -C $testRoot check-ignore --quiet -- backend/.env.example
    if ($LASTEXITCODE -eq 0) { throw 'backend/.env.example fue ignorado por error' }

    $output = Invoke-Configurator
    $first = [System.IO.File]::ReadAllText($envPath)
    Assert-UniqueSettings $first
    Assert-GeneratedKey $first
    if ($first -notmatch 'JOIN_CODE_KEY_VERSION=1' -or $first -notmatch 'OTHER=value' -or $first -notmatch '# comentario') {
        throw 'la creación no preservó el archivo'
    }
    if (($output -join "`n") -match '[A-Za-z0-9+/]{40,}={0,2}') { throw 'la salida filtró una clave' }

    Invoke-Configurator | Out-Null
    if ([System.IO.File]::ReadAllText($envPath) -ne $first) { throw 'la ejecución por defecto rotó o reescribió' }

    Invoke-Configurator -Rotate | Out-Null
    $rotated = [System.IO.File]::ReadAllText($envPath)
    Assert-GeneratedKey $rotated
    if ($rotated -notmatch 'JOIN_CODE_KEY_VERSION=2' -or $rotated -notmatch 'JOIN_CODE_ENCRYPTION_KEYS=1:.+,2:.+') {
        throw 'la rotación no preservó la clave anterior'
    }
    if (@(Get-ChildItem (Join-Path $testRoot 'backend') -Filter '.env.backup-*').Count -lt 1) { throw 'no se creó backup' }

    $keyLine = ([regex]::Match($first, '(?m)^JOIN_CODE_ENCRYPTION_KEYS=.+$')).Value.Trim()
    $duplicate = "$keyLine`nJOIN_CODE_KEY_VERSION=1`nJOIN_CODE_KEY_VERSION=1`nOTHER=keep`n"
    Write-Exact $envPath $duplicate
    Invoke-Configurator -ExpectFailure | Out-Null
    if ([System.IO.File]::ReadAllText($envPath) -ne $duplicate) { throw 'un archivo con variables duplicadas fue modificado' }

    $invalid = "JOIN_CODE_KEY_VERSION=`"1`"`nJOIN_CODE_ENCRYPTION_KEYS=1:bad`nOTHER=keep`n"
    Write-Exact $envPath $invalid
    Invoke-Configurator -ExpectFailure | Out-Null
    if ([System.IO.File]::ReadAllText($envPath) -ne $invalid) { throw 'una configuración inválida fue modificada' }

    Assert-Repair "HEAD=keep`nJOIN_CODE_KEY_VERSION=1:key`nJOIN_CODE_ENCRYPTION_KEYS=bad`n# comentario`nTAIL=keep`n"
    Assert-Repair "HEAD=keep`nJOIN_CODE_KEY_VERSION=`"1`"`nJOIN_CODE_ENCRYPTION_KEYS=1:bad`nTAIL=keep`n"
    Assert-Repair "HEAD=keep`nJOIN_CODE_KEY_VERSION= 1 `nJOIN_CODE_ENCRYPTION_KEYS=1:bad`nTAIL=keep`n"
    Assert-Repair "HEAD=keep`nJOIN_CODE_KEY_VERSION=1`nTAIL=keep`n"
    Assert-Repair "HEAD=keep`nJOIN_CODE_ENCRYPTION_KEYS=`nJOIN_CODE_KEY_VERSION=`nTAIL=keep`n"

    $duplicateRepair = "JOIN_CODE_KEY_VERSION=bad`nJOIN_CODE_KEY_VERSION=worse`nJOIN_CODE_ENCRYPTION_KEYS=bad`nOTHER=keep`n"
    Write-Exact $envPath $duplicateRepair
    Invoke-Configurator -Repair -ExpectFailure | Out-Null
    if ([System.IO.File]::ReadAllText($envPath) -ne $duplicateRepair) { throw 'la reparación de duplicados ambiguos modificó el archivo' }

    $beforeCombined = [System.IO.File]::ReadAllText($envPath)
    Invoke-Configurator -Repair -Rotate -ExpectFailure | Out-Null
    if ([System.IO.File]::ReadAllText($envPath) -ne $beforeCombined) { throw '-Repair combinado con -Rotate modificó el archivo' }

    Write-Exact $envPath 'ONLY=value'
    Invoke-Configurator | Out-Null
    $singleNoNewline = [System.IO.File]::ReadAllText($envPath)
    Assert-UniqueSettings $singleNoNewline
    Assert-GeneratedKey $singleNoNewline
    if ([regex]::Matches($singleNoNewline, '(?m)^ONLY=value$').Count -ne 1) { throw 'la línea única sin salto fue duplicada' }
    if ($singleNoNewline.EndsWith("`n")) { throw 'se agregó un salto final que no existía' }

    Write-Exact $envPath "ONLY=value`n"
    Invoke-Configurator | Out-Null
    $singleWithNewline = [System.IO.File]::ReadAllText($envPath)
    Assert-UniqueSettings $singleWithNewline
    Assert-GeneratedKey $singleWithNewline
    if ([regex]::Matches($singleWithNewline, '(?m)^ONLY=value$').Count -ne 1) { throw 'la línea única con salto fue duplicada' }
    if (-not $singleWithNewline.EndsWith("`n")) { throw 'no se preservó el salto final' }

    Write-Exact $envPath ''
    Invoke-Configurator | Out-Null
    $emptyResult = [System.IO.File]::ReadAllText($envPath)
    Assert-UniqueSettings $emptyResult
    Assert-GeneratedKey $emptyResult
    if ($emptyResult.StartsWith("`n")) { throw 'el archivo vacío generó una línea inicial espuria' }

    Write-Host 'Pruebas del configurador: OK'
} finally {
    if ([System.IO.Directory]::Exists($testRoot)) { [System.IO.Directory]::Delete($testRoot, $true) }
}
