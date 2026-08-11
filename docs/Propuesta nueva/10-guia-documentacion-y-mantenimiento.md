# Guía de documentación y mantenimiento de Poli-REDI

**Estado:** NORMA INTERNA

## 1. Principio central

Documentar decisiones, contratos e invariantes que un desarrollador competente no obtiene fácilmente del código. No narrar lo obvio.

## 2. Qué documentar

- reglas de negocio;
- permisos y privacidad;
- contratos frontend/backend/base;
- tiempo, zona horaria e intervalos;
- transacciones y concurrencia;
- constraints, índices, triggers y vistas;
- efectos secundarios y auditoría;
- diferencias entre aprobado, implementado y verificado;
- recuperación y precondiciones de scripts peligrosos.

## 3. Qué evitar

- comentar cada línea;
- copiar nombres de funciones como explicación;
- documentar intenciones no comprobadas;
- usar `TODO` sin incertidumbre concreta;
- mezclar PostgreSQL con el stack vigente;
- recomendar modo dev en ambientes públicos;
- incluir secretos o URLs internas sensibles;
- declarar cierre por un ítem de backlog.

## 4. Go

Usar GoDoc cuando un identificador exportado requiera explicar contrato, permiso, efecto secundario o invariante. Comentar especialmente la identidad tomada del contexto, traducción de errores SQL, transacciones y conversión temporal.

## 5. Vue y JavaScript

Documentar contratos de props/emits no evidentes, estado derivado, efectos reactivos externos, deduplicación, respuestas obsoletas y manejo de `America/Santiago`. Evitar comentarios sobre markup y bindings autoexplicativos.

## 6. SQL Server

Documentar reglas representadas por constraints y triggers, índices únicos filtrados, vistas contractuales, locks, orden transaccional, `DATETIME2` de agenda y timestamps UTC.

## 7. Cabecera mínima de un documento

```markdown
# Título

**Estado:** CANÓNICO / OPERATIVO / EVIDENCIA / HISTÓRICO
**Corte:** AAAA-MM-DD
**Propósito:** ...
**Fuentes o evidencias:** ...
```

## 8. Registro de evidencia

Toda verificación debe incluir:

- fecha;
- ambiente;
- commit o versión;
- comando o pasos;
- resultado;
- observaciones y limitaciones.

## 9. Manejo de discrepancias

1. conservar el comportamiento real observado;
2. no inventar conciliaciones;
3. registrar la contradicción;
4. escalar una decisión cuando afecte alcance;
5. actualizar documentos dependientes una vez resuelta.

Formato recomendado:

```text
TODO(documentación): confirmar <incertidumbre concreta> con <responsable o fuente>.
```

## 10. Validación por tipo de cambio

| Cambio | Validación mínima |
|---|---|
| Go | `gofmt`, `go test ./...` |
| Frontend | pruebas aplicables y `npm run build` |
| SQL | sintaxis, orden, dependencias, postcheck e idempotencia |
| Markdown | enlaces, términos, estados, comandos y fechas |
| Despliegue | health, identidad, CORS, API, DB y reversa |

## 11. Control de duplicados

- No crear copias con sufijos `(1)`, `(2)` o “final-final”.
- Usar Git para historial.
- Mantener un documento canónico por tema.
- Mover cortes anteriores a `historico/` con fecha explícita.

## 12. Revisión periódica

Antes de una entrega:

1. ejecutar pruebas;
2. actualizar estado ejecutivo;
3. revisar requisitos y backlog;
4. comprobar links;
5. revisar secretos;
6. generar acta de evidencia;
7. confirmar que los documentos históricos no aparecen como vigentes.
