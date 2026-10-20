# MVP3 — Cierre del módulo 4D

**Proyecto:** Poli-REDI
**Fecha:** 20 de octubre de 2026
**Rama:** feature/mvp3
**Módulo:** Bloqueos administrativos de disponibilidad

## Implementación

Se implementó la gestión administrativa de bloqueos
de disponibilidad de recursos.

Funcionalidades:
- Consultar bloqueos administrativos.
- Crear bloqueos de disponibilidad.
- Desactivar bloqueos conservando el historial.
- Validar recursos, fechas y tipos de bloqueo.
- Prevenir solapamientos y conflictos.
- Integrar bloqueos con el calendario.
- Restringir operaciones a administradores.

## Validaciones realizadas

| Prueba | Resultado |
|---|---|
| Autorización HTTP | PASS |
| Restricciones PostgreSQL | PASS |
| Integración del repositorio | PASS |
| Control de concurrencia | PASS |
| Integración con calendario | PASS |
| Pruebas unitarias frontend | PASS |
| Pruebas de componentes Vue | PASS |
| Build de producción | PASS |
| Creación desde interfaz | PASS |
| Desactivación desde interfaz | PASS |

## Despliegue

Backend validado:
localhost/poliredi-backend:mvp3-blocks-4d

Base de datos:
PostgreSQL 16

Frontend:
Vue 3 + Vite, servido mediante Caddy.

## Seguridad e integridad

- Autenticación de desarrollo deshabilitada en despliegue.
- Autorización administrativa en endpoints protegidos.
- Restricciones y triggers PostgreSQL verificados.
- Pruebas de integración ejecutadas en base aislada.
- Respaldo previo generado.
- Imagen anterior conservada para rollback.

## Observaciones

Podman Compose 1.3.0 recreó PostgreSQL durante
un despliegue del backend, pese a solicitar --no-deps.

Se requiere corregir el procedimiento de despliegue
antes de futuras actualizaciones.

La integración del calendario fue validada mediante
pruebas automatizadas. La creación y desactivación
se validaron desde la interfaz publicada.

## Estado

Módulo 4D: IMPLEMENTADO Y VALIDADO.

Pendiente: validación de regresión de extremo a extremo
del MVP3 completo.
