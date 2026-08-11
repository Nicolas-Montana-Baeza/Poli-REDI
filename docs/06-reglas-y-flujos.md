# Reglas y flujos de Poli-REDI

> **Estado:** CANDIDATO CANÓNICO DEL NUEVO ÁRBOL — contrastado con el código local<br>
> **Corte:** 2026-08-11<br>
> **Alcance:** reservas individuales, `OPEN_USE`, solicitudes grupales y talleres<br>
> **No demuestra:** validación integrada online, concurrencia real en Azure SQL ni migraciones aplicadas

## Resumen

- Una reserva no grupal válida nace `CONFIRMED`; una grupal nace `PENDING` o
  `CONFIRMED` según el owner inicial y el mínimo de su política.
- `OPEN_USE` no consume frecuencia ni exclusividad del recurso, pero respeta
  la agenda personal y bloqueos administrativos.
- El grupo cambia reversiblemente entre `PENDING` y `CONFIRMED` antes del
  deadline; el owner no puede retirar su propia participación.
- Talleres conservan episodios: alta `CONFIRMED`, baja `CANCELLED` y una
  reinscripción futura crea una fila nueva.
- El conflicto de alta de talleres vigente es taller ↔ taller, no una agenda
  unificada taller ↔ reserva.

## 1. Transiciones operativas de una reserva

**Objetivo:** representar las transiciones implementadas sin afirmar que un
retiro ordinario cancela automáticamente la solicitud.

```mermaid
stateDiagram-v2
    [*] --> CONFIRMED : Crear reserva no grupal
    [*] --> PENDING : Crear solicitud grupal; conteo inicial bajo el mínimo
    PENDING --> PENDING : Confirmación o retiro sin alcanzar el mínimo
    PENDING --> CONFIRMED : Confirmación deja conteo >= mínimo
    CONFIRMED --> CONFIRMED : Cambio mantiene conteo >= mínimo
    CONFIRMED --> PENDING : Retiro no-owner deja conteo < mínimo
    PENDING --> CANCELLED : Deadline vencido y conteo < mínimo
    PENDING --> CANCELLED : Cancela propietario o administrador
    CONFIRMED --> CANCELLED : Cancela propietario o administrador
    CANCELLED --> [*]
    note right of PENDING
      El propietario ya cuenta como
      participante CONFIRMED.
    end note
    note right of CONFIRMED
      Finalizar el horario no cambia
      el estado persistido; Historial
      lo deriva de fecha y hora.
    end note
```

- El deadline es inclusivo: confirmar, retirar o editar el objetivo se admite
  mientras `now` no sea posterior al límite.
- `REJECTED` y `EXPIRED` existen en modelo/constraint, pero no poseen una ruta
  backend que produzca esas transiciones en este corte.
- La cancelación manual implementada está disponible para owner/admin mientras
  la reserva no haya finalizado; un cutoff anticipado configurable es futuro.

## 2. Reserva individual y recurso `OPEN_USE`

**Objetivo:** separar conflicto de agenda personal de exclusividad del recurso.

