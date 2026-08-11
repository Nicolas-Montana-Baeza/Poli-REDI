# ADR-001 — Gobierno documental

**Audiencia:** todo rol que crea, modifica o aprueba documentación

**Propósito:** establecer una sola fuente vigente por tema

**Estado:** aceptada el 2026-08-11

**Fuente:** revisión documental integral y decisión del orquestador

## Resumen

Se adopta el árbol compacto iniciado en `docs/00-gobierno-documental.md` como canon operativo. Las fuentes anteriores se preservan durante la transición, pero no compiten con el canon ni se reescriben para aparentar estado vigente.

## Contexto

El repositorio acumuló documentos raíz, `docs/Propuesta nueva/`, referencias y actas históricas con contenido solapado. La repetición dificultaba identificar propietario, estado y evidencia real.

## Decisión

1. Cada tema tiene un documento propietario definido en [00-gobierno-documental.md](../00-gobierno-documental.md).
2. Los documentos nuevos contienen cabecera de audiencia, propósito, estado y fuente.
3. Los ejecutivos resumen y enlazan; no duplican contratos, reglas ni checklists.
4. Los históricos y fuentes supersedidas se preservan hasta una fase de archivo explícita.
5. Las correcciones de decisiones pasadas se registran mediante ADR, acta o anotación, no alterando el original.
6. La documentación académica se subordina al canon técnico para el estado del producto, conservando su función de tesis.

## Consecuencias

- Mejora navegación, propiedad y legibilidad.
- Los enlaces entrantes deben migrarse gradualmente.
- Durante la transición coexistirán fuentes antiguas, rotuladas como referencia o histórico.
- Eliminar o mover fuentes requiere inventario de enlaces, revisión de binarios y autorización separada.

## Superación

Solo otro ADR aceptado puede cambiar este modelo. Debe incluir manifiesto de migración, enlaces afectados y tratamiento de históricos.
