# MVP3 — Auditoría ADMIN-005

Fecha: 20 de octubre de 2026
Rama: feature/mvp3
Estado: implementación parcial, en validación.

## 1. Implementación identificada

- Administración de unidades institucionales.
- Asignación de gestores y miembros.
- Creación de actividades institucionales.
- Programación SINGLE y WEEKLY.
- Detección de conflictos de múltiples elementos.
- Resolución KEEP, ALLOW, CANCEL y RESCHEDULE.
- Integración con disponibilidad y calendario.
- Interfaz Vue y servicios de comunicación con API.
- Persistencia de resoluciones administrativas.

## 2. Evidencia PostgreSQL 16

Se ejecutó el verificador MVP3 en una base temporal.

Pruebas ADMIN-005 aprobadas:

1. TestInstitutionalUnitAuthorizationIntegration
2. TestInstitutionalActivitiesIntegration
3. TestInstitutionalAvailabilityIntegration
4. TestInstitutionalCalendarIntegration
5. TestSchedulingConflictMergeIntegration
6. TestSchedulingConflictResolutionIntegration

Resultado: 6/6 PASS.

El verificador completo terminó con 17/17
pruebas de integración aprobadas.

Esta evidencia corresponde al entorno aislado,
no constituye por sí sola aceptación funcional online.

## 3. Decisión funcional EV-010

El responsable del proyecto indicó que no deben
aplicarse cancelaciones automáticas ante conflictos
de programación institucional.

Comportamiento propuesto:

- Detectar y registrar el conflicto automáticamente.
- No cancelar reservas por la sola creación de actividades.
- Permitir resolución mediante administrador autorizado.
- Exigir justificación y confirmación de decisiones.
- Conservar trazabilidad de la resolución.
- Notificar a los usuarios afectados.

No se aprueba ninguna modificación automática
de reservas por prioridad institucional.

La decisión debe formalizarse mediante el proceso
de aprobación institucional correspondiente antes
de sustituir definitivamente RF-023/v1.

## 4. Pendientes de verificación

- Pruebas Vue de las tres vistas institucionales.
- Pruebas HTTP de autorización y permisos.
- Smoke test de extremo a extremo.
- Validar notificación transaccional de CANCEL.
- Verificar destinatario y privacidad de notificaciones.
- Comprobar comportamiento visual del calendario.
- Probar flujos desde interfaz contra base aislada.
- Revisar tratamiento de conflictos pendientes.
- Validar experiencia administrativa y gestores.

## 5. Pendientes documentales

- Actualizar ADMIN-005 en docs/07-backlog.md.
- Revisar RF-016 y RF-023.
- Registrar evolución EV-010.
- Corregir estados obsoletos del roadmap MVP3.
- Incorporar resultados de pruebas y aceptación.
- No declarar cerrado ADMIN-005 hasta validar E2E.

## 6. Restricciones

- No modificar datos persistentes con pruebas.
- No cancelar reservas automáticamente.
- No desplegar cambios antes de validación.
- No fusionar MVP3 con main durante esta auditoría.

## Evidencia complementaria: frontend y notificaciones

Fecha: 2026-10-20.

### Pruebas de componentes Vue

- InstitutionalActivitiesView: 6/6 PASS.
- InstitutionalUnitsView: 7/7 PASS.
- SchedulingConflictsView: 5/5 PASS.
- Total ADMIN-005 frontend: 18/18 PASS.
- Suite general de componentes: 43/43 PASS.
- Las suites institucionales también aprobaron conjuntamente.

### Integración de cancelación y notificaciones

Prueba ampliada:
`TestSchedulingConflictResolutionIntegration`.

Resultados PostgreSQL 16:

- MVP2: PASS. No se genera notificación automática.
- MVP3: PASS. Se genera exactamente una notificación.
- El destinatario es el propietario de la reserva.
- El administrador no recibe la notificación del propietario.
- El tipo es `RESERVATION_CANCELLED_INSTITUTIONAL`.
- El mensaje no contiene datos personales de terceros.
- Una resolución incompatible no genera notificaciones.
- La cancelación administrativa actualiza la reserva a `CANCELLED`.

