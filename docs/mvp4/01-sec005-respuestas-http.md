# SEC-005 - Seguridad de respuestas HTTP

Fecha: 2026-11-12
MVP: 4
Estado: EN PROGRESO

## Objetivo

Evitar que respuestas HTTP expongan errores internos
de PostgreSQL, autenticacion o infraestructura.

## Fase A - Sanitizacion directa

Se eliminaron detalles tecnicos de respuestas HTTP
en los siguientes archivos:

- activities_handlers.go
- notifications_handlers.go
- reservations_handlers.go
- resources_handlers.go
- auth_middleware.go

Se eliminaron detalles de validacion JWT que no
deben entregarse al cliente.

Resultado:
- go test ./...: PASS
- go vet ./...: PASS
- git diff --check: PASS

## Fase B - Clasificacion de errores

Se introdujeron clasificadores de errores publicos
para crear y cancelar reservas.

Los errores de negocio conocidos conservan
respuestas controladas.

Los errores inesperados reciben mensajes genericos
con HTTP 500.

Pruebas:
- Creacion: 5/5 PASS
- Cancelacion: 4/4 PASS
- Suite Go: PASS

## Regresion PostgreSQL 16

Fecha: 2026-11-12

Validacion mediante verificador MVP3 efimero.

Resultados:
- NOTIF-001 HTTP E2E: PASS
- Politicas grupales: PASS
- Sedes y recursos: PASS
- Bloqueos administrativos: PASS
- Calendario de bloqueos: PASS
- Verificador completo: PASS
- Codigo de salida: 0

El contenedor y volumen efimeros fueron eliminados.

## Pendientes

- Probar respuestas HTTP 500 ante errores simulados.
- Verificar ausencia de detalles SQL/JWT en los endpoints.
- Revisar errores de negocio aun no clasificados.
- Definir codigos publicos estables.
- Registrar errores internos de forma segura.
- Ejecutar regresion final antes de cerrar SEC-005.

## Restricciones

No modificar la base de datos habitual ni
el despliegue MVP3 validado.

El requisito SEC-005 permanece abierto.
