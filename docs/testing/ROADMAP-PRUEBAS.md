# Roadmap de pruebas y aseguramiento de calidad — Poli-REDI

**Proyecto:** Poli-REDI
**Documento:** Roadmap de pruebas de software
**Versión:** 1.0
**Estado:** Propuesta de planificación
**Fecha:** 02-11-2026

---

## 1. Propósito

Este documento establece la estrategia de pruebas para Poli-REDI, definiendo las actividades necesarias para verificar el cumplimiento de los requisitos funcionales y no funcionales del sistema.

El roadmap organiza las pruebas por etapas de desarrollo y dominios funcionales, especificando qué se probará, qué herramientas se utilizarán, qué evidencias deben recopilarse y qué condiciones se deben cumplir para considerar una prueba aprobada.

## 2. Alcance

El plan considera:

- Frontend desarrollado con Vue y Vite.
- Backend desarrollado en Go mediante API REST.
- Persistencia en PostgreSQL.
- Autenticación institucional mediante Microsoft Entra ID.
- Integración entre frontend, backend y base de datos.
- Reglas de negocio relacionadas con recursos, actividades y reservas.
- Despliegue y ejecución mediante contenedores.

La planificación contempla pruebas funcionales, unitarias, de integración, rendimiento, concurrencia, seguridad, regresión y aceptación.

**Nota:** la asignación definitiva de funcionalidades a cada MVP debe contrastarse con los requisitos aprobados y la implementación existente. La presencia de una tabla, ruta o componente no demuestra por sí sola que una funcionalidad esté terminada.

## 3. Organización de las pruebas por MVP

### 3.1. MVP 1 — Funcionalidades base

**Objetivo:** verificar las operaciones fundamentales del sistema de reservas.

Pruebas previstas:

- Autenticación y acceso al sistema.
- Consulta de recursos disponibles.
- Consulta de disponibilidad.
- Creación de reservas.
- Visualización de reservas.
- Cancelación de reservas.
- Validación de permisos de usuario.

### 3.2. MVP 2 — Reservas grupales

**Objetivo:** verificar las reglas de participación y confirmación asociadas a las reservas grupales.

Pruebas previstas:

- Creación de reservas grupales.
- Generación y utilización de enlaces de invitación.
- Registro de participantes.
- Confirmación de participación.
- Validación del mínimo de diez participantes.
- Restricción de una reserva por usuario cada siete días.
- Prevención de participantes duplicados.
- Validación de estados y vencimientos.
- Consistencia de la información registrada.

### 3.3. MVP 3 — Administración y reglas institucionales

**Objetivo:** verificar la administración de recursos y la aplicación de las reglas institucionales.

Pruebas previstas:

- Administración de recursos.
- Configuración de disponibilidad.
- Bloqueos por mantención.
- Prevención de solapamientos.
- Reservas asociadas a actividades EFI.
- Prioridad administrativa.
- Validación de permisos administrativos.
- Aplicación de restricciones institucionales.

### 3.4. MVP 4 — Funcionalidades complementarias

**Objetivo:** validar las funcionalidades complementarias contempladas en el alcance final del proyecto.

Ámbitos propuestos, sujetos a confirmación con los requisitos y la implementación:

- Solicitudes de campeonatos.
- Registro de incumplimientos.
- Restricciones y bloqueos de usuarios.
- Reportes y estadísticas.
- Trazabilidad de operaciones.
- Validación de información consolidada.

No se considerará implementada una funcionalidad hasta contar con su requisito, implementación verificable y caso de prueba.

## 4. Etapas del roadmap

| Etapa | Actividad | Resultado esperado |
|---|---|---|
| 0 | Preparación del entorno | Ambiente de pruebas identificado y operativo |
| 1 | Inventario técnico y funcional | Endpoints, módulos y reglas identificados |
| 2 | Pruebas funcionales MVP 1 | Funcionalidades base verificadas |
| 3 | Pruebas funcionales MVP 2 | Reglas de reservas grupales verificadas |
| 4 | Pruebas administrativas MVP 3 | Reglas institucionales verificadas |
| 5 | Pruebas MVP 4 | Funcionalidades complementarias verificadas |
| 6 | Pruebas transversales | Rendimiento, seguridad y regresión evaluados |
| 7 | Pruebas de aceptación | Cumplimiento del alcance validado |
| 8 | Cierre y documentación | Informe final de pruebas |

### Etapa 0 — Preparación del entorno

**Actividades:**

- Verificar disponibilidad del frontend y backend.
- Verificar conexión con PostgreSQL y autenticación.
- Identificar ambiente, versiones de aplicación y dependencias.
- Preparar usuarios y datos de prueba.

**Evidencias:** versiones del software, configuración del ambiente sin secretos, capturas, resultado de endpoints de salud y registro de datos de prueba.

**Herramientas:** Podman, PostgreSQL/`psql`, navegador, Git y `curl`.

### Etapa 1 — Inventario técnico y funcional

**Actividades:**

- Identificar endpoints disponibles.
- Relacionar endpoints con requisitos.
- Identificar operaciones de lectura y escritura.
- Identificar reglas de negocio, roles y permisos.
- Identificar funcionalidades pendientes.

