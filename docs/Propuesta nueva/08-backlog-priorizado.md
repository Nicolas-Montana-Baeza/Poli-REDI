# Backlog priorizado de Poli-REDI

**Estado:** RESUMEN CANÓNICO  
**Fuente detallada:** [`referencia/07-backlog-completo.md`](referencia/07-backlog-completo.md)

## 1. Regla de lectura

El backlog representa trabajo, no evidencia. Un ítem marcado como terminado debe respaldarse con código, prueba o acta fechada.

## 2. Prioridad inmediata — P2 de cierre

| ID sugerido | Trabajo | Criterio de cierre |
|---|---|---|
| DB-007 | Ejecutar migración `007` en copia y Azure SQL | Backup, pre/postcheck, idempotencia, no retroactividad y recuperación. |
| DB-008 | Ejecutar migración `008` en copia y Azure SQL | Casos bidireccionales de solape, contigüidad e idempotencia. |
| QA-ONLINE | Validar Entra ID, CORS, API y DB desplegada | Flujo de usuario y admin completo con evidencia. |
| QA-VISUAL | Validar responsive, teclado, lector y privacidad | Matriz 377/500/768/1440 px aprobada. |
| NOTIF-CORE | Completar lectura y eventos de notificación | Estados, destinos y acciones observables. |
| ADMIN-CORE | Completar operación institucional priorizada | Alcance explícito, permisos y auditoría. |

## 3. Prioridad posterior — P3

| Trabajo | Resultado esperado |
|---|---|
| Optimización del bundle | División o reducción sin regresiones. |
| Matriz de navegadores | Evidencia manual reproducible. |
| Historial institucional MVP 3 | Relación explícita usuario-actividad; no inferir asistencia. |
| Reportes institucionales | Backend y vistas SQL como fuente, no cálculos aislados de UI. |
| Consulta administrativa de auditoría | Acceso protegido, filtros y minimización. |

## 4. Administración pendiente

- inventario oficial completo;
- activación, modos y datos de recursos;
- usuarios y bloqueos;
- disponibilidad y actividades institucionales;
- resolución de conflictos;
- infracciones;
- políticas de reserva y corrección excepcional auditada.

## 5. Deuda técnica y documental

- eliminar referencias vigentes a PostgreSQL fuera del histórico;
- mantener sincronizados endpoints, requisitos y casos de uso;
- registrar el ambiente de cada prueba;
- revisar rutas y links Markdown;
- evitar estados absolutos como “cerrado” sin evidencia integrada;
- aprobar o descartar formalmente la migración propuesta `009` antes de incluirla en producción.

## 6. Orden recomendado

1. Congelar alcance de cierre de MVP 2.
2. Validar migraciones en copia recuperable.
3. Validar integración online.
4. Ejecutar QA visual, accesibilidad y privacidad.
5. Resolver defectos de cierre.
6. Registrar acta final.
7. Solo después iniciar ampliaciones de MVP 3.

## 7. Definición de terminado

Una tarea se considera terminada cuando:

- el alcance está explícito;
- la implementación existe;
- las pruebas aplicables pasan;
- la documentación se actualiza;
- no expone secretos ni datos indebidos;
- existe evidencia del ambiente objetivo cuando corresponde;
- los riesgos residuales están registrados.
