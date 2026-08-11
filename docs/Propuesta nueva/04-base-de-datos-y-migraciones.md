# Base de datos y migraciones de Poli-REDI

**Estado:** CANÓNICO  
**Motor:** Azure SQL Database / T-SQL

## 1. Propósito

La base persiste el estado operativo y aplica una segunda línea de defensa para reglas críticas. La documentación explica el modelo; no reemplaza la ejecución de scripts ni sus postchecks.

## 2. Entidades principales

- `venues`: sedes y ubicación.
- `users`: identidad local, Entra, rol, RUT y bloqueo.
- `resources`: espacios y modos `RESERVABLE`, `OPEN_USE`, `INFORMATIVE`, `ADMIN_ONLY`.
- `activities`: catálogo de actividades.
- `reservations`: propietario, recurso, horario, estado, política y snapshots.
- `participants`: participación en solicitudes grupales.
- `reservation_join_code_secrets`: secreto cifrado recuperable owner-only.
- `reservation_group_expirations`: marca idempotente de expiración.
- `workshops`, `workshop_occurrences`, `workshop_enrollments`: oferta, horarios e inscripciones.
- `availability_blocks`, `scheduled_activities`: operación institucional.
- `notifications`, `audit_logs`, `violations`: comunicación y trazabilidad.

## 3. Reglas de integridad

- RUT canónico, único y de escritura protegida.
- Estados y duraciones permitidos mediante `CHECK`.
- Índices únicos filtrados para identidades y registros activos.
- Intervalos semiabiertos y control de solapes.
- Política versionada por solicitud.
- Snapshot de capacidad para flujo grupal.
- Auditoría append-only donde corresponde.
- Talleres activos e inscripciones `CONFIRMED` participan en conflictos.
- `OPEN_USE` permite concurrencia del recurso, pero no solape personal.

## 4. Instalación limpia vs. evolución

### Base nueva o ambiente descartable

1. `drop.sql`, solo cuando la destrucción esté autorizada.
2. `schema.sql` como modelo canónico.
3. `seed.sql` para datos iniciales controlados.
4. `verify.sql` si está disponible en el paquete de base mejorado.

### Base existente

No ejecutar `drop.sql`, `schema.sql` ni `seed.sql` como mecanismo de recuperación. Usar exclusivamente las migraciones incrementales en orden y conservar `GO`.

## 5. Secuencia documentada del repositorio original

1. `001_mvp2_group_participants.sql`
2. `002_mvp2_target_participants.sql`
3. `003_open_use_frequency_scope.sql`
4. `004_group_flow_completion.sql`
5. `005_rut_integrity_and_admin_exemption.sql`
6. `006_workshop_occurrences.sql`
7. `007_repair_bootstrap_group_policy.sql`
8. `008_personal_overlap_includes_participations.sql`

Las migraciones `007` y `008` están documentadas como pendientes de ejecución real sobre Azure SQL.

## 6. Relación con el paquete de base mejorado

El paquete generado previamente incluye una migración propuesta `009_database_hardening_and_consistency.sql`. Esa migración no forma parte de la evidencia original del repositorio y no debe considerarse desplegada. Debe revisarse, probarse en copia controlada y aprobarse antes de integrarla a la secuencia oficial.

## 7. Protocolo obligatorio por migración

1. Crear backup o export recuperable.
2. Abrir sesión nueva y comprobar:

```sql
SELECT @@TRANCOUNT AS transaction_count, XACT_STATE() AS transaction_state;
```

Ambos valores deben ser `0` antes de comenzar.

3. Ejecutar el archivo completo con herramienta compatible con `GO`.
4. Revisar prechecks y postchecks.
5. Ejecutar una segunda vez para demostrar idempotencia.
6. Ejecutar casos funcionales y de concurrencia.
7. Conservar evidencia de resultados.

## 8. Recuperación

Ante un fallo:

- detener el despliegue;
- abrir una sesión limpia;
- inspeccionar objetos sin escribir;
- no editar políticas o datos para “forzar” el postcheck;
- restaurar el backup si el estado parcial no puede explicarse y demostrarse compatible;
- no usar scripts destructivos en la base única.

## 9. Validaciones específicas pendientes

### Migración `007`

- reconocer solo el bootstrap inequívoco;
- fallar cerrada ante política administrada o estructura divergente;
- no modificar reservas históricas;
- verificar precheck, postcheck e idempotencia.

### Migración `008`

- rechazar reserva contra participación confirmada;
- rechazar confirmación contra reserva o participación solapada;
- permitir extremos contiguos;
- no reclasificar datos históricos.

## 10. Archivos relacionados

- Esquema y seeds mejorados: paquete `poliredi_database_improved` generado anteriormente.
- Contratos funcionales: [`06-flujos-y-reglas-de-negocio.md`](06-flujos-y-reglas-de-negocio.md).
- Operación y redeploy: [`05-instalacion-despliegue-y-recuperacion.md`](05-instalacion-despliegue-y-recuperacion.md).
- Pruebas: [`07-calidad-pruebas-y-checklists.md`](07-calidad-pruebas-y-checklists.md).
