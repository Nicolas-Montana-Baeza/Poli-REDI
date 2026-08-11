# Auditoría de alcance e implementación — 2026-08-11

**Audiencia:** orquestación, Arquitecto, QA, Documentador y evaluación académica

**Propósito:** preservar el corte de evidencia y brechas observado el 2026-08-11

**Estado:** evidencia histórica inmutable del corte

**Fuente:** revisión de código, documentación y ejecuciones disponibles; commit base `939ba51`

## Resumen ejecutivo

- MVP 1 era demostrable localmente, pero no estaba revalidado online.
- MVP 2 estaba parcial y dependía de integración frontend de disponibilidad, migraciones `007`/`008` y Azure.
- MVP 3 tenía piezas parciales, sin cierre administrativo integrado.
- MVP 4 estaba pendiente y acotado a calidad, soporte, notificaciones core, reportes básicos, auditoría y despliegue.
- La disponibilidad por rango existía en backend; el frontend principal seguía pendiente de adopción completa.
- `009` era una propuesta no aprobada y no debía ejecutarse.

## Método y límites

La revisión contrastó documentos, rutas, servicios, modelos, frontend y scripts de base de datos. Separó existencia de código, prueba local, integración y validación online.

No se realizó una nueva campaña completa de QA, migración Azure ni publicación. Por tanto, este anexo no certifica la operación actual del ambiente online.

## Evidencia ejecutada

| Verificación | Resultado | Interpretación |
|---|---|---|
| Pruebas Node | 18 aprobadas | Revalidadas sobre `939ba51`. |
| Vitest | Inicio observado, sin cierre registrado | No se declara aprobado en este corte. |
| Go | Sin nueva ejecución registrada | No revalida resultados históricos. |
| Build frontend | Sin nueva ejecución registrada | Pendiente para cierre. |
| Azure/Entra/E2E | Sin nueva ejecución | No se declara MVP online. |
| `007`/`008` | No validadas en copia y Azure | Dependencia de MVP 2. |

## Dictamen por capacidad

| Capacidad | Observación | Estado al corte |
|---|---|---|
| Identidad, roles y RUT | Flujo local observable; Entra online no revalidado | Parcial integrado |
| Catálogo y reserva básica | Funciones locales disponibles | MVP 1 local demostrable |
| Disponibilidad por rango | Backend existente; adopción frontend incompleta | P1 pendiente |
| Reserva grupal | Código, participantes, estados y UI parciales | MVP 2 parcial |
| Privacidad del código | Contratos por audiencia existentes; E2E pendiente | Parcial |
| Talleres | Alta, baja, reinscripción e historial local observables | Parcial; online pendiente |
| Migraciones | `001`–`008` inventariadas; `007`/`008` sin cierre | P0 operativo |
| Administración | Componentes y modelos parciales | MVP 3 parcial |
| Pantalla pública | No cerrada | MVP 3 pendiente |
| Notificaciones y reportes | Fragmentos o borradores | MVP 4 pendiente |
| Seguridad/accesibilidad | Controles parciales, campaña pendiente | MVP 4 pendiente |
| Despliegue Azure | Evidencia histórica no revalidada | No verificado |

## Hallazgos priorizados

### P0 — integridad y ambiente

1. Ensayar y ejecutar `007` y `008` con backup, pre/postcheck, idempotencia y recuperación.
2. Mantener `009` fuera de la secuencia hasta una aprobación formal.
3. Validar esquema, aplicación y datos como una unidad antes de cerrar MVP 2.
4. Conservar una sola base y evitar scripts destructivos o DDL improvisado.

### P1 — integración funcional

1. Integrar en frontend el contrato de disponibilidad por rango.
2. Eliminar reconstrucción paralela de talleres cuando el contrato unificado sea fuente suficiente.
3. Propagar política, intervalos y usuario actual de forma coherente.
4. Representar `open_use`, bloqueos y tipos institucionales sin aplicar restricciones indebidas.
5. Filtrar recursos y alinear slots sin romper selección, timeline ni responsive.
6. Completar E2E de grupo, talleres, privacidad y solapes.

### P2 — cierre institucional

1. Completar CRUD y permisos para usuarios, inventario y bloqueos.
2. Definir y probar prioridad institucional con motivos y notificación específica.
3. Crear pantalla pública sanitizada e historial institucional.
4. Completar sistema core de notificaciones, reportes básicos y auditoría consultable.

### P3 — deuda y decisiones futuras

1. Retirar parsers y contratos legacy cuando el nuevo flujo esté estabilizado.
2. Resolver texto dañado y archivos comprimidos/binarios fuera de ubicación canónica.
3. Consolidar el árbol de base y archivar fuentes supersedidas solo con manifiesto.
4. Decidir el futuro de `009` mediante ADR independiente.
5. Mantener BI avanzado, IA, integración académica, multisede y Google bidireccional fuera del prototipo.

## Decisiones funcionales fijadas

- `open_use` queda libre de frecuencia semanal, pero no de solape personal.
- Talleres en MVP 2 controlan taller↔taller; no taller↔reserva entre recursos.
- Alta y baja de taller se permiten mientras esté activo, sin cutoff adicional.
- Cancelar reserva se permite antes de finalizar; cutoff configurable es futuro.
- Cambios administrativos son prospectivos y las reservas conservan snapshots.
- Historial de clases y eventos institucionales pertenece a MVP 3.

## Referencias del nuevo canon

- Estado vigente: [01-estado-actual.md](../../01-estado-actual.md)
- Requisitos: [03-requisitos-y-trazabilidad.md](../../03-requisitos-y-trazabilidad.md)
- Evidencia operativa: [07-calidad-y-evidencia.md](../../07-calidad-y-evidencia.md)
- Plan: [08-plan-entrega-2026.md](../../08-plan-entrega-2026.md)
- Checklist: [09-checklist-cierre.md](../../09-checklist-cierre.md)

Este anexo preserva lo observado el 2026-08-11. Las ejecuciones posteriores deben registrarse como nueva evidencia; no se modifica este corte para representar resultados futuros.
