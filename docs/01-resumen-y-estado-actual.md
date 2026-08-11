# Poli-REDI - Resumen Ejecutivo y Estado Actual del Producto

> **Revisión vigente:** 2026-07-30. El resultado local es `APROBABLE` para
> continuar el cierre condicionado de MVP 2. La evidencia, riesgos y límites
> constan en el
> [acta integral](./historico_y_checklists/15-acta-revision-integral-2026-07-30.md).

> **Fecha de consolidación:** 2026-07-23  
> **Propósito:** Documento único para comprender la visión ejecutiva, el estado funcional de cada módulo, las reglas vigentes y la matriz de brechas del prototipo.

---

## 1. Problema y Objetivo

El proceso de reserva del Polideportivo institucional (3 canchas y 1 sala multiusos) dependía históricamente de una gestión manual en Google Calendar, concentrando decisiones en el encargado sin visibilidad pública de disponibilidad ni trazabilidad.

**Objetivo:** Desarrollar un prototipo funcional (aplicación web desacoplada) que centralice la consulta de disponibilidad, reservas particulares e institucionales, aplicando validaciones de horario, frecuencia semanal y quorum mínimo en servidor.

---

## 2. Convenciones de Estado

* `APROBADO`: Definido explícitamente en el alcance o acordado con contraparte.
* `IMPLEMENTADO`: Comportamiento observable en código Go/Vue o base de datos.
* `VERIFICADO`: Comprobado mediante pruebas unitarias/builds con resultado satisfactorio.
* `ACCEPTED LOCALLY`: Incremento aceptado mediante pruebas locales, todavía pendiente de validación integrada en Azure SQL.
* `IMPLEMENTADO LOCAL`: Funcionalidad integrada en el código y comprobada en el ambiente local mediante pruebas o build; todavía pendiente de validación en el ambiente desplegado.
* `PENDIENTE`: Funcionalidad planeada o en trabajo futuro.

---

## 3. Estado Funcional por Módulos

### Delta verificado el 2026-07-30

* La revisión de roles y contratos no mantiene hallazgos P0/P1 abiertos dentro
  del alcance revisado. Esto no equivale a una prueba de penetración.
* La política de reservas falla cerrada si no existe una política publicada
  válida; la interfaz filtra recursos según estado, modalidad y rol.
* Disponibilidad, Mis Reservas, Historial y confirmación por código reutilizan el
  mismo detalle con capacidades explicitas. El codigo se obtiene bajo demanda y
  solo para el propietario autorizado.
* Las tarjetas personales son seleccionables completas. Los dialogos administran
  foco, Escape y restauración; la línea de tiempo y el progreso tienen semántica
  de teclado/ARIA.
* El dashboard evita duplicar la próxima reserva.
* Talleres e inscripciones propias son MVP 2; clases, campeonatos y otros eventos
  institucionales son MVP 3 y requieren relación explícita antes de atribuir
  participación personal.
* `007` y `008` son prospectivas y no reescriben reservas históricas.

**Validación local exacta:** `go test ./...`, 18 pruebas Node, 119 pruebas
Vitest, build frontend de producción y `diff-check` aprobados. El build conserva
la advertencia conocida
por un bundle de 531.79 kB. Siguen pendientes Azure SQL `007`/`008`, el flujo
online completo con Entra ID/CORS/API y QA visual en 377, 500, 768 y 1440 px.

### Tipos de bloque y privacidad de disponibilidad

* La taxonomía visual distingue `Reserva`, `Reserva grupal`, `Uso libre`,
  `Taller`, `Clase`, `Entrenamiento`, `Campeonato`, `Evento` e
  `Institucional`. El chip comunica el tipo u origen del bloque; `Pendiente`,
  `Confirmada`, `Cancelada` o `Programada` se mantienen como estado separado.
* Un helper único resuelve el tipo y los componentes
  `AvailabilityTypeChip`/`AvailabilityTypeLegend` lo presentan de forma
  consistente en las vistas Por recurso y Agenda del día. La leyenda se titula
  `Tipos de bloque`.
* `OPEN_USE` conserva su mapa de intensidad y agrega un chip en la cabecera del
  recurso con la explicación de que la intensidad representa reservas
  simultáneas; no convierte cada asistencia en una tarjeta individual.
