# Poli-REDI - Plan de inicio MVP4

Fecha: 2026-11-11
Rama: feature/mvp4
Base: fc304f0
Estado: Preparacion iniciada

## 1. Proposito

Consolidar la entrega, calidad, seguridad, soporte y
funcionalidades complementarias del prototipo Poli-REDI.

El MVP4 hereda el alcance tecnico validado de MVP3.
No modifica retroactivamente su cierre.

## 2. Alcance preliminar

### Seguridad
- SEC-005: Sanitizacion de errores HTTP.
- SEC-006: Identidad estable Entra mediante tenant y OID.
- Revision de dependencias npm y vulnerabilidades.

### Experiencia de usuario
- BACK-021: Responsive y shell compartido.
- BACK-022: Accesibilidad de modales y calendario.
- QA-002: Regresion frontend ampliada.

### Funcionalidad
- REP-001: Reportes administrativos ampliados.
- REP-002: Infracciones y auditoria.
- Confirmar alcance de solicitudes de campeonatos.

### Calidad e infraestructura
- QA-001: Regresion backend y PostgreSQL.
- INFRA-007: Completar evidencia restante de migraciones.
- CI reproducible, despliegue seguro y rollback.
- Documentacion y trazabilidad de requisitos.

### Operacion institucional
- EV-010: Ratificacion institucional pendiente.
- OPS-001: Corte Google Calendar solo si se aprueba
  el paso a operacion real.

## 3. Linea base verificada

Fecha: 2026-11-11
Commit inicial: fc304f0

- go test ./...: PASS.
- go vet ./...: PASS.
- npm test: 43/43 PASS.
- npm run test:components: 73/73 PASS.
- npm run build: PASS.

La suite Go normal no sustituye las pruebas de
integracion PostgreSQL en entorno efimero.

## 4. Hallazgos de preparacion

- npm ci reporto 7 vulnerabilidades:
  5 moderadas y 2 altas.
- lucide-vue-next esta deprecado.
- Vite advierte el uso de __dirname.
- NOTIF-001 fue cerrado tecnicamente en MVP3.
- El instalador MVP3 incluye PG16_0001 a PG16_0015;
  verificar criterios restantes de INFRA-007.
- Revisar y reconciliar estados historicos del backlog.

## 5. Orden inicial de trabajo

1. Auditar dependencias y riesgos de seguridad.
2. Implementar SEC-005 con regresion automatizada.
3. Analizar SEC-006 y definir migracion segura.
4. Resolver BACK-021 y BACK-022.
5. Completar REP-001 y REP-002.
6. Consolidar QA, CI, infraestructura y documentacion.

## 6. Condiciones de aceptacion

Cada tarea debe incluir:

- Requisitos y criterios de aceptacion identificados.
- Implementacion verificada.
- Pruebas positivas, negativas y de permisos aplicables.
- Evaluacion de seguridad y regresiones.
- Documentacion actualizada cuando corresponda.
- Evidencia reproducible y commit trazable.

El cierre del prototipo no representa certificacion
de produccion ni aceptacion institucional.

## 7. Restricciones

- No modificar el despliegue MVP3 validado.
- No ejecutar migraciones sobre datos habituales
  durante la preparacion.
- No almacenar credenciales ni tokens.
- No ejecutar npm audit fix --force automaticamente.
- Preservar el worktree y los cambios locales de MVP3.

## 8. Primera auditoria de dependencias

Fecha: 2026-11-11

### Resultado inicial

- 7 vulnerabilidades: 2 altas y 5 moderadas.
- Una vulnerabilidad alta afectaba dependencias de produccion.
- Se aplico `npm audit fix` sin `--force`.

### Correcciones verificadas

- follow-redirects: 1.16.0 a 1.16.3.
- happy-dom: 21.0.1 a 21.0.4.
- npm audit --omit=dev: 0 vulnerabilidades.
- npm audit completo: 0 altas, 5 moderadas.

### Regresion

- npm test: 43/43 PASS.
- npm run test:components: 73/73 PASS.
- npm run build: PASS.
- git diff --check: PASS.

### Riesgos residuales

Las cinco alertas moderadas corresponden a la cadena
@vue/test-utils, js-beautify, editorconfig, glob y minimatch.

No ejecutar npm audit fix --force sin evaluar primero
compatibilidad y regresiones.

### Pendientes independientes

- Evaluar la migracion de lucide-vue-next a @lucide/vue.
- Corregir la advertencia __dirname de Vite.
- Iniciar SEC-005: sanitizacion de errores HTTP.
