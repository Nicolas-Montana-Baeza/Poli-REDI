# Calidad y evidencia

**Audiencia:** QA, desarrollo, DevOps, Documentador y evaluación académica

**Propósito:** definir qué evidencia permite afirmar el estado de una capacidad

**Estado:** canónico, corte 2026-08-11

**Fuente:** auditoría técnica del 2026-08-11 y checklists preservados

## Resumen

Una prueba local no certifica Azure, Entra ID, migraciones ni experiencia visual. Cada ejecución debe registrar fecha, ambiente, versión, procedimiento, resultado y evidencia accesible. Las casillas de cierre viven únicamente en [09-checklist-cierre.md](09-checklist-cierre.md).

La auditoría revalidó 18 pruebas Node sobre el commit `939ba51`. No produjo evidencia nueva de finalización de Vitest, Go, build frontend ni operación online.

## Estado de evidencia al corte

| Verificación | Resultado documentado | Alcance de la afirmación |
|---|---|---|
| Node | 18 pruebas aprobadas | Revalidado en auditoría sobre `939ba51`. |
| Vitest | Ejecución iniciada, sin resultado final registrado | No puede declararse aprobada en este corte. |
| Go | Sin ejecución nueva registrada | Evidencia histórica no revalida el código actual. |
| Build frontend | Sin ejecución nueva registrada | Debe repetirse antes de cierre. |
| Migraciones `007`/`008` | Pendientes de copia recuperable y Azure SQL | Bloquean cierre integrado de MVP 2. |
| Azure/Entra/E2E | Sin nueva validación | No hay base para declarar MVP online. |

El detalle inmutable de la revisión está en [anexos/evidencia/auditoria-2026-08-11.md](anexos/evidencia/auditoria-2026-08-11.md).

La [bitácora self-hosted del 2026-08-13](anexos/operacion/despliegue-self-hosted-mvp2.md) añade evidencia de imágenes API MVP 1/MVP 2 aisladas con Azure SQL y Entra, y refiere merge/build/tests. No eleva el ambiente a “validado online” hasta demostrar Quadlet API+web estable y E2E funcional público.

## Niveles de evidencia

| Nivel | Requisito mínimo | Uso permitido |
|---|---|---|
| Definido | Requisito y criterio de aceptación aprobados | Planificación. |
| Implementado | Código o esquema observable | No implica que funcione. |
| Probado localmente | Ejecución reproducible con resultado | Demostración local acotada. |
| Integrado | Frontend, API y DB compatibles | Cierre técnico local. |
| Validado online | Azure, Entra, CORS y datos reales de prueba | Declaración de ambiente objetivo. |
| Aceptado | Checklist y riesgos residuales aprobados | Cierre del MVP. |

## Matriz mínima de pruebas

| Capa | Cobertura | Evidencia |
|---|---|---|
| Dominio backend | Políticas, permisos, solapes, estados y errores | Resultado `go test`, versión y casos. |
| Frontend unitario | Servicios, stores y componentes críticos | Resultado Vitest y reporte. |
| Frontend build | Compilación y rutas | Log de build y artefacto. |
| Base de datos | Precheck, ejecución, postcheck e idempotencia | Script, base/copia y consultas de control. |
| Integración | Identidad, frontend, API y DB | Matriz de casos y evidencias. |
| Azure | Entra, CORS, TLS, cuotas y smoke | URL/ambiente, versión y resultado seguro. |
| UX/accesibilidad | Estados, teclado, foco, contraste y responsive | Capturas seguras y registro por viewport. |

## Casos críticos

### Identidad y privacidad

- usuario normal y administrador autenticados por el modo correspondiente;
- rechazo de acceso administrativo a usuario normal;
- RUT solicitado solo al usuario normal cuando falta y es necesario;
- código grupal visible solo para audiencias autorizadas;
- respuestas inexistente, cancelada o vencida sin filtración de estado privado;
- logs y capturas sin tokens, secretos ni datos personales innecesarios.

### Reservas y agenda

- creación válida, contigüidad permitida y solape rechazado;
- `open_use` sin restricción semanal, pero con rechazo de solape personal;
- grupo pendiente hasta alcanzar mínimo;
- retiro, reconfirmación, rotación de código y expiración idempotente;
- cancelación autorizada antes de finalizar;
- concurrencia real en Azure SQL.

### Talleres

- alta con RUT cuando corresponde, cupo y taller activo;
- rechazo solo por otro taller activo solapado;
- no imponer taller↔reserva personal entre recursos en MVP 2;
- baja sin cutoff mientras el taller siga activo;
- segunda baja idempotente;
- reinscripción con nuevo episodio e historial preservado.

### Administración y público

- matriz de permisos y motivos obligatorios;
- inventario, bloqueos y programación con trazabilidad;
- prioridad institucional determinista y notificación específica;
- pantalla pública sanitizada y sin acciones privadas;
- auditoría visible únicamente para roles autorizados.

## Base de datos

Para `007` y `008` registrar por separado:

1. respaldo recuperable;
2. precheck;
3. primera ejecución;
4. postcheck de esquema y datos;
5. segunda ejecución idempotente;
6. prueba de recuperación;
7. no retroactividad y snapshots;
8. solapes, extremos contiguos y concurrencia;
9. repetición controlada en Azure SQL.

`009` no se prueba como parte del cierre: continúa como propuesta no aprobada.

## QA visual y accesibilidad

| Dimensión | Criterio |
|---|---|
| Viewports | 377, 500, 768 y 1440 px o matriz equivalente aprobada. |
| Estados | Skeleton inicial, actualización sin parpadeo, vacío, error y reintento. |
| Interacción | Tarjetas seleccionables, teclado, Escape, foco visible y restaurado. |
| Semántica | Nombres accesibles, anuncios, `aria-busy` y estructura comprensible. |
| Color | Texto además de color, contraste y chips consistentes. |
| Movimiento | Respeto de preferencias de movimiento reducido. |
| Privacidad | Contenido ajustado a usuario, administrador y público. |

## Registro de ejecución

| Campo | Valor esperado |
|---|---|
| Fecha y responsable por rol | Momento y rol que ejecutó/revisó. |
| Ambiente | Local, copia recuperable o Azure. |
| Commit/artefacto/esquema | Identificadores reproducibles. |
| Procedimiento | Comando, caso o guion versionado. |
| Resultado | Aprobado, con observaciones o fallido. |
| Evidencia | Log, captura sanitizada, reporte o acta. |
| Riesgo residual | Decisión y responsable de aceptación. |

## Criterio de salida

No se cierra un MVP con un P0/P1 abierto, fuga de datos, permiso incorrecto, migración no recuperable o incompatibilidad del ambiente objetivo. El responsable del cierre usa [09-checklist-cierre.md](09-checklist-cierre.md) y el hito de [08-plan-entrega-2026.md](08-plan-entrega-2026.md); este documento no duplica sus casillas ni fechas.
