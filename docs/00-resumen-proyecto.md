# Poli-REDI - Resumen vigente para compartir

Fecha de corte: 2026-07-20

## Proposito

Este documento es el punto de entrada para revisar Poli-REDI sin asumir que todo lo documentado esta implementado ni que todo lo implementado forma parte del alcance aprobado. Resume el problema, el estado observable y los documentos que deben acompañar una revision de producto.

Para el analisis detallado, criterios y contradicciones usar `docs/13-estado-actual-producto.md`.

## Problema y objetivo

El proceso levantado depende de gestion manual y Google Calendar, concentra decisiones en el encargado y no ofrece una vista institucional unica de disponibilidad, reglas ni trazabilidad.

El objetivo del prototipo es validar que una aplicacion web puede centralizar la consulta de espacios deportivos, las reservas particulares y la informacion administrativa basica, respetando las reglas que la institucion confirme.

## Convenciones de estado

- `APROBADO`: definido por una decision explicita o un documento de alcance confirmado.
- `IMPLEMENTADO`: existe comportamiento observable en el codigo o esquema.
- `VERIFICADO`: existe una comprobacion ejecutada con resultado satisfactorio.
- `PROPUESTO`: requiere aprobacion antes de convertirse en alcance.
- `PENDIENTE`: no esta resuelto o carece de validacion suficiente.
- `DESCARTADO`: excluido de manera explicita.

Un elemento puede estar implementado sin estar aprobado y puede estar implementado sin haber sido verificado en un ambiente integrado.

## Estado ejecutivo

| Area | Estado | Evidencia y limite principal |
| --- | --- | --- |
| Autenticacion y sesion | IMPLEMENTADO | Microsoft Entra ID y modo local exclusivo de desarrollo; no se verifico el ambiente online en este corte. |
| Identidad, roles y bloqueo | IMPLEMENTADO | El servidor obtiene al usuario desde la identidad validada, consulta rol/bloqueo local y protege rutas administrativas. |
| Perfil y RUT | IMPLEMENTADO | Usuario normal debe registrar RUT para reservar o inscribirse en talleres. |
| Recursos | IMPLEMENTADO PARCIAL | Catalogo y cambio administrativo de imagen; no existe gestion completa de altas, datos, modos y activacion. |
| Disponibilidad | IMPLEMENTADO PARCIAL | Integra reservas y actividades programadas; la interfaz agrega talleres recurrentes. Los bloqueos no se muestran y no hay filtros de rango en el servidor. |
| Creacion de reservas | IMPLEMENTADO PARCIAL | Propietario, zona horaria, jornada y duraciones son controlados por servidor; falta aplicar la confirmacion condicional aprobada segun tipo de recurso. |
| Reglas institucionales | APROBADO / PENDIENTE DE IMPLEMENTACION | La espera semanal y el minimo de 10 participantes fueron confirmados el 2026-07-20. La duracion puede variar; falta precisar si se mantiene exactamente el rango actual de 30 a 180 minutos. |
| Cancelacion | IMPLEMENTADO PARCIAL | Propietario o administrador pueden cancelar estados activos no finalizados; la confirmacion visible no es consistente en todos los accesos. |
| Talleres | IMPLEMENTADO | Consulta e inscripcion con RUT, cupo y duplicado controlados; no existe desinscripcion. |
| Notificaciones | IMPLEMENTADO PARCIAL | Consulta y contador existen; no se marcan como leidas y la generacion cubre solo eventos limitados. |
| Administracion | IMPLEMENTADO PARCIAL | Panel, lectura de usuarios, recursos, reservas e indicadores; faltan gestion del inventario oficial, usuarios, bloqueos, programacion, conflictos institucionales e infracciones. |
| Reportes | IMPLEMENTADO PARCIAL | Indicadores calculados en frontend; no constituyen reportes institucionales completos ni consumen las vistas SQL dedicadas. |
| Auditoria | IMPLEMENTADO PARCIAL | El esquema registra cambios de reservas, pero no existe consulta administrativa. |
| Calidad local | VERIFICADO PARCIAL | `go test ./...`, `npm test` y `npm run build` aprobaron el 2026-07-20. No se ejecutaron pruebas integradas con Azure SQL, navegador ni ambiente online. |
| Despliegue | IMPLEMENTADO SEGUN REPOSITORIO | Existen configuracion y documentacion de demo Azure; su disponibilidad actual no fue verificada en este corte. |

## Evaluacion por incremento

