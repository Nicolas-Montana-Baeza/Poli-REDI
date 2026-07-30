# Acta de revision integral - 2026-07-30

## Alcance

Revision tecnica y documental del incremento local de Poli-REDI: autorizacion
por roles, contratos de error, politica de reservas, experiencia unificada de
detalle, accesibilidad, dashboard, historial y migraciones `007`/`008`.

Esta acta no certifica el despliegue online ni una ejecucion de migraciones sobre
Azure SQL. Conserva la distincion entre implementado localmente y validado en el
ambiente integrado.

## Resultado

**Estado de entrega:** APROBABLE.

**Hallazgos P0/P1:** no quedan hallazgos abiertos de estas severidades en el
alcance revisado. Los defectos detectados en permisos, exposicion del codigo,
politica ausente, errores, responsive, teclado y foco fueron corregidos y
cubiertos localmente.

**Siguiente rol recomendado:** QA/Despliegue. Arquitectura debe acompañar la
aprobacion y ejecucion de migraciones en la unica base.

## Decisiones confirmadas

1. El backend determina identidad, rol y estado; la interfaz no otorga permisos.
2. Sin politica publicada valida, la creacion de reservas falla cerrada.
3. El detalle de reserva se reutiliza en Disponibilidad, Mis Reservas, Historial
   y confirmacion por codigo.
4. La reutilizacion visual usa capacidades explicitas y no amplia permisos.
5. El codigo se consulta bajo demanda, owner-only y solo para reservas grupales
   en estado admitido.
6. Toda la tarjeta de Mis Reservas es seleccionable con puntero, Enter y Espacio.
7. Los dialogos controlan foco, Escape, fondo y restauracion del foco.
8. Talleres e inscripciones propias son ampliacion de MVP 2. Clases,
   campeonatos y otros eventos institucionales son MVP 3 y no prueban asistencia
   personal sin una relacion usuario-actividad.
9. `007` y `008` son prospectivas y no retroactivas.

## Migraciones y recuperacion

### 007

Repara exclusivamente el bootstrap reconocible. Debe fallar cerrada ante una
politica administrada o una estructura divergente. Requiere backup, preflight,
postcheck, segunda ejecucion idempotente y confirmacion de que no se alteraron
reservas historicas.

### 008

Extiende la defensa de agenda personal a participaciones `CONFIRMED`. Un solape
real se rechaza; un extremo contiguo se permite. Requiere backup, postcheck,
segunda ejecucion idempotente y pruebas integradas en ambas direcciones
(reserva contra participacion y confirmacion contra reserva/participacion).

Ante un fallo: detener el despliegue, abrir una sesion nueva, comprobar
`@@TRANCOUNT = 0` y `XACT_STATE() = 0`, conservar evidencia y restaurar el backup
si no puede demostrarse un estado compatible. No usar scripts destructivos.

## Evidencia local

| Verificacion | Resultado |
| :--- | :--- |
| Pruebas Node | 18 aprobadas |
| Pruebas Vitest | 77 aprobadas |
| Build frontend de produccion | Aprobado |
| Revision de roles y contratos | Sin P0/P1 abiertos |
| Azure SQL 007/008 | Pendiente |
| Flujo online Entra/CORS/API | Pendiente |

## Pendientes reales

### P2

* Ejecutar `007` y `008` en copia controlada y Azure SQL con backup, pre/postcheck,
  idempotencia y simulacion de recuperacion.
* Ejecutar validacion integrada online con Microsoft Entra ID, CORS, API y base
  desplegada.
* Completar notificaciones (lectura, destinos y eventos) y administracion
  institucional que excede el flujo de usuario.

### P3

* Resolver la advertencia conocida de tamaño del bundle frontend.
* Ampliar matriz manual de navegadores y dispositivos con evidencia visual.
* Incorporar historial institucional de clases, campeonatos y otros eventos en
  MVP 3, sin inferir asistencia.

## Recomendacion de MVP

MVP 1 permanece demostrable. MVP 2 es **APROBABLE para cierre tecnico
condicionado**, no cerrado: el flujo local y su UX tienen evidencia automatizada,
pero faltan migraciones reales e integracion online. No adelantar funcionalidad
de MVP 3 para declarar cerrado MVP 2.
