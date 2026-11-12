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

## Fase C - Pruebas HTTP de errores internos

Fecha: 2026-11-12

Se validaron ocho endpoints mediante Fiber.App.Test
con una conexion PostgreSQL deliberadamente cerrada.

Escenarios:
- Listar reservas.
- Consultar reservas propias.
- Consultar disponibilidad.
- Consultar recursos.
- Consultar actividades.
- Consultar notificaciones.
- Crear reserva.
- Cancelar reserva.

Resultados:
- 8/8 pruebas HTTP: PASS.
- Respuestas HTTP 500 con mensajes controlados.
- Sin campos tecnicos adicionales en las respuestas.
- go test ./...: PASS.
- go vet ./...: PASS.
- git diff --check: PASS.

La prueba no utiliza una base de datos real
ni valida el flujo real de Microsoft Entra ID.

## Fase D1 y D2 - Contrato HTTP y diagnostico seguro

Fecha: 2026-11-12
Estado: VALIDADO PARCIALMENTE

Se incorporo un contrato de errores publicos que conserva
el campo error utilizado por el frontend y agrega code.

En los ocho endpoints HTTP evaluados, los errores internos
devuelven:
- Estado HTTP 500.
- Mensaje publico controlado.
- Codigo INTERNAL_ERROR.

Se incorporo diagnostico por identificador de operacion
y tipo de error, sin registrar el mensaje original.

Resultados:
- Pruebas HTTP: 8/8 PASS.
- Clasificacion de errores: 9/9 PASS.
- go test ./...: PASS.
- go vet ./...: PASS.
- git diff --check: PASS.

El contrato aun no cubre todos los endpoints de Poli-REDI.
La autenticacion Entra ID y los demas codigos publicos
requieren pruebas adicionales.

## Pendientes

- [x] Probar respuestas HTTP 500 ante errores simulados en ocho endpoints.
- Verificar ausencia de detalles SQL/JWT en los endpoints.
- Revisar errores de negocio aun no clasificados.
- Definir codigos publicos estables.
- Registrar errores internos de forma segura.
- Ejecutar regresion final antes de cerrar SEC-005.

## Restricciones

No modificar la base de datos habitual ni
el despliegue MVP3 validado.

El requisito SEC-005 permanece abierto.
