# Base de datos y migraciones de Poli-REDI

> **Estado:** CANDIDATO CANÓNICO DEL NUEVO ÁRBOL — contrastado con `database/schema.sql`<br>
> **Corte:** 2026-08-11<br>
> **Alcance:** modelo vigente, integridad y evolución acumulativa de la única base<br>
> **No demuestra:** que las migraciones pendientes hayan sido ejecutadas o validadas en Azure SQL

## Resumen

- Poli-REDI utiliza una sola base Azure SQL; no se crea otra base para MVP 2.
- `schema.sql` es apropiado para instalaciones nuevas o descartables. Una base
  existente evoluciona únicamente mediante migraciones aprobadas.
- La cadena oficial presente en `database/migrations/` llega de `001` a `008`.
- `009` existe solo en el árbol duplicado de propuesta mejorada y permanece
  **NO APROBADA**, fuera de la cadena oficial.
- El modelo entidad-relación completo se mantiene en un anexo para que esta
  guía se concentre en invariantes, ejecución y recuperación.

## 1. Entidades e invariantes

| Dominio | Entidades principales | Invariantes relevantes |
|---|---|---|
| Identidad | `users` | correo único, rol, bloqueo, RUT condicionado y write-once por API |
| Inventario | `venues`, `resources`, `activities` | modo de reserva, capacidad, actividad e instalación |
| Política | `reservation_policies` y tablas puente | versiones publicadas, scope, duraciones y recursos grupales |
| Reservas | `reservations`, `participants` | snapshot de política, owner único, capacidad congelada y objetivo válido |
| Código grupal | `reservation_join_code_secrets`, `reservation_group_expirations` | hash para búsqueda, secreto cifrado recuperable y expiración idempotente |
| Programación | `scheduled_activities`, `availability_blocks` | intervalos, recurso, autor y estado activo |
| Talleres | `workshops`, `workshop_occurrences`, `workshop_enrollments` | ocurrencias normalizadas, cupo, episodio activo único y conservación histórica |
| Trazabilidad | auditorías, `notifications`, `violations` | actor, instante, relación opcional y retención del episodio |

El [modelo entidad-relación completo](anexos/diagramas/modelo-entidad-relacion.md)
documenta las relaciones y sus cardinalidades.

## 2. Nulabilidad y estados

Las claves foráneas anulables del esquema no deben confundirse con el contrato
de las rutas. Por ejemplo, `reservations.resource_id` es nullable en SQL por
compatibilidad histórica, aunque `POST /api/reservations` exige un recurso
válido. También son opcionales:

- `reservations.activity_id`, necesario para `OPEN_USE`;
- `scheduled_activities.activity_id`;
- `notifications.reservation_id` y `violations.reservation_id`;
- `violations.created_by_user_id`, `audit_logs.user_id` y
  `reservation_policies.created_by_user_id`.

`ck_reservations_status` admite `PENDING`, `CONFIRMED`, `CANCELLED`, `REJECTED`
y `EXPIRED`. Las rutas actuales producen transiciones operativas entre los tres
primeros; no existe una ruta backend que produzca `REJECTED` o `EXPIRED` en
este corte.

## 3. Defensas de integridad

- Constraints de dominio protegen modos, estados, duraciones y valores
  positivos.
- Índices únicos filtrados impiden owner duplicado e inscripción activa
  duplicada sin borrar episodios cancelados.
- Triggers set-based complementan los repositorios ante escrituras SQL
  externas.
- Repositorios de reservas y talleres usan transacciones `SERIALIZABLE`,
  `UPDLOCK` y `HOLDLOCK`.
- El orden soportado para talleres bloquea primero usuario y luego taller.
- Los intervalos usan semántica semiabierta; horarios contiguos no chocan.
- Un recurso grupal debe estar permitido, tener capacidad suficiente y no ser
  `OPEN_USE`.

Los triggers no sustituyen las validaciones de dominio ni autorizan a ejecutar
scripts fuera de la secuencia aprobada.

## 4. Instalación limpia y evolución

### Base nueva o descartable