**Evidencias:** inventario de endpoints, matriz requisito–funcionalidad, listado de reglas de negocio y registro de funcionalidades pendientes.

**Herramientas:** código fuente Go, PostgreSQL, Git y Postman o Bruno.

### Etapa 2 — Pruebas funcionales MVP 1

**Actividades:** ejecutar escenarios positivos y negativos, verificar respuestas HTTP, persistencia y permisos.

**Evidencias:** solicitudes y respuestas HTTP, capturas de interfaz, consultas SQL y resultado de cada caso de prueba.

### Etapa 3 — Pruebas funcionales MVP 2

**Actividades:** crear una reserva grupal, registrar participantes, validar quorum, confirmar participación, verificar restricciones semanales, duplicidad, persistencia y estados.

**Evidencias:** requests y responses, estados de las reservas, participantes registrados, consultas SQL antes y después, y logs del backend.

### Etapa 4 — Pruebas administrativas MVP 3

**Actividades:** configurar recursos, modificar disponibilidad, crear bloqueos, verificar conflictos, prioridad administrativa y autorización de operaciones.

**Evidencias:** capturas administrativas, respuestas de API, registros persistidos y casos de conflicto y autorización.

### Etapa 5 — Pruebas MVP 4

**Actividades:** ejecutar escenarios asociados a las funcionalidades complementarias efectivamente implementadas, verificar reglas, estados, registros históricos y reportes.

**Evidencias:** registros de ejecución, reportes generados, consultas SQL y capturas de interfaz.

### Etapa 6 — Pruebas transversales

**Actividades:** pruebas unitarias, integración, carga, estrés, concurrencia, seguridad, regresión e integridad.

**Evidencias:** resultados automatizados, métricas de rendimiento, logs, capturas y registro de defectos.

### Etapa 7 — Pruebas de aceptación

**Actividades:** ejecutar escenarios representativos, verificar criterios de aceptación y requisitos principales, registrar observaciones y comprobar resolución de defectos críticos.

**Evidencias:** matriz de aceptación, registro de observaciones, acta de validación y resultado de escenarios.

### Etapa 8 — Cierre

**Actividades:** consolidar resultados, revisar defectos pendientes, documentar limitaciones, preparar informe final y organizar evidencias.

**Evidencias:** informe de pruebas, matriz de trazabilidad, resumen de defectos y evidencias organizadas por etapa.

## 5. Matriz de tipos de prueba y herramientas

| Tipo de prueba | Objetivo | Herramienta | Evidencia |
|---|---|---|---|
| Unitaria | Validar funciones y reglas aisladas | Go test | Resultado de ejecución |
| Integración | Validar comunicación entre componentes | Go test + PostgreSQL | Logs y resultados |
| Funcional | Verificar operaciones de negocio | Postman / Bruno | Requests y responses |
| Interfaz | Verificar comportamiento visual | Navegador / Playwright | Capturas y reportes |
| Seguridad | Verificar autenticación y autorización | Postman / Bruno | Resultados HTTP |
| Rendimiento | Medir tiempos de respuesta | k6 | Métricas de latencia |
| Carga | Evaluar comportamiento bajo demanda | k6 | Throughput y errores |
| Estrés | Identificar degradación bajo carga creciente | k6 | Métricas y límites observados |
| Concurrencia | Detectar conflictos por operaciones simultáneas | k6 | Resultados de solicitudes |
| Integridad | Verificar consistencia de persistencia | PostgreSQL | Consultas SQL |
| Regresión | Detectar fallos introducidos por cambios | Go test + k6 | Comparación de ejecuciones |
| Aceptación | Validar requisitos del usuario | Navegador y escenarios guiados | Acta de aceptación |

Las herramientas se utilizarán según el tipo de prueba. No es necesario ejecutar todas las herramientas en cada etapa.

## 6. Criterios de aprobación

Cada caso de prueba debe tener un resultado esperado previamente definido.

**Estados permitidos:**

- **APROBADO:** el resultado obtenido coincide con el esperado.
- **RECHAZADO:** existe una diferencia o defecto.
- **BLOQUEADO:** no puede ejecutarse por una dependencia o problema externo.
- **NO EJECUTADO:** todavía no se ha realizado la prueba.

Umbrales iniciales propuestos para pruebas de rendimiento:

| Indicador | Umbral propuesto |
|---|---|
| Solicitudes fallidas | Menos del 1 % |
| Latencia p95 | Menor a 500 ms |
| Latencia p99 | Menor a 1.000 ms |

Estos valores son objetivos de evaluación y deben contrastarse con los requisitos no funcionales definitivos. Una prueba de carga no se considera aprobada únicamente por alcanzar una cantidad determinada de usuarios virtuales; también deben evaluarse errores, latencias, estabilidad y comportamiento de los componentes.

## 7. Ficha estándar de caso de prueba