```mermaid
flowchart TD
    START(["Usuario selecciona recurso, fecha y hora"])
    LOAD["Frontend carga /api/resources,<br/>/api/reservation-policy/current y disponibilidad"]
    CLIENT["Frontend valida política visible:<br/>ventana, jornada, slot y duración"]
    POST["POST /api/reservations"]
    AUTH{"¿Sesión válida y<br/>usuario habilitado?"}
    RUT{"¿Usuario normal<br/>sin RUT válido?"}
    MODE{"reservationMode"}
    DENY["Rechazar recurso inactivo,<br/>INFORMATIVE o ADMIN_ONLY sin rol"]
    OPEN["OPEN_USE:<br/>activityId = null<br/>no consume frecuencia"]
    EXCLUSIVE["RESERVABLE / ADMIN_ONLY:<br/>validar frecuencia y exclusividad"]
    PERSONAL_OPEN["Validar solape personal contra<br/>reservas propias y participaciones CONFIRMED"]
    PERSONAL_EXCLUSIVE["Validar solape personal contra<br/>reservas propias y participaciones CONFIRMED"]
    ADMINBLOCK["Validar bloqueo administrativo<br/>del mismo recurso"]
    SHARED["No bloquear por reservas concurrentes,<br/>scheduled activities ni talleres"]
    RESOURCEBLOCK["Validar reservas, actividades y talleres<br/>solapados sobre el mismo recurso"]
    POLICY["Repositorio abre transacción SERIALIZABLE,<br/>carga política publicada y colecciones"]
    GROUP{"¿Recurso incluido en<br/>groupResourceIds?"}
    INSERT["Insertar reserva CONFIRMED"]
    GROUPFLOW["Continuar en flujo grupal"]
    OK(["201 Reserva creada"])
    ERROR(["400 / 403 / 409 / 500<br/>error de dominio seguro"])

    START --> LOAD --> CLIENT --> POST --> AUTH
    AUTH -- "No" --> ERROR
    AUTH -- "Sí" --> RUT
    RUT -- "Sí" --> ERROR
    RUT -- "No o administrador" --> MODE
    MODE -- "INACTIVE / INFORMATIVE / ADMIN_ONLY no autorizado" --> DENY --> ERROR
    MODE -- "OPEN_USE" --> OPEN --> PERSONAL_OPEN
    PERSONAL_OPEN --> ADMINBLOCK --> SHARED --> POLICY
    MODE -- "RESERVABLE / ADMIN_ONLY autorizado" --> EXCLUSIVE
    EXCLUSIVE --> PERSONAL_EXCLUSIVE
    PERSONAL_EXCLUSIVE --> RESOURCEBLOCK --> POLICY
    POLICY --> GROUP
    GROUP -- "Sí" --> GROUPFLOW
    GROUP -- "No" --> INSERT --> OK
    NOTE["Regla semiabierta:<br/>inicioA < finB y inicioB < finA.<br/>Los extremos contiguos son válidos."]
    PERSONAL_OPEN -.-> NOTE
    PERSONAL_EXCLUSIVE -.-> NOTE
```

- `OPEN_USE` no puede ser recurso grupal, no consume frecuencia y fuerza
  `activityId = null`.
- Sí lo afectan los bloqueos administrativos y los solapes de agenda del actor.
  No lo bloquean las reservas concurrentes, actividades ni talleres del recurso.
- El solape personal incluye reservas propias y participaciones grupales
  `CONFIRMED`; no incluye de forma general talleres en otro recurso.

## 3. Solicitud grupal completa

**Objetivo:** cubrir creación, código, participación, quórum, expiración y
cancelación sin exponer PII ni el secreto en listados.

```mermaid
flowchart TD
    CREATE["POST /api/reservations<br/>recurso incluido en groupResourceIds"]
    VALIDATE["Validar objetivo entre mínimo y<br/>capacidad snapshot"]
    TX["Transacción SERIALIZABLE"]
    OWNER["Crear reserva y participante owner CONFIRMED<br/>conteo inicial = 1"]
    INITIAL{"¿El conteo inicial 1<br/>alcanza el mínimo?"}
    INITIALCONFIRMED["Reserva CONFIRMED"]
    PENDING["Reserva PENDING"]
    SECRET["Guardar hash en reservations y secreto<br/>AES versionado en join_code_secrets"]
    DETAIL["Propietario abre detalle"]
    GETCODE["GET /api/reservations/:id/join-code"]
    SHARE["Copiar o compartir /join/:code"]
    LOOKUP["GET /api/group-reservations/:code"]
    FOUND{"¿Código activo y válido?"}
    PROGRESS["Mostrar progreso agregado,<br/>deadline y membresía"]
    ACTION{"Acción"}
    CONFIRM["PUT .../confirmation"]
    WITHDRAW["DELETE .../confirmation"]
    CONFIRMACCOUNT{"¿Cuenta activa y RUT?"}
    CONFIRMDEADLINE{"¿Deadline abierto?"}
    CONFIRMVALID{"¿Hay cupo y no existe<br/>solape personal?"}
    WITHDRAWACCOUNT{"¿Cuenta activa y RUT?"}
    WITHDRAWDEADLINE{"¿Deadline abierto?"}
    OWNERWITHDRAW{"¿Es propietario?"}
    COUNT["Actualizar participante y recalcular conteo"]
    STATE{"conteo >= mínimo"}
    CONFIRMED["Reserva CONFIRMED"]
    STILLPENDING["Reserva PENDING"]
    DEADLINE["Worker o lectura perezosa detecta<br/>deadline vencido bajo mínimo"]
    CANCELLED["Reserva CANCELLED<br/>CONFIRMATION_DEADLINE"]
    CANCEL["PATCH /api/reservations/cancel<br/>propietario o administrador"]
    ERROR404["404 código inexistente/no disponible"]
    ERROR403["403 cuenta o RUT no elegible"]
    ERROR409["409 cupo, solape o retiro owner"]
    ERROR410["410 deadline vencido"]
    CREATE --> VALIDATE --> TX --> OWNER --> INITIAL
    INITIAL -- "Sí" --> INITIALCONFIRMED --> SECRET
    INITIAL -- "No" --> PENDING --> SECRET
    SECRET --> DETAIL
    DETAIL --> GETCODE --> SHARE --> LOOKUP --> FOUND
    FOUND -- "No" --> ERROR404
    FOUND -- "Sí" --> PROGRESS --> ACTION
    ACTION -- "Confirmar" --> CONFIRM --> CONFIRMACCOUNT
    CONFIRMACCOUNT -- "No" --> ERROR403
    CONFIRMACCOUNT -- "Sí" --> CONFIRMDEADLINE
    CONFIRMDEADLINE -- "No" --> ERROR410
    CONFIRMDEADLINE -- "Sí" --> CONFIRMVALID
    CONFIRMVALID -- "No" --> ERROR409
    CONFIRMVALID -- "Sí" --> COUNT
    ACTION -- "Retirar" --> WITHDRAW --> WITHDRAWACCOUNT
    WITHDRAWACCOUNT -- "No" --> ERROR403
    WITHDRAWACCOUNT -- "Sí" --> WITHDRAWDEADLINE
    WITHDRAWDEADLINE -- "No" --> ERROR410
    WITHDRAWDEADLINE -- "Sí" --> OWNERWITHDRAW
    OWNERWITHDRAW -- "Sí" --> ERROR409
    OWNERWITHDRAW -- "No" --> COUNT
    COUNT --> STATE
    STATE -- "Sí" --> CONFIRMED
    STATE -- "No" --> STILLPENDING
    STILLPENDING --> DEADLINE --> CANCELLED
    PENDING --> CANCEL
    INITIALCONFIRMED --> CANCEL
    CONFIRMED --> CANCEL
    CANCEL --> CANCELLED
```