La prueba utiliza usuarios y reservas sintéticos en PostgreSQL
efímero. No utiliza la base de datos de producción.

### Validación funcional y pendientes

- Cancelación institucional mediante Funnel: verificada.
- Notificación observada desde la cuenta del propietario: verificada.
- La notificación inicialmente requirió F5 para actualizarse.
- Corrección de actualización al abrir la campana: pruebas locales aprobadas.
- Publicación de la corrección y comprobación sin F5: pendiente.
- Ratificación institucional de RF-023/v2 conforme a EV-010: pendiente.

### Conclusión actualizada

ADMIN-005 tiene aprobadas las pruebas de integración disponibles
y la cobertura de componentes de sus tres vistas principales.

La implementación continúa pendiente de cierre funcional integral.
Estas pruebas no constituyen autorización de despliegue.

## Evidencia complementaria: atomicidad transaccional

Fecha: 2026-10-20.

Prueba:
`TestSchedulingConflictResolutionIntegration`.

Entorno: PostgreSQL 16 efímero, `MVP_SCOPE=mvp3`.

Se agregó una restricción temporal para provocar el rechazo de
notificaciones del tipo `RESERVATION_CANCELLED_INSTITUTIONAL`.

Resultados:

- El intento de cancelación recibió el error SQL esperado: PASS.
- La reserva permaneció `CONFIRMED`: PASS.
- El motivo de cancelación permaneció vacío: PASS.
- El elemento del conflicto permaneció `PENDING`: PASS.
- El conflicto permaneció `PENDING`: PASS.
- No se generaron notificaciones parciales: PASS.
- Tras retirar la restricción, la resolución normal aprobó: PASS.
- Se generó una única notificación para el propietario: PASS.

**Resultado: ATOMICIDAD TRANSACCIONAL APROBADA.**

El fallo fue inducido exclusivamente en la base temporal del
verificador. No se modificó el esquema ni los datos de producción.

### Pendientes después de la validación de atomicidad

- Publicar de forma controlada la corrección de NotificationBell.vue.
- Verificar en Funnel que abrir la campana actualiza las notificaciones sin F5.
- Mantener pendiente la ratificación institucional EV-010.

La funcionalidad principal de ADMIN-005 cuenta con evidencia de
integración y validación manual en Funnel.


## Evidencia funcional en Funnel y compilación

Fecha: 2026-10-20.

### Validación funcional publicada

El usuario confirmó una cancelación desde Conflictos de programación
utilizando el Funnel publicado de Poli-REDI.

La notificación institucional apareció en la cuenta del propietario
tras actualizar la página mediante F5.

Este resultado confirma la generación, persistencia y consulta de la
notificación en el despliegue probado. No demuestra todavía
actualización automática de la interfaz.

### Mejora UX de notificaciones

Archivo: `frontend/src/components/layout/NotificationBell.vue`.

La campana fue modificada para consultar notificaciones al abrirse.

Se incorporó `NotificationBell.test.js`, con dos pruebas de regresión.
El usuario confirmó que las pruebas finalizaron correctamente.

La mejora todavía no ha sido publicada en el Funnel.

### Validación de producción

- Backend: `MVP_SCOPE=mvp3`.
- Backend: `DEV_AUTH_ENABLED=false`.
- Frontend: `frontend/.env.production.local` define `VITE_MVP_SCOPE=mvp3`.
- Compilación Vite 8.3.4: PASS.
- Módulos transformados: 1897.
- `git diff --check`: PASS.

Advertencia no bloqueante: `__dirname` en `vite.config.js`
requiere una futura adecuación al cargador nativo de Vite.

### Dictamen técnico

La funcionalidad principal de ADMIN-005 cuenta con evidencia
automatizada y una comprobación funcional en el Funnel.

La publicación de la mejora UX y la ratificación institucional
de EV-010 permanecen pendientes.

No se autoriza despliegue mediante esta auditoría.
