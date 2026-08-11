# Cambios realizados en la documentación

## 1. Estructura

- Se eliminaron 15 duplicados exactos con sufijo `(1)`.
- Se consolidaron más de veinte archivos en once documentos canónicos.
- Se separaron estado, arquitectura, requisitos, base de datos, operación, flujos, calidad y backlog.
- Se movió la revisión inicial a histórico.
- Se preservó el backlog completo y el catálogo detallado de requisitos en `referencia/`.

## 2. Coherencia

- Se priorizó la evidencia del acta actualizada al 2026-08-04.
- Se corrigieron estados antiguos que todavía indicaban frontend grupal o expiración como no implementados.
- Se mantuvo explícita la diferencia entre verificación local e integración online.
- Se registró que `007` y `008` continúan pendientes en Azure SQL.
- La migración `009` generada en el paquete mejorado de base se describe como propuesta, no como desplegada.

## 3. Seguridad operacional

- Se reforzó la prohibición de usar scripts destructivos para recuperar la base única.
- Se eliminaron valores concretos de URLs y nombres de recursos cloud de la guía canónica.
- Se mantuvo `DEV_AUTH_ENABLED=false` como regla para nube.
- Se agregó jerarquía de fuentes y definición de terminado.

## 4. Trazabilidad

- Se añadió mapa entre documentos anteriores y documentos canónicos.
- Se consolidaron criterios de estado y cierre de MVP.
- Se integraron pruebas, checklists y pendientes reales.