* El backend entrega `availabilityKind` con los valores
  `RESERVATION`, `GROUP_RESERVATION` o `SCHEDULED_ACTIVITY`. Para un usuario
  normal, una reserva ajena conserva solamente el tipo grupal seguro cuando
  corresponde y se presenta como `Reserva`: no expone PII, actividad, métricas
  de participantes ni plazo. La reserva propia conserva el detalle funcional;
  el administrador conserva el detalle operacional según su audiencia.
* Una actividad programada se presenta con su `activityType`, usando una
  categoría institucional genérica cuando no existe una correspondencia
  conocida. La definición de nuevos bloqueos entre tipos institucionales queda
  fuera del alcance de este incremento.

### Delta de estados asíncronos y skeletons

* Se corrigió la regresión de Disponibilidad: su carga inicial ahora incluye la
  política y las actividades, además de autenticación, recursos, reservas y
  talleres. La omisión de `policyLoading` y `activities` permitía retirar el
  estado de carga antes de completar el contrato necesario para reservar.
* El skeleton se presenta solo durante `initialLoading` y cuando aún no existen
  datos. Un refresh conserva los datos visibles y usa un indicador discreto; una
  mutación mantiene el contenido y muestra un spinner local en su acción.
* Las superficies cubiertas son Disponibilidad, Dashboard, Mis Reservas,
  detalle de reserva por ruta, Historial, Notificaciones y las vistas
  administrativas de usuarios, recursos, reservas, talleres, reportes y
  configuración.
* Historial conserva y muestra los datos disponibles cuando falla solo una de
  sus fuentes, acompañado por una advertencia parcial. Solo muestra error
  terminal cuando no existe ningún dato útil.
* No se usa skeleton en Join, en mutaciones ni en un modal que ya dispone del
  objeto seleccionado; esos casos conservan el contenido y comunican la acción
  localmente.

| Módulo | Estado | Descripción y Evidencia |
| :--- | :---: | :--- |
| **Autenticación y Sesión** | `IMPLEMENTADO` | Integración con Microsoft Entra ID (JWT) + Modo local Dev Auth (`X-Dev-Auth-*`). |
| **Identidad y Perfil (RUT)** | `ACCEPTED LOCALLY` | Modal condicionado a `/api/me` listo, usuario no administrador y RUT ausente/inválido. Registro write-once, lectura posterior e idempotencia del mismo valor. |
| **Recursos y Disponibilidad** | `IMPLEMENTADO` | Consulta de disponibilidad por fecha/recurso entre 08:00 y 22:00 (`GET /api/resources`). |
| **Creación de Reservas** | `IMPLEMENTADO` | Reservas particulares e institucionales controladas por servidor. |
| **Flujo Grupal y Código de Unión**| `ACCEPTED LOCALLY` | Progreso, código cifrado recuperable solo por propietario, rotación, `/join`, confirmación, retiro, reconfirmación, deadline inclusivo y expiración. Migración 004 y concurrencia real en Azure SQL pendientes. |
| **Diseño y UI/UX (Tokens & Components)**| `IMPLEMENTADO LOCAL` | `StatusBadge` se usa en tarjetas, detalle y próximas reservas; `ConfirmModal` reemplaza confirmaciones nativas de cancelación y protege la rotación de códigos; `MetricCard` cubre las tres métricas administrativas; `PrimaryButton` se usa en los modales y acciones del progreso grupal. Calendarios y bloques con semántica especial conservan sus controles propios. La validación integrada/online continúa pendiente. |
| **Control de Frecuencia Semanal** | `IMPLEMENTADO` | Restricción de 7 días corridos entre reservas solicitadas por el mismo usuario. |
| **Prioridad Institucional** | `PENDIENTE` | La regla está aprobada, pero el flujo administrativo de resolución y cancelación automática aún no está implementado. |
| **Cancelación de Reservas** | `IMPLEMENTADO` | Propietario o administrador pueden cancelar reservas activas (`PATCH /api/reservations/cancel`). |
| **Talleres e Inscripciones** | `ACCEPTED LOCALLY` | Cupos, ocurrencias normalizadas, prevención de solapes taller↔taller y desinscripción propia idempotente. `POST /api/workshops/:id/enroll` crea un episodio `CONFIRMED`; `DELETE /api/workshops/:id/enrollment` cancela solo la inscripción activa del usuario autenticado. |
| **Historial Personal** | `PARCIAL` | El historial básico de reservas propias o participadas pertenece a MVP 1 y está implementado. La ampliación de MVP 2 conserva inscripciones de taller confirmadas y canceladas; una cancelada aparece como `Inscripción cancelada` y no demuestra asistencia. |
| **Historial Institucional** | `PENDIENTE` | Las clases, actividades institucionales y otros eventos se incorporarán en MVP 3. No se atribuirán como participación personal mientras no exista una relación explícita usuario–actividad. |
| **Notificaciones Internas** | `PARCIAL` | Consulta y contador (`GET /api/notifications`) y notificación única de expiración verificada localmente; lectura, destinos, otros eventos y sistema completo pendientes. |
| **Panel Administrador** | `PARCIAL` | Lectura operacional, indicadores, imágenes de recursos y políticas; gestión completa de usuarios, recursos, bloqueos y programación pendiente. |
| **Auditoría y Trazabilidad** | `PARCIAL` | El esquema registra cambios de reservas, participantes, objetivos y expiraciones; falta consulta administrativa integral. |

