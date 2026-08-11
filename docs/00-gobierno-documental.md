# Gobierno documental de Poli-REDI

**Audiencia:** todos los roles del proyecto.

**Propósito:** establecer autoridad, ubicación y mantenimiento de la documentación.

**Estado:** CANÓNICO desde 2026-08-11.

**Fuente propietaria:** este archivo es la única norma de gobierno documental.

## Resumen

La documentación vigente reside en la raíz de `docs/`, con decisiones inmutables y anexos separados. Los árboles `docs/Propuesta nueva/` y `docs/historico_y_checklists/` se conservan temporalmente como fuentes de migración, pero dejan de ser autoridad para contenido nuevo.

## Propietario por tema

| Tema | Fuente propietaria |
|---|---|
| Gobierno y precedencia | `docs/00-gobierno-documental.md` |
| Estado vigente | `docs/01-estado-actual.md` |
| Arquitectura y contratos | `docs/02-arquitectura-y-contratos.md` |
| Requisitos y trazabilidad | `docs/03-requisitos-y-trazabilidad.md` |
| Datos y migraciones | `docs/04-base-de-datos-y-migraciones.md` |
| Instalación y operación | `docs/05-instalacion-despliegue-recuperacion.md` |
| Reglas y flujos | `docs/06-reglas-y-flujos.md` |
| Calidad y evidencia viva | `docs/07-calidad-y-evidencia.md` |
| Plan, backlog y fechas | `docs/08-plan-entrega-2026.md` |
| Checklist de cierre | `docs/09-checklist-cierre.md` |
| Decisiones duraderas | `docs/decisiones/ADR-*.md` |
| Evidencia fechada | `docs/anexos/evidencia/` |
| Diagramas | `docs/anexos/diagramas/` |
| Operaciones excepcionales | `docs/anexos/operacion/` |

## Precedencia

1. Un ADR aprobado gobierna la decisión que registra.
2. El documento propietario gobierna el contenido vivo de su tema.
3. Código, esquema y configuración demuestran implementación observable.
4. Pruebas y auditorías fechadas demuestran únicamente la versión y ambiente indicados.
5. Documentos académicos e históricos aportan contexto, no sustituyen el estado vigente.

Si dos fuentes difieren, no se combinan por intuición: se identifica el propietario, se registra la discrepancia y se actualizan las fuentes derivadas.

## Estados permitidos

| Estado | Significado |
|---|---|
| Aprobado | Existe decisión de alcance vigente. |
| Implementado | Existe comportamiento observable en código o datos. |
| Verificado localmente | Hay ejecución local fechada. |
| Validado integradamente | Frontend, API, identidad y datos se probaron juntos. |
| Desplegado | Hay evidencia del artefacto o migración en el ambiente indicado. |
| Parcial | Falta parte relevante del requisito o su evidencia. |
| Pendiente | No existe implementación o evidencia suficiente. |
| Fuera de alcance | Una decisión vigente excluye la capacidad. |

## Árbol vigente

```text
README.md
CONTRIBUTING.md
docs/
  00-gobierno-documental.md
  01-estado-actual.md
  02-arquitectura-y-contratos.md
  03-requisitos-y-trazabilidad.md
  04-base-de-datos-y-migraciones.md
  05-instalacion-despliegue-recuperacion.md
  06-reglas-y-flujos.md
  07-calidad-y-evidencia.md
  08-plan-entrega-2026.md
  09-checklist-cierre.md
  decisiones/
  anexos/
```

## Recorridos recomendados

| Audiencia | Recorrido |
|---|---|
| Dirección | `01` → `08` → `09` |
| Análisis y UX | `03` → `06` → `09` |
| Desarrollo | `02` → `03` → `04` → `06` → `07` |
| Operación | `05` → `08` → `09` |
| Tesis | `01` → `03` → auditoría → diagramas |

## Regla de actualización

| Cambio | Actualizar |
|---|---|
| Decisión de alcance | ADR, `03`, `01` y `08` |
| Contrato técnico | `02`, `04` o `06`; pruebas en `07` |
| Estado o evidencia | `01`, `07` y anexo fechado |
| Prioridad o fecha | Solo `08`; `09` si cambia un gate |
| Procedimiento operativo | Solo `05` o anexo operativo |

Los ADR no se reescriben para cambiar una decisión: se crea otro que los sustituya. La evidencia fechada tampoco se actualiza retrospectivamente.

## Fuentes académicas

`../Documentos/01` a `04` permanecen como documentos académicos. Deben enlazar este árbol y declarar su carácter derivado cuando resuman estado, requisitos o fechas.

Antes de retirar fuentes anteriores se deben actualizar enlaces entrantes, comprobar UTF-8, validar Markdown y conservar un punto recuperable en Git.