| Campo | Descripción |
|---|---|
| Identificador | Código único del caso |
| MVP | Etapa funcional asociada |
| Requisito | Identificador del requisito relacionado |
| Nombre | Nombre del escenario |
| Objetivo | Qué se desea comprobar |
| Precondiciones | Estado necesario para ejecutar |
| Datos de entrada | Información utilizada |
| Procedimiento | Pasos de ejecución |
| Resultado esperado | Comportamiento definido |
| Resultado obtenido | Comportamiento observado |
| Estado | Aprobado, rechazado, bloqueado o no ejecutado |
| Herramienta | Herramienta utilizada |
| Evidencia | Archivo, captura o registro |
| Fecha | Fecha de ejecución |
| Versión | Versión del sistema evaluada |
| Observaciones | Defectos o comentarios |

### Ejemplo de caso

**ID:** `TC-MVP2-001`
**Nombre:** Validación del mínimo de participantes.

- Objetivo: comprobar que una reserva grupal no alcanza el quorum con menos de diez participantes.
- Precondición: usuario autenticado y recurso disponible.
- Entrada: nueve participantes registrados.
- Resultado esperado: la reserva no cumple el mínimo requerido.
- Herramientas: Postman o Bruno y PostgreSQL.
- Evidencia: respuesta de API y consulta de estado en base de datos.
- Estado: pendiente de ejecución.

## 8. Organización de evidencias

Estructura propuesta:

```text
testing/
├── 00-plan/
│   ├── ROADMAP-PRUEBAS.md
│   ├── test-strategy.md
│   └── test-environment.md
├── 01-mvp1/
│   ├── test-cases.md
│   ├── postman/
│   └── evidence/
├── 02-mvp2/
│   ├── test-cases.md
│   ├── postman/
│   └── evidence/
├── 03-mvp3/
│   ├── test-cases.md
│   └── evidence/
├── 04-mvp4/
│   ├── test-cases.md
│   └── evidence/
├── 05-performance/
│   ├── scripts/
│   ├── results/
│   └── reports/
├── 06-security/
├── 07-regression/
└── 08-final-report/
```

Las evidencias deberán identificarse con el código del caso, incluir fecha y versión, mantener relación con el requisito, evitar credenciales o información sensible y conservar los resultados originales de las herramientas automatizadas.

## 9. Estado inicial de las pruebas

### 9.1. Prueba de carga: 50 usuarios virtuales

**Endpoint evaluado:** `GET /api/resources`.

| Indicador | Resultado |
|---|---:|
| Usuarios virtuales máximos | 50 |
| Solicitudes completadas | 6.048 |
| Solicitudes por segundo | 33,59 |
| Errores HTTP | 0 % |
| Latencia promedio | 15,13 ms |
| Latencia p95 | 34,87 ms |
| Latencia p99 | 65,64 ms |

**Resultado:** umbrales configurados cumplidos.

### 9.2. Prueba de estrés: hasta 100 usuarios virtuales

Se informó la finalización satisfactoria de una prueba progresiva que alcanzó 100 usuarios virtuales.

**Estado:** resultado informado, pendiente de incorporar el reporte numérico completo y sus evidencias originales.

### 9.3. Alcance de los resultados

Las pruebas anteriores evalúan el endpoint de consulta de recursos. No constituyen evidencia de rendimiento de todos los procesos de reserva ni de la capacidad integral del sistema.

Las pruebas de disponibilidad, creación de reservas, registro de participantes y confirmación deberán ejecutarse como escenarios independientes.

## 10. Orden de ejecución recomendado

| Orden | Actividad | Prioridad | Estado inicial |
|---:|---|---|---|
| 1 | Inventario de endpoints y funcionalidades | Alta | Pendiente |
| 2 | Matriz de requisitos y casos de prueba | Alta | Pendiente |
| 3 | Pruebas funcionales MVP 1 | Alta | Pendiente |
| 4 | Pruebas funcionales MVP 2 | Alta | Pendiente |
| 5 | Pruebas de concurrencia de reservas | Alta | Pendiente |
| 6 | Pruebas administrativas MVP 3 | Media | Pendiente |
| 7 | Pruebas MVP 4 | Media | Pendiente de alcance |
| 8 | Pruebas de seguridad | Alta | Pendiente |
| 9 | Pruebas de regresión | Alta | Pendiente |
| 10 | Pruebas de aceptación | Alta | Pendiente |
| 11 | Informe final | Alta | Pendiente |

## 11. Condiciones de cierre

El proceso de pruebas se considerará finalizado cuando:

1. Los casos de prueba definidos para el alcance estén ejecutados.
2. Los resultados estén documentados.
3. Los defectos críticos estén resueltos o formalmente justificados.
4. Las evidencias estén organizadas y vinculadas a sus casos.
5. Los requisitos tengan trazabilidad hacia sus pruebas.
6. Se haya elaborado el informe final.

## 12. Consideraciones finales

El roadmap establece una estrategia progresiva de validación para Poli-REDI. Las pruebas funcionales se ejecutarán por MVP, mientras que las pruebas de rendimiento, seguridad, integridad y regresión se aplicarán transversalmente.

La ejecución se realizará de manera incremental, priorizando primero la identificación de las funcionalidades implementadas y sus requisitos asociados.

Los resultados de cada etapa servirán como evidencia técnica para evaluar la calidad, estabilidad y cumplimiento funcional del sistema.