- Owner cuenta desde la creación y no puede retirarse; debe cancelar la
  solicitud completa si corresponde.
- Confirmar y retirar exigen cuenta no bloqueada, RUT y deadline abierto.
  Cupo y solape personal se evalúan solo al confirmar.
- El código no viaja en listados ni en el POST. El owner lo consulta o rota;
  código inválido, rotado, terminal o ajeno recibe una respuesta uniforme.
- El worker periódico y las lecturas perezosas cancelan únicamente solicitudes
  `PENDING` vencidas que siguen bajo el mínimo.

## 4. Talleres

**Objetivo:** mostrar alta, baja, reinscripción e historial como episodios
distintos, sin inferir asistencia.

```mermaid
flowchart TD
    LIST["GET /api/workshops<br/>incluye isEnrolled y enrolledCount"]
    CHOICE{"Acción del usuario"}
    ENROLL["POST /api/workshops/:id/enroll"]
    RUT{"¿Usuario normal con RUT<br/>o administrador?"}
    ETX["Transacción SERIALIZABLE:<br/>lock usuario y luego taller"]
    ACTIVE{"¿Taller activo?"}
    CURRENT{"¿Ya existe inscripción CONFIRMED?"}
    VALID["Validar cupo, ocurrencias normalizadas<br/>y solape taller ↔ taller"]
    CREATED["Crear nuevo episodio CONFIRMED<br/>y audit WORKSHOP_ENROLLMENT_CREATED"]
    ENROLLOK["201 nueva inscripción"]
    IDEMPOTENT["200 inscripción ya vigente"]
    WITHDRAW["DELETE /api/workshops/:id/enrollment"]
    WTX["Transacción SERIALIZABLE:<br/>lock usuario y luego taller"]
    WACTIVE{"¿Taller activo?"}
    HASROW{"¿Existe episodio CONFIRMED propio?"}
    CANCEL["Cambiar solo ese episodio a CANCELLED,<br/>liberar cupo y registrar auditoría"]
    WOK["200 changed=true"]
    WIDEM["200 changed=false"]
    REENROLL["Nueva acción futura: POST de inscripción<br/>crea otro episodio CONFIRMED"]
    HISTORY["GET /api/workshop-enrollments/mine"]
    VIEW["Historial combina reservas e inscripciones;<br/>conserva CONFIRMED y CANCELLED"]
    ERR403["403 RUT requerido para alta"]
    ERR404["404 taller no encontrado"]
    ERR409["409 sin cupo, horario inválido,<br/>solape o taller cerrado"]

    LIST --> CHOICE
    CHOICE -- "Inscribirse" --> ENROLL --> RUT
    RUT -- "No" --> ERR403
    RUT -- "Sí" --> ETX --> ACTIVE
    ACTIVE -- "No" --> ERR404
    ACTIVE -- "Sí" --> CURRENT
    CURRENT -- "Sí" --> IDEMPOTENT
    CURRENT -- "No" --> VALID
    VALID -- "Error" --> ERR409
    VALID -- "Válido" --> CREATED --> ENROLLOK
    CHOICE -- "Reinscribirse tras una baja" --> REENROLL --> RUT
    CHOICE -- "Desinscribirse" --> WITHDRAW --> WTX --> WACTIVE
    WACTIVE -- "No" --> ERR409
    WACTIVE -- "Sí" --> HASROW
    HASROW -- "No" --> WIDEM
    HASROW -- "Sí" --> CANCEL --> WOK
    IDEMPOTENT --> HISTORY
    ENROLLOK --> HISTORY
    WIDEM --> HISTORY
    WOK --> HISTORY
    HISTORY --> VIEW
    NOTE1["La baja no exige RUT ni tiene corte horario.<br/>No permite indicar un usuario tercero."]
    WITHDRAW -.-> NOTE1
    NOTE2["Implementación actual: el conflicto de alta<br/>es taller ↔ taller; no cubre toda reserva personal."]
    VALID -.-> NOTE2
```

