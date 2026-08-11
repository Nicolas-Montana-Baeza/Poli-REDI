# Resumen ejecutivo y estado actual de Poli-REDI

**Estado documental:** CANÓNICO  
**Último incremento documentado:** 2026-08-04

## 1. Problema y objetivo

El proceso levantado depende de gestión manual y Google Calendar, concentra decisiones en el encargado y no ofrece una vista institucional única de disponibilidad, reglas ni trazabilidad.

Poli-REDI busca validar un prototipo web desacoplado que centralice recursos deportivos, disponibilidad, reservas particulares, solicitudes grupales e información administrativa básica, aplicando reglas críticas en servidor y base de datos.

## 2. Lectura correcta del estado

Poli-REDI es una **demo funcional avanzada**, no una plataforma institucional completamente cerrada. El sistema tiene evidencia local importante, pero siguen pendientes las migraciones reales sobre Azure SQL y la validación integrada online con Microsoft Entra ID, CORS, API y base desplegada.

## 3. Estado por área

| Área | Estado vigente | Límite de la evidencia |
|---|---|---|
| Autenticación e identidad | IMPLEMENTADO | Entra ID y modo local existen; la integración online del corte sigue pendiente. |
| Roles y autorización | IMPLEMENTADO Y REVISADO LOCALMENTE | El backend determina identidad, rol y bloqueo; la UI no concede permisos. |
| Perfil y RUT | IMPLEMENTADO | Usuarios normales requieren RUT para reservar e inscribirse; desinscribirse de un taller no exige RUT. |
| Recursos | IMPLEMENTADO PARCIAL | Catálogo y actualización de imagen; falta administración completa del inventario. |
| Disponibilidad | IMPLEMENTADO PARCIAL | Integra reservas, talleres y actividades; falta incorporar bloqueos al contrato completo y filtros backend por rango. |
| Reserva particular | IMPLEMENTADO LOCALMENTE | Servidor controla usuario, horario, duración, ventana, frecuencia y conflictos. |
| Flujo grupal | APROBABLE CONDICIONADO | Código, progreso, confirmación, retiro, rotación, deadline y expiración tienen evidencia local; Azure SQL real sigue pendiente. |
| Talleres | IMPLEMENTADO Y VERIFICADO LOCALMENTE | Alta, desinscripción idempotente, liberación de cupo, historial y reinscripción; integración online pendiente. |
| Cancelación | IMPLEMENTADO PARCIAL | Propietario o admin pueden cancelar reservas activas permitidas; falta cerrar coherencia operacional completa. |
| Notificaciones | IMPLEMENTADO PARCIAL | Consulta y contador existen; lectura y cobertura de eventos son incompletas. |
| Administración | IMPLEMENTADO PARCIAL | Panel y lecturas básicas; faltan usuarios, inventario, bloqueos, programación, conflictos e infracciones completos. |
| Reportes y auditoría | IMPLEMENTADO PARCIAL | Hay indicadores y registros, pero no una experiencia institucional completa. |
| Accesibilidad y UX | VERIFICADO LOCALMENTE PARCIAL | Foco, teclado, skeletons, estados asíncronos y privacidad fueron mejorados; falta QA visual manual multiancho. |
| Despliegue | CONFIGURADO | La disponibilidad actual de la demo no fue verificada al generar este paquete. |

## 4. Estado de los MVP

| MVP | Estado recomendado | Condición de cierre |
|---|---|---|
| MVP 1 — Base técnica | DEMOSTRABLE | Mantener pruebas y validar ambiente integrado cuando corresponda. |
| MVP 2 — Flujo de usuario | APROBABLE PARA CIERRE TÉCNICO CONDICIONADO | Ejecutar migraciones pendientes y validar flujo online completo. |
| MVP 3 — Administración institucional | PARCIAL | Completar inventario, bloqueos, programación, conflictos e historial institucional. |
| MVP 4 — Calidad y soporte | EN DESARROLLO | Cerrar reportes, notificaciones, pruebas, evidencia y soporte. |

## 5. Evidencia local más reciente documentada

| Verificación | Resultado |
|---|---|
| Backend | `go test ./... -count=1` aprobado en todos los paquetes. |
| Pruebas Node | 18 aprobadas. |
| Pruebas Vitest | 144 aprobadas. |
| Build frontend | Aprobado. |
| `diff-check` | Aprobado. |
| Desinscripción de talleres | Implementada y verificada localmente el 2026-08-04. |
| Azure SQL migraciones `007`/`008` | Pendiente. |
| Flujo online Entra/CORS/API/DB | Pendiente. |

La advertencia de bundle de 531.79 kB corresponde al último valor explícitamente documentado antes del incremento del 2026-08-04; no se dispone de una medición posterior en los archivos fuente.

## 6. Decisiones vigentes

1. El backend determina identidad, rol, bloqueo y capacidades.
2. La creación falla cerrada si no existe una política publicada válida.
3. La agenda usa hora institucional de muro `America/Santiago`; los timestamps técnicos usan UTC.
4. `OPEN_USE` no consume frecuencia grupal ni bloquea a otros usuarios, pero un mismo usuario no puede solapar su agenda personal.
5. Las tres canchas grupales exigen mínimo de 10 participantes; el propietario cuenta y no puede retirar su propia participación.
6. La solicitud grupal bloquea el horario desde `PENDING` y confirma al alcanzar el mínimo.
7. El código se obtiene bajo demanda, solo por el propietario y en estados permitidos.
8. Talleres e inscripciones propias forman parte de la ampliación controlada de MVP 2.
9. Clases, campeonatos y otros eventos institucionales pertenecen a MVP 3 y no demuestran asistencia personal sin una relación explícita.
10. Las migraciones `007` y `008` son prospectivas y no reescriben reservas históricas.

## 7. Pendientes que impiden declarar cierre total

### P2

- Ejecutar `007` y `008` en copia controlada y Azure SQL con backup, precheck, postcheck, segunda ejecución idempotente y prueba de recuperación.
- Ejecutar validación online de Microsoft Entra ID, CORS, API y base desplegada.
- Completar notificaciones y administración institucional fuera del flujo base de usuario.
- Ejecutar QA visual y de privacidad por audiencia en 377, 500, 768 y 1440 px.

### P3

- Dividir u optimizar el bundle frontend.
- Ampliar matriz manual de navegadores, teclado, lector de pantalla y movimiento reducido.
- Incorporar historial institucional de clases, campeonatos y eventos en MVP 3.

## 8. Mensaje recomendado para una presentación

> Poli-REDI cuenta con una base técnica funcional y un flujo de usuario avanzado verificado localmente. MVP 2 es aprobable para cierre técnico condicionado, porque aún deben ejecutarse las migraciones reales sobre Azure SQL y validarse de punta a punta la integración online.
