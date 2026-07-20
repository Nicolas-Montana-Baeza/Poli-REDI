# Poli-REDI - Guia de documentacion tecnica y legibilidad

## Objetivo

Esta guia define como documentar codigo, SQL y documentos tecnicos de Poli-REDI sin cambiar comportamiento ni agregar comentarios de bajo valor.

La documentacion debe ayudar a entender reglas de negocio, contratos entre capas, decisiones tecnicas e invariantes que no sean evidentes al leer el codigo.

## Principio central

Antes de agregar un comentario, aplicar esta prueba:

```txt
¿Aporta informacion importante que un desarrollador competente no obtendria facilmente del codigo?
```

Si la respuesta es no, no se agrega.

## Que documentar

- Reglas de negocio.
- Contratos entre frontend, backend y base de datos.
- Autenticacion y autorizacion.
- Decisiones de arquitectura.
- Invariantes de seguridad.
- Manejo temporal y zona horaria.
- Transacciones, concurrencia y prevencion de conflictos.
- Constraints, indices, triggers y vistas de SQL Server.
- Efectos secundarios relevantes.
- Diferencias entre funcionalidad implementada y trabajo pendiente.

## Que evitar

- Narrar instrucciones obvias.
- Repetir nombres de funciones, variables o campos.
- Comentar cada linea.
- Agregar encabezados decorativos sin informacion nueva.
- Documentar intenciones no comprobadas.
- Usar `TODO` sin incertidumbre concreta.
- Introducir conceptos de PostgreSQL en scripts actuales de SQL Server.
- Recomendar `DEV_AUTH_ENABLED=true` fuera de desarrollo.
- Exponer secretos, tokens, passwords o cadenas de conexion reales.

## Go

Usar GoDoc cuando un identificador exportado necesite explicar contrato, permisos, efectos secundarios o invariantes.

Buenas razones para comentar:

- Un handler ignora datos enviados por el cliente por seguridad.
- Un service duplica una validacion que tambien existe en base de datos.
- Un repository convierte una hora entre contrato institucional y `DATETIME2`.
- Un error SQL se traduce a mensaje de dominio.

No comentar:

- Asignaciones directas.
- Validaciones autoexplicativas.
- Control de flujo evidente.

## Vue y JavaScript

Usar JSDoc o comentarios breves solo cuando aclaren:

- Contrato de props o emits no evidente.
- Estado compartido entre stores.
- Efectos reactivos con consecuencias externas.
- Diferencia entre datos del servidor, estado derivado y estado visual.
- Manejo de fechas que dependa de `America/Santiago`.

No comentar markup, estilos o bindings autoexplicativos.

## SQL Server y T-SQL

Documentar especialmente:

- Constraints que representan reglas de negocio.
- Triggers que previenen conflictos.
- Indices unicos filtrados.
- Vistas que actuan como contrato de lectura.
- Scripts destructivos o con precondiciones.
- Uso de `DATETIME2` para hora institucional de muro.
- Uso de `SYSUTCDATETIME()` para timestamps UTC.

No repetir cada columna si el nombre ya lo explica.

## Contratos criticos de Poli-REDI

- El cliente nunca decide la identidad del usuario autenticado.
- Las rutas administrativas deben validar rol admin en backend.
- `DEV_AUTH_ENABLED` es solo para desarrollo.
- Las reservas usan hora institucional de muro en `America/Santiago`.
- `DATETIME2` de agenda no debe tratarse como UTC.
- `created_at` y `updated_at` generados con `SYSUTCDATETIME()` representan UTC.
- Los modos `OPEN_USE`, `RESERVABLE`, `INFORMATIVE` y `ADMIN_ONLY` tienen semanticas diferentes.
- Los triggers de base de datos son parte efectiva del contrato, no solo respaldo tecnico.
- La disponibilidad puede combinar reservas, bloqueos y actividades institucionales.

## Cuando codigo y documentacion discrepan

Se debe conservar el comportamiento real del codigo y reportar la diferencia.

No inventar conciliaciones. Si la intencion no puede demostrarse, usar:

```txt
TODO(documentacion): confirmar <incertidumbre concreta>
```

## Validacion esperada

Segun el tipo de cambio:

- Backend Go: `gofmt` y `go test ./...`.
- Frontend: `npm run build`; `npm test` si existe y aplica.
- SQL: revision de sintaxis, orden de ejecucion, triggers y dependencias.
- Markdown: coherencia de rutas, estados, comandos y terminos.

Si una validacion no se ejecuta, se debe indicar el motivo.

## Relacion con otros documentos

Esta guia debe mantenerse coherente con:

- `docs/03-base-de-datos.md`
- `docs/04-backend.md`
- `docs/06-flujo-reservas.md`
- `docs/07-backlog.md`
- `docs/08-requisitos-historias-casos-uso.md`
- `docs/09-mvps-roadmap.md`
- `docs/10-guia-redeploy.md`
