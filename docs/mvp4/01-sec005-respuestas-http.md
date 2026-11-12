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

## Fase D3 - Autenticacion JWT HTTP

Fecha: 2026-11-12
Estado: VALIDADO PARCIALMENTE

Se separo la inicializacion de JWKS de la validacion
JWT para permitir pruebas aisladas con claves RSA
efimeras generadas localmente.

La configuracion de produccion mantiene JWKS real.

Escenarios HTTP aprobados:
- Token ausente: 401.
- Cabecera incorrecta: 401.
- JWT malformado: 401.
- Firma invalida: 401.
- Emisor invalido: 401.
- Audiencia invalida: 401.
- JWT expirado: 401.
- Identidad sin email: 403.
- Repositorio no disponible: 500.

Resultados:
- Pruebas HTTP: 9/9 PASS.
- go test ./...: PASS.
- go vet ./...: PASS.
- git diff --check: PASS.

Las pruebas no utilizan tokens institucionales reales.
No verifican la descarga de JWKS en produccion.

Pendiente: probar usuarios bloqueados, errores durante
la sincronizacion de identidad y autorizaciones
institucionales adicionales.

## Fase D4 - Identidad y usuarios bloqueados

Fecha: 2026-11-12
Estado: VALIDADO

Se incorporo inyeccion de dependencias al constructor
interno del middleware de autenticacion Entra ID.

La ruta de produccion conserva los repositorios originales.

Se retiro el campo detail de la respuesta Entra ID
para usuarios bloqueados.

### Pruebas HTTP aisladas

- Error al consultar usuario: HTTP 500, PASS.
- Usuario bloqueado: HTTP 403, PASS.
- Error al sincronizar identidad: HTTP 500, PASS.
- Autenticacion exitosa: HTTP 200, PASS.
- Total D4: 4/4 PASS.
- Regresion JWT D3: 9/9 PASS.

No se utilizaron tokens institucionales reales.

### Regresion PostgreSQL 16

Se ejecuto:
infra/local/quadlet/verify-mvp3-policies-ephemeral.sh

Resultados:
- Politicas grupales: PASS.
- Historial y reglas de reservas: PASS.
- Sedes y recursos: PASS.
- Bloqueos administrativos: PASS.
- Calendario de bloqueos: PASS.

Resultado final:
CIERRE AUTOMATIZADO MVP3 POSTGRESQL 16: PASS

El verificador incluye limpieza automatica del
contenedor y volumen efimeros.

### Alcance y limitaciones

La validacion del middleware utiliza repositorios simulados.
La regresion PostgreSQL utiliza integraciones independientes.

No se ha validado una autenticacion completa contra
Microsoft Entra ID y PostgreSQL institucional reales.

La identidad estable por tenant y OID corresponde a SEC-006.

## Fase D5 - Errores PostgreSQL y actualizacion RUT

Fecha: 2026-11-12
Estado: VALIDADO

### Correcciones

- Los errores internos al actualizar RUT responden HTTP 500.
- Los errores de validacion del RUT conservan HTTP 400.
- Los SQLSTATE desconocidos conservan el error tipado interno,
  sin utilizar su mensaje como respuesta publica.
- Se conservo el mensaje historico de P1008.
- El clasificador HTTP reconoce ambas variantes de P1008.

### Evidencia

- Actualizacion HTTP de RUT: 3/3 PASS.
- Mapeo PostgreSQL: 3/3 PASS.
- Clasificacion de reservas: 7/7 PASS.
- Regresion historica P1008: PASS.
- go test -count=1 ./...: PASS.
- go vet ./...: PASS.
- git diff --check: PASS.
- Regresion PostgreSQL 16 efimera: PASS.

### Restricciones

No se modifico el esquema PostgreSQL ni la base de datos
habitual. SEC-005 sigue abierto hasta completar la
auditoria transversal de respuestas y errores.

## Fase D6 - Proteccion transversal de errores de dominio

Fecha: 2026-11-12
Estado: VALIDADO

### D6A - Clasificadores de dominio

Se introdujo sec005KnownSentinelMessage para devolver
exclusivamente el mensaje de errores conocidos.

Se protegieron ocho clasificadores HTTP:
- Administracion de accesos y roles.
- Actividades y unidades institucionales.
- Ocurrencias y talleres.
- Conflictos de programacion.
- Participacion grupal.

Se conservaron los estados HTTP existentes.

### D6B - Recursos, sedes y politicas

Se extendio la proteccion a:
- Recursos y creacion de recursos.
- Creacion y actualizacion de sedes.
- Bloqueos administrativos.
- Detalle de reservas.
- Politicas de reservas.

Los errores de validacion locales mantienen sus
respuestas funcionales.

Se evito transformar errores tecnicos envueltos
en mensajes publicos de validacion de politicas.

### Pruebas

- D6A: 8/8 PASS.
- D6B: mensajes de dominio 6/6 PASS.
- D6B: politicas 5/5 PASS.
- go test -count=1 ./...: PASS.
- go vet ./...: PASS.
- git diff --check: PASS.
- Regresion PostgreSQL 16 efimera: PASS.

Resultado integrado:
CIERRE AUTOMATIZADO MVP3 POSTGRESQL 16: PASS

### Limitaciones

Las pruebas D6A verifican respuestas HTTP de
clasificadores con errores simulados.

Las pruebas D6B verifican seleccion de mensajes
y clasificacion de politicas.

La auditoria de seguridad completa permanece abierta.
SEC-005 no se declara cerrado en esta fase.

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