1. Ejecutar `database/schema.sql` con una herramienta compatible con `GO`.
2. Ejecutar `database/seed.sql` solo si corresponde al ambiente.
3. Configurar llaves de código grupal sin exponerlas.
4. Ejecutar pruebas y smoke test.

### Base existente

1. Obtener backup o export recuperable.
2. Identificar la última migración aplicada con evidencia.
3. Ejecutar solo la siguiente migración oficial completa, conservando `GO`.
4. Verificar preflight, postcheck, idempotencia y estado transaccional.
5. Continuar una por una; nunca sustituir la evolución por `schema.sql` o
   `seed.sql`.

No se debe ejecutar `drop.sql` sobre la base única.

## 5. Cadena oficial `001`–`008`

| Migración | Propósito resumido |
|---|---|
| `001_mvp2_group_participants.sql` | participantes y base de solicitud grupal |
| `002_mvp2_target_participants.sql` | objetivo, capacidad y política asociada |
| `003_open_use_frequency_scope.sql` | modos y alcance de frecuencia `OPEN_USE` |
| `004_group_flow_completion.sql` | hash, secreto, expiración, auditoría y cierre grupal |
| `005_rut_integrity_and_admin_exemption.sql` | integridad de RUT y excepción administrativa |
| `006_workshop_occurrences.sql` | ocurrencias normalizadas de talleres |
| `007_repair_bootstrap_group_policy.sql` | reparación acotada del bootstrap reconocido |
| `008_personal_overlap_includes_participations.sql` | solape personal incluye participaciones confirmadas |

`007` debe fallar de forma cerrada si no reconoce inequívocamente la política
bootstrap. `008` debe probar solape en ambas direcciones y permitir extremos
contiguos. Ambas siguen requiriendo ejecución y evidencia en una copia
recuperable antes de Azure SQL.

## 6. Migración `009`

La ruta
`database/poliredi_database_improved/poliredi_database_improved/migrations/009_database_hardening_and_consistency.sql`
contiene una **propuesta**. No forma parte de `database/migrations/`, no está
aprobada, no se considera aplicada y no acredita compatibilidad con Azure.

Antes de incorporarla debe existir una decisión registrada sobre:

1. objetivo y efecto prospectivo;
2. preflight contra datos reales;
3. idempotencia y orden respecto de `007`/`008`;
4. postcheck verificable;
5. rollback o restauración;
6. incorporación a la cadena oficial sin ejecutar el árbol duplicado.

## 7. Protocolo y recuperación

Por cada migración:

1. conservar versión, fecha, actor y ambiente;
2. ejecutar backup/export y probar que sea utilizable;
3. correr preflight sin forzar datos para que pase;
4. ejecutar el archivo completo con `GO`;
5. revisar indicadores y reejecutar para demostrar idempotencia;
6. conservar logs sanitizados y resultados;
7. ante fallo, detener publicación y abrir una sesión nueva;
8. comprobar `@@TRANCOUNT = 0` y `XACT_STATE() = 0`;
9. restaurar si no puede demostrarse un estado compatible.

## 8. Diagramas aún pendientes

No deben presentarse como cerrados hasta aprobar decisión e implementación:

- publicación prospectiva frente a corrección excepcional;
- migración completa con preflight, postcheck y rollback;
- bloqueo/desbloqueo administrativo y su auditoría;
- prioridad institucional frente a reservas existentes.

Los flujos funcionales vigentes están en
[Reglas y flujos](06-reglas-y-flujos.md).

## Fuentes

- [`database/schema.sql`](../database/schema.sql)
- [`database/migrations/README.md`](../database/migrations/README.md)
- [`database/migrations`](../database/migrations)
- [`backend/internal/repositories/reservations_repository.go`](../backend/internal/repositories/reservations_repository.go)
- [`backend/internal/repositories/participants_repository.go`](../backend/internal/repositories/participants_repository.go)
- [`backend/internal/repositories/workshops_repository.go`](../backend/internal/repositories/workshops_repository.go)
- [`backend/internal/repositories/reservation_policies_repository.go`](../backend/internal/repositories/reservation_policies_repository.go)