- Un taller inactivo devuelve `404` al inscribir y
  `409 WORKSHOP_ENROLLMENT_CLOSED` al intentar la baja.
- La baja propia es idempotente, libera cupo, no exige RUT y no posee cutoff
  mientras el taller siga activo.
- Reinscribir crea un episodio `CONFIRMED` nuevo; no reactiva la fila
  `CANCELLED`.
- El historial ordena por `enrolledAt`; no atribuye asistencia ni usa cada
  ocurrencia semanal como fecha de inscripción.

## 5. Privacidad, disponibilidad e historial

- Una reserva ajena expone ocupación y tipo seguro, no owner, actividad,
  participantes, objetivo, capacidad ni deadline.
- Una actividad institucional ajena conserva la categoría segura y usa un
  título genérico.
- Historial reúne reservas propias/participadas e inscripciones de talleres.
  Una actividad programada no se vuelve personal sin una relación explícita
  usuario–actividad.
- El feed unificado de disponibilidad por rango existe en backend, pero la
  integración completa del frontend continúa pendiente/no validada.

## 6. Flujos pendientes de MVP 3/MVP 4

No deben documentarse como cerrados todavía:

1. prioridad institucional y resolución de conflictos;
2. publicación prospectiva frente a corrección excepcional;
3. bloqueo/desbloqueo y CRUD de disponibilidad;
4. composición completa de disponibilidad por audiencia;
5. notificaciones, lectura, reportes e infracciones;
6. relación usuario–actividad institucional;
7. corte y conciliación con Google Calendar.

## 7. Secuencias y documentos relacionados

- [Secuencia completa MVP 1](anexos/diagramas/secuencia-mvp1.md)
- [Secuencias MVP 2](anexos/diagramas/secuencias-mvp2.md)
- [Arquitectura y contratos](02-arquitectura-y-contratos.md)
- [Base de datos y migraciones](04-base-de-datos-y-migraciones.md)
- [Modelo entidad-relación](anexos/diagramas/modelo-entidad-relacion.md)

## Fuentes

- [`backend/internal/services/reservations_service.go`](../backend/internal/services/reservations_service.go)
- [`backend/internal/repositories/reservations_repository.go`](../backend/internal/repositories/reservations_repository.go)
- [`backend/internal/repositories/participants_repository.go`](../backend/internal/repositories/participants_repository.go)
- [`backend/internal/services/workshops_service.go`](../backend/internal/services/workshops_service.go)
- [`backend/internal/repositories/workshops_repository.go`](../backend/internal/repositories/workshops_repository.go)
- [`frontend/src/services/reservations.service.js`](../frontend/src/services/reservations.service.js)
- [`frontend/src/services/workshops.service.js`](../frontend/src/services/workshops.service.js)
- [`database/schema.sql`](../database/schema.sql)