---

## 4. Brechas y Decisiones Aprobadas

1. **Duración de Reservas:** Se permiten selecciones de 30, 60, 90, 120, 150 y 180 minutos.
2. **Ventana de Reserva:** El usuario puede reservar desde el día actual hasta el día anterior al mismo día de la semana siguiente.
3. **Mínimo de Participantes:** Mínimo 10 participantes obligatorios para Canchas 1, 2 y 3.
4. **Cancelación Automática por Prioridad:** Un conflicto entre actividad institucional y reserva particular cancela automáticamente la reserva particular y notifica al afectado.

> **Límite de verificación:** El MVP 2 está `ACCEPTED LOCALLY`. Continúan pendientes la ejecución de la migración 004, su idempotencia y las pruebas de concurrencia real en Azure SQL.

Las extensiones de integridad de RUT y horarios de talleres también están `ACCEPTED LOCALLY`. Migraciones 005/006, DDL, idempotencia y carreras reales en Azure SQL siguen pendientes.

La desinscripción de talleres pertenece a la ampliación controlada de MVP 2. No
exige RUT ni permite retirar inscripciones ajenas: cambia únicamente el episodio
`CONFIRMED` propio a `CANCELLED`, libera el cupo y deja de bloquear solapes. El
taller debe continuar activo; de lo contrario la API responde `409` con
`WORKSHOP_ENROLLMENT_CLOSED`. Mientras no exista un período formal del taller,
no se aplica corte horario. Una reinscripción posterior crea un episodio nuevo
`CONFIRMED` y conserva el cancelado en el historial.

Evidencia del 2026-08-04 para este incremento: `go test ./... -count=1`
aprobado en todos los paquetes, 18 pruebas Node, 144 pruebas Vitest, build
frontend de producción y `diff-check` aprobados.

### Alcance aprobado del historial

1. **MVP 1:** historial básico de reservas propias o participadas, incluyendo
   reservas pasadas y canceladas.
2. **MVP 2 (ampliación controlada):** consulta de talleres e inscripciones del
   usuario, sin presentar una inscripción como asistencia comprobada.
3. **MVP 3:** historial institucional de clases, actividades programadas y otros
   eventos, con acceso acorde al rol.
4. Una clase o evento solo podrá aparecer en el historial personal cuando el
   modelo registre una relación explícita de participación del usuario. La mera
   existencia del evento en la agenda no permite inferir asistencia.

### Matriz RUT por rol y acción

| Actor sin RUT | Reserva normal | Reserva grupal propia | Inscripción a taller | Confirmar como participante |
| :--- | :---: | :---: | :---: | :---: |
| Usuario normal | No | No | No | No |
| Administrador | Sí | Sí | Sí | No |

Una vez registrado, el RUT no puede cambiarse ni borrarse. Repetir el mismo valor es idempotente; un valor diferente o duplicado devuelve `409`. El inicio de sesión local conserva el RUT existente.
