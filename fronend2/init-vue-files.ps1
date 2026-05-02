# ===== CONFIG =====
$root = Join-Path (Get-Location) "frontend"

if (!(Test-Path $root)) {
    Write-Host "ERROR: No existe la carpeta frontend en $root"
    exit
}

Write-Host "Buscando archivos .vue en $root"

# ===== BUSCAR ARCHIVOS =====
$vueFiles = Get-ChildItem -Path $root -Recurse -Filter *.vue -File

if ($vueFiles.Count -eq 0) {
    Write-Host "No se encontraron archivos .vue"
    exit
}

Write-Host "Archivos encontrados:" $vueFiles.Count

# ===== TEMPLATE =====
$template = @"
<script setup>
</script>

<template>
  <div></div>
</template>
"@

# ===== PROCESAR =====
foreach ($file in $vueFiles) {

    $content = Get-Content $file.FullName -Raw

    if ([string]::IsNullOrWhiteSpace($content)) {
        Set-Content -Path $file.FullName -Value $template -Encoding UTF8
        Write-Host "Inicializado:" $file.FullName
    }
    else {
        Write-Host "Saltado (ya tiene contenido):" $file.FullName
    }
}

Write-Host "Proceso terminado"