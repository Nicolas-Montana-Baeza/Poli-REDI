# Contribuir a Poli-REDI

**Audiencia:** desarrollo, QA, DevOps y documentación.

**Propósito:** definir un flujo de cambio reproducible sin duplicar contratos técnicos ni procedimientos operativos.

**Estado:** vigente desde 2026-08-11.

**Fuente propietaria:** este archivo mantiene las normas de contribución; la operación pertenece a [`docs/05-instalacion-despliegue-recuperacion.md`](docs/05-instalacion-despliegue-recuperacion.md).

## Antes de cambiar

1. Leer el [estado actual](docs/01-estado-actual.md).
2. Identificar el requisito afectado en [trazabilidad](docs/03-requisitos-y-trazabilidad.md).
3. Revisar contratos y reglas en `docs/02`, `docs/04` y `docs/06`.
4. Confirmar la evidencia esperada en [calidad](docs/07-calidad-y-evidencia.md).
5. Preservar cambios ajenos y archivos no relacionados.

## Flujo de trabajo

- Usar una rama acotada y describir el objetivo funcional.
- Mantener frontend, backend, SQL, pruebas y documentación en el mismo incremento cuando cambie un contrato.
- No corregir históricos para representar el presente; registrar una nueva decisión o evidencia.
- No declarar una función cerrada por existir código sin prueba y ambiente identificados.
- No ejecutar migraciones propuestas o scripts destructivos por inferencia.

## Preparación local

### Backend

```powershell
cd backend
go mod download
go run ./cmd
```

### Frontend

```powershell
cd frontend
npm ci
npm run dev
```

La configuración completa, SQL Server local y variables se documentan en [`docs/05-instalacion-despliegue-recuperacion.md`](docs/05-instalacion-despliegue-recuperacion.md). Nunca copiar valores reales a ejemplos o capturas.

## Verificación mínima

Ejecutar solo los conjuntos aplicables y registrar el resultado real:

```powershell
cd backend
go test ./...

cd ..\frontend
npm test
npm run build
```

Una ejecución incompleta no equivale a aprobación ni a fallo del conjunto. Usar la plantilla de [`docs/07-calidad-y-evidencia.md`](docs/07-calidad-y-evidencia.md).

## Base de datos

- Base nueva: usar esquema y seed únicamente en un ambiente descartable autorizado.
- Base existente: aplicar migraciones `001`–`008` en orden.
- Ejecutar backup, precheck, postcheck, idempotencia y recuperación cuando corresponda.
- No ejecutar `009`; no pertenece a la secuencia aprobada.
- No reconstruir la única base para resolver un fallo de migración.

La decisión completa está en [`ADR-002`](docs/decisiones/ADR-002-evolucion-base-unica.md).

## Estilo de implementación

- Go: handlers delgados; reglas en servicios; persistencia en repositorios.
- Vue: componentes con responsabilidad clara; acceso remoto mediante servicios o stores.
- SQL: scripts idempotentes cuando sea posible y sin supuestos destructivos.
- API: actor, rol y estado controlados por servidor; errores públicos sin detalles internos.
- UI: teclado, foco visible, texto además de color y estados de carga/error/vacío.

## Documentación

- UTF-8 sin mojibake ni espacios finales.
- Máximo tres niveles de encabezado.
- Párrafos breves y una idea por viñeta.
- Tablas de hasta siete columnas.
- Enlaces relativos y nombres sin espacios para archivos nuevos.
- Un dato se mantiene solo en su documento propietario.
- Las decisiones duraderas se registran como ADR.

## Entrega del cambio

La descripción debe incluir:

- requisito o decisión afectada;
- archivos y contratos modificados;
- comandos ejecutados y resultados;
- ambiente y commit de la evidencia;
- migraciones o variables requeridas;
- riesgos residuales y pasos de recuperación.

El cambio no se presenta como cerrado hasta satisfacer el criterio correspondiente de [`docs/09-checklist-cierre.md`](docs/09-checklist-cierre.md).