| Incremento | Estado recomendado | Lectura de producto |
| --- | --- | --- |
| MVP 1 | Demo funcional, cierre pendiente | La base opera y tiene pruebas locales, pero requiere validacion integrada/online, seguridad de errores y cierre de accesibilidad/responsive. |
| MVP 2 | Avanzado con brechas obligatorias | El flujo de usuario existe, pero faltan la restriccion semanal, confirmaciones de participantes y estados condicionales ya aprobados. |
| MVP 3 | Parcial | Hay lectura y resumen administrativo, no gestion institucional completa. |
| MVP 4 | En desarrollo | Documentacion y pruebas iniciales existen; reportes, notificaciones, infracciones y evidencia final siguen incompletos. |

## Brechas y contradicciones que impiden declarar cierre

1. El alcance academico definitivo excluye autenticacion institucional real y despliegue productivo, pero el repositorio documenta Entra ID y una demo Azure ya implementados.
2. Se aprobo que la duracion no tiene que ser exactamente una hora, pero falta confirmar si el catalogo actual de 30 a 180 minutos es definitivo o si depende del recurso.
3. La espera semanal y el minimo de 10 participantes son obligatorios, pero el flujo actual no los solicita ni valida.
4. El estado debe depender del recurso: `OPEN_USE` no requiere confirmacion de participantes; recursos grupales como multicancha deben quedar pendientes hasta alcanzar el minimo. El flujo actual confirma todas las reservas inmediatamente.
5. Ante actividad institucional versus reserva particular, la reserva debe cancelarse automaticamente; ante dos actividades, el administrador debe poder cancelar una o mantener ambas. El esquema actual rechaza esos conflictos.
6. Los ocho recursos del seed representan el inventario oficial, pero el administrador aun no puede mantenerlo de forma completa.

Estas diferencias no se resuelven en este documento. Requieren decision del Orquestador y, cuando corresponda, confirmacion institucional.

## Documentos iniciales para compartir

Compartir en este orden:

1. `docs/00-resumen-proyecto.md`: entrada ejecutiva, estado y contradicciones.
2. `docs/13-estado-actual-producto.md`: analisis completo de producto, requisitos, reglas, criterios y trazabilidad.
3. `docs/08-requisitos-historias-casos-uso.md`: catalogo funcional vigente del repositorio.
4. `docs/09-mvps-roadmap.md`: agrupacion incremental y pendientes.
5. `docs/12-checklist-demo-mvp1.md`: evidencia automatizada y validaciones manuales pendientes.

Agregar segun el destinatario:

- `docs/06-flujo-reservas.md` para revisar el flujo funcional.
- `docs/02-arquitectura.md` y `docs/03-base-de-datos.md` para Arquitectura o Datos.
- `docs/10-guia-redeploy.md` para operacion o despliegue.
- `../Documentos/alcance_definitivo_prototipo_poli_redi.txt` y `../Documentos/levantamiento_poli_redi.txt`, relativos a la raiz del proyecto `Poli-REDI`, para resolver alcance y reglas originales.

No usar `docs/00-revision-inicial.md` como estado vigente; es un registro historico.

## Decisiones aprobadas el 2026-07-20

1. La espera semanal y el minimo de 10 participantes son obligatorios.
2. La duracion de una reserva no tiene que ser necesariamente una hora.
3. `OPEN_USE` no requiere confirmacion de integrantes; recursos grupales como multicancha se confirman automaticamente al alcanzar el minimo de integrantes confirmados.
4. Un conflicto entre actividad institucional y reserva particular cancela automaticamente la reserva particular.
5. Un conflicto entre dos actividades debe informarse al administrador, quien puede cancelar una de ellas o mantener ambas.
6. Los ocho recursos actuales constituyen el inventario oficial y deben ser administrables en un MVP posterior.

## Precisiones pendientes

1. Definir el limite exacto de la semana y que estados de reserva cuentan para la restriccion.
2. Definir si el solicitante cuenta entre los 10 integrantes y el plazo para reunir confirmaciones.
3. Identificar en el inventario cuales recursos, ademas de multicancha, requieren confirmacion grupal.
4. Confirmar el catalogo exacto de duraciones por recurso.
5. Definir la notificacion al usuario cuya reserva particular sea cancelada automaticamente.

## Evidencia local del corte

- Backend: `go test ./...` aprobado; paquetes con pruebas de reloj, JSON, reglas horarias y servicio de reservas.
- Frontend: `npm test` aprobado; 9 pruebas de zona horaria y reglas de agenda.
- Frontend: `npm run build` aprobado; build de produccion generado sin error.
- No verificado en este corte: Azure SQL en ejecucion, navegacion manual, accesibilidad, responsive, Microsoft Entra ID real y demo online.
