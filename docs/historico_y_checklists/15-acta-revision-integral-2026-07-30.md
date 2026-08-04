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

---

## Actualizacion: estados asincronos y skeletons

Esta seccion agrega evidencia al acta sin sustituir las conclusiones historicas
anteriores.

### Causa y contrato corregido

La regresion de Disponibilidad se origino porque el calculo de carga inicial
omitia `policyLoading` y la carga de actividades. El contrato corregido espera
autenticacion, politica, recursos, reservas, talleres y actividades antes de
retirar el skeleton inicial.

El skeleton se limita a `initialLoading` sin datos. Un refresh conserva datos y
presenta un indicador discreto; una mutacion conserva el contexto y presenta un
spinner local. Los stores mantienen `status`, `hasLoaded`, `refreshing`,
deduplicacion de consultas, `requestId` y proteccion contra respuestas obsoletas.
Historial conserva resultados parciales y advierte cuando falla una sola fuente.

### Sistema visual y accesibilidad

`SkeletonLoader` incorpora los arquetipos `availability-timelines`,
`media-grid`, `card-grid`, `list`, `detail`, `metrics-table` y `compact-rows`.
`AsyncRegion` comunica `aria-busy`, estado y anuncios para lectores de pantalla.
El movimiento reducido desactiva shimmer y transiciones; los medios reservan
una proporcion 16:9.

Las superficies cubiertas son Disponibilidad, Dashboard, Mis Reservas, detalle
por ruta, Historial, Notificaciones y vistas administrativas de usuarios,
recursos, reservas, talleres, reportes y configuracion. No se aplica skeleton a
Join, mutaciones ni modales que ya disponen del objeto de detalle.

### Evidencia y pendientes de esta actualizacion

| Verificacion | Resultado |
| :--- | :--- |
| Pruebas Node | 18 aprobadas |
| Pruebas Vitest | 98 aprobadas |
| Build frontend de produccion | Aprobado |
| `diff-check` | Aprobado |
| Bundle frontend | Advertencia conocida: 527.73 kB |

Permanece pendiente QA visual con anchos de 377, 500, 768 y 1440 px, incluyendo
teclado, lectores de pantalla y movimiento reducido. Tambien permanece pendiente
la optimizacion o division del bundle.

---

## Actualizacion: taxonomia visual y privacidad de disponibilidad

Esta actualizacion amplia el acta sin reemplazar las decisiones ni la evidencia
historica anteriores.

### Contrato incorporado

Se separo el tipo u origen del bloque de su estado. La taxonomia visible incluye
Reserva, Reserva grupal, Uso libre, Taller, Clase, Entrenamiento, Campeonato,
Evento e Institucional. El helper unico y
`AvailabilityTypeChip`/`AvailabilityTypeLegend` mantienen la misma
clasificacion en Por recurso y Agenda del dia; la leyenda se denomina `Tipos de
bloque`. `OPEN_USE` conserva el heatmap y explica que su intensidad corresponde
a reservas simultaneas.

El backend publica `availabilityKind` como `RESERVATION`,
`GROUP_RESERVATION` o `SCHEDULED_ACTIVITY`. Las actividades programadas
conservan la categoria segura mediante `activityType` y usan Institucional ante
un valor generico o desconocido.

### Privacidad por audiencia

Una reserva ajena se presenta como `Reserva` y no expone identidad, actividad,
metricas de participantes, capacidad, plazo ni permisos. Solo conserva el tipo
grupal seguro cuando corresponde. La reserva propia conserva el detalle y las
acciones necesarias; el administrador conserva el detalle operacional. Una
actividad institucional no expone la identidad de quien la creo.

La accesibilidad usa texto ademas de color, no agrega foco a los chips y compone
un `aria-label` con tipo, titulo seguro, estado, horario y accion. La definicion
de futuros bloqueos entre categorias institucionales no forma parte de este
incremento.

### Evidencia actualizada

| Verificacion | Resultado |
| :--- | :--- |
| Backend | `go test ./...` aprobado |
| Pruebas Node | 18 aprobadas |
| Pruebas Vitest | 119 aprobadas |
| Build frontend de produccion | Aprobado |
| `diff-check` | Aprobado |
| Bundle frontend | Advertencia conocida: 531.79 kB |

El alcance queda **APROBABLE condicionado** a QA visual y de privacidad por
audiencia, sin declarar implementados bloqueos futuros ni ampliar el cierre de
MVP 2 con funciones institucionales de MVP 3.

---

## Actualizacion 2026-08-04: desinscripcion de talleres

Esta actualizacion incorpora al alcance controlado de MVP 2 la cancelacion de
la inscripcion propia a talleres. No sustituye la evidencia historica anterior.

### Decisiones registradas

1. `DELETE /api/workshops/:id/enrollment` usa al usuario autenticado y cancela
   solo su episodio `CONFIRMED`; no recibe ni permite seleccionar a un tercero.
2. La desinscripcion no exige RUT. El requisito de RUT permanece en el alta.
3. La operacion es idempotente y solo procede sobre un taller activo. Un taller
   inactivo responde `409` con `WORKSHOP_ENROLLMENT_CLOSED`.
4. La transicion a `CANCELLED` libera cupo, deja de bloquear solapes y conserva
   el episodio en Historial como `Inscripcion cancelada`.
5. Una reinscripcion crea un episodio nuevo `CONFIRMED`; no reactiva ni elimina
   el episodio cancelado.
6. La auditoria distingue `WORKSHOP_ENROLLMENT_CANCELLED` y
   `WORKSHOP_ENROLLMENT_CREATED`.
7. No existe corte horario hasta definir formalmente el periodo del taller.

### Evidencia del incremento

| Verificacion | Resultado |
| :--- | :--- |
| Backend | `go test ./... -count=1` aprobado en todos los paquetes |
| Pruebas Node | 18 aprobadas |
| Pruebas Vitest | 144 aprobadas |
| Build frontend de produccion | Aprobado |
| `diff-check` | Aprobado |

El incremento queda **IMPLEMENTADO Y VERIFICADO LOCALMENTE**. La validacion
integrada/online y las condiciones generales que mantienen al MVP 2 como cierre
condicionado no cambian.
