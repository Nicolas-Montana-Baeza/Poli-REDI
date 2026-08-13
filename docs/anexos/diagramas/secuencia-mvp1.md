# Anexo — Secuencia completa MVP 1

**Audiencia:** Analista, Arquitecto, desarrollo y QA

**Propósito:** representar contexto, disponibilidad, reserva individual, consulta y cancelación de MVP 1

**Estado:** IMPLEMENTADO LOCALMENTE, CON INTEGRACIÓN PENDIENTE

**Corte:** 2026-08-11

**Fuente:** código local y arquitectura y reglas canónicas

**No demuestra:** login Entra, CORS, API y Azure SQL funcionando juntos online

## Resumen

- El frontend obtiene perfil, catálogos, política y disponibilidad antes de
  crear la reserva.
- El backend controla identidad, RUT, política, frecuencia y conflictos.
- El feed por rango existe en backend; la vista principal aún usa el contrato
  legado sin `from/to`.
- La creación individual exitosa queda `CONFIRMED`.
- La cancelación requiere owner o administrador y una reserva no finalizada.

Volver a [Reglas y flujos](../../06-reglas-y-flujos.md).

## Secuencia

**Objetivo:** mostrar las capas reales y los errores principales de MVP 1.

```mermaid
sequenceDiagram
    actor U as Usuario
    participant FE as Vue / Pinia / Axios
    participant API as Fiber Routes + Handlers
    participant SVC as Reservations Service
    participant REP as Repositories
    participant SQL as Azure SQL

    U->>FE: Abre /availability
    par Contexto de usuario
        FE->>API: GET /api/me
        API->>REP: Resolver usuario local
        REP->>SQL: SELECT users
        SQL-->>REP: Usuario, rol, bloqueo y RUT
        REP-->>API: LocalAuthUser
        API-->>FE: Perfil sanitizado
    and Catálogos y política
        FE->>API: GET /api/resources, /api/activities y /api/reservation-policy/current
        API->>REP: Consultar catálogos y política publicada
        REP->>SQL: SELECT resources, activities y reservation_policies
        SQL-->>REP: Datos vigentes
        REP-->>API: Modelos
        API-->>FE: JSON
    and Disponibilidad
        FE->>API: GET /api/availability/reservations
        Note over FE,REP: from/to y feed unificado: backend implementado; frontend pendiente/no validado
        API->>SVC: GetAvailabilityItems o GetAvailabilityItemsForRange
        SVC->>REP: Consultar fuentes de agenda
        REP->>SQL: SELECT agenda
        SQL-->>REP: Bloques
        REP-->>SVC: Elementos
        SVC-->>API: Payload por audiencia
        API-->>FE: Disponibilidad sanitizada
    end

    U->>FE: Selecciona recurso, fecha, hora y duración
    FE->>FE: Validación de política visible
    FE->>API: POST /api/reservations
    API->>API: RequireAuth, JSON estricto, RUT y hora institucional

    alt Datos o sesión inválidos
        API-->>FE: 400 / 401 / 403
        FE-->>U: Error accionable
    else Solicitud válida
        API->>SVC: CreateReservation(usuario de sesión)
        SVC->>REP: GetResourceByID y frecuencia aplicable
        REP->>SQL: SELECT resource y última solicitud activa
        SQL-->>REP: Contexto
        SVC->>SVC: Validar pasado, modo, frecuencia y taller del recurso
        SVC->>REP: AddReservationWithPolicy
        REP->>SQL: BEGIN SERIALIZABLE
        REP->>SQL: Cargar política publicada y colecciones
        REP->>SQL: Validar solape personal
        REP->>SQL: INSERT reservation
        SQL->>SQL: Constraints y triggers de conflictos

        alt Conflicto o regla de negocio
            SQL-->>REP: Error numerado y rollback
            REP-->>SVC: Error
            SVC-->>API: Error de dominio
            API-->>FE: 400 / 409 / 500
            FE-->>U: Motivo seguro
        else Creación exitosa
            SQL-->>REP: Reserva CONFIRMED
            REP->>SQL: COMMIT
            REP-->>SVC: Reserva creada
            SVC-->>API: Resultado
            API-->>FE: 201
            FE-->>U: Detalle actualizado
        end
    end

    U->>FE: Abre /reservations o /history
    FE->>API: GET /api/reservations/mine
    API->>SVC: GetMyReservations
    SVC->>REP: Reservas propias o participadas
    REP->>SQL: SELECT reservations + participants
    SQL-->>REP: Lista
    REP-->>SVC: Reservas
    SVC-->>API: Lista sanitizada
    API-->>FE: 200

    opt Propietario o administrador cancela
        U->>FE: Confirma cancelación
        FE->>API: PATCH /api/reservations/cancel
        API->>SVC: CancelReservation
        SVC->>REP: Snapshot de dueño, estado y horario
        REP->>SQL: SELECT reservation
        alt Sin permiso, terminal o finalizada
            API-->>FE: 400 / 403 / 404
        else Cancelable
            SVC->>REP: CancelReservation
            REP->>SQL: UPDATE status = CANCELLED
            SQL-->>REP: Reserva cancelada
            REP-->>SVC: Resultado
            SVC-->>API: Resultado
            API-->>FE: 200
        end
    end
```

## Decisiones y divergencias

- Recursos y actividades usan handlers con acceso directo a repositorios; la
  política posee servicio, aunque la secuencia agrupa las lecturas para mantener
  legibilidad.
- `GET /api/reservations` es administrativo; el usuario normal usa
  `GET /api/reservations/mine`.
- La rama exitosa representa reserva no grupal. El flujo grupal está en
  [Secuencias MVP 2](secuencias-mvp2.md).
- No existe cutoff anticipado configurable de cancelación en este corte: se
  impide cancelar cuando la reserva ya finalizó.

## Fuentes

- [`backend/internal/routes/routes.go`](../../../backend/internal/routes/routes.go)
- [`backend/internal/handlers/reservations_handlers.go`](../../../backend/internal/handlers/reservations_handlers.go)
- [`backend/internal/services/reservations_service.go`](../../../backend/internal/services/reservations_service.go)
- [`backend/internal/repositories/reservations_repository.go`](../../../backend/internal/repositories/reservations_repository.go)
- [`frontend/src/services/reservations.service.js`](../../../frontend/src/services/reservations.service.js)
- [`frontend/src/stores/reservations.js`](../../../frontend/src/stores/reservations.js)
