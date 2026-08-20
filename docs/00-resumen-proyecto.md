# Poli-REDI - Resumen vigente para compartir

Fecha de corte: 2026-08-20

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
| Autenticacion y sesion | IMPLEMENTADO | Microsoft Entra ID y modo local de desarrollo. El frontend diferencia sesion no resuelta, autenticada y anonima, y utiliza una pantalla intermedia compartida para entrada/cierre sin mostrar prematuramente el login. El ambiente online no fue reverificado en este corte. |
| Identidad, roles y bloqueo | IMPLEMENTADO | El servidor obtiene al usuario desde la identidad validada, consulta rol/bloqueo local y protege rutas administrativas. |
| Perfil y RUT | IMPLEMENTADO | Usuario normal debe registrar RUT para reservar o inscribirse en talleres. |
| Recursos | IMPLEMENTADO PARCIAL | Catalogo y cambio administrativo de imagen; no existe gestion completa de altas, datos, modos y activacion. |
| Disponibilidad | IMPLEMENTADO PARCIAL | Integra reservas, actividades programadas y talleres; el frontend permite filtrar por recurso y por existencia de bloques disponibles. El recurso puede precargarse desde Inicio mediante query string. Permanecen pendientes las brechas de backend/rango que correspondan. |
| Creacion de reservas | IMPLEMENTADO PARCIAL | Propietario, zona horaria, jornada y duraciones son controlados por servidor. Los recursos grupales pueden comenzar `PENDING`, registran al solicitante y alcanzan `CONFIRMED` segun participantes; quedan integraciones de cierre por completar. |
| Reglas institucionales | APROBADO / IMPLEMENTACION PARCIAL | Ventana, frecuencia, versionado y flujo grupal tienen implementacion incremental. Participantes, `PENDING -> CONFIRMED` y `AT_RISK` existen; vencimiento, notificaciones, interfaz administrativa y correcciones excepcionales requieren cierre adicional. |
| Cancelacion | IMPLEMENTADO | Propietario o administrador pueden cancelar estados activos no finalizados. El frontend utiliza confirmacion destructiva inline y evita `window.confirm`; el cambio de estado visible actua como confirmacion de la operacion. |
| Talleres | IMPLEMENTADO | Consulta e inscripcion con RUT, cupo y duplicado controlados; no existe desinscripcion. |
| Notificaciones | IMPLEMENTADO PARCIAL | Consulta y contador existen; no se marcan como leidas y la generacion cubre solo eventos limitados. |
| Administracion | IMPLEMENTADO PARCIAL | Panel, lectura de usuarios, recursos, reservas e indicadores; faltan gestion del inventario oficial, usuarios, bloqueos, programacion, conflictos institucionales e infracciones. |
| Reportes | IMPLEMENTADO PARCIAL | Indicadores calculados en frontend; no constituyen reportes institucionales completos ni consumen las vistas SQL dedicadas. |
| Auditoria | IMPLEMENTADO PARCIAL | El esquema registra cambios de reservas, pero no existe consulta administrativa. |
| Calidad local | VERIFICADO PARCIAL | `npm run build`, `npm test` y `git diff --check` aprobaron el 2026-08-20; la suite frontend ejecuto 25 pruebas correctamente. La validacion integrada con infraestructura, navegador y ambiente online sigue siendo una evidencia separada. |
| Despliegue | IMPLEMENTADO SEGUN REPOSITORIO | Existen configuracion y documentacion de demo Azure; su disponibilidad actual no fue verificada en este corte. |

## Criterio de evolucion del producto

Los requisitos de Poli-REDI son versionables entre MVPs. Una regla puede refinarse cuando el incremento implementado entrega evidencia de que otro modelo representa mejor el caso de uso.

El requisito anterior permanece como trazabilidad historica; la decision refinada pasa a ser la regla vigente y debe propagarse a requisitos, backlog, flujos, pruebas y siguientes MVP.

Este criterio permite distinguir entre:

- defecto: la implementacion incumple una regla vigente;
- refinamiento: la regla vigente cambia deliberadamente por aprendizaje del MVP.

La adopcion de `CONFIRMED + AT_RISK` para reservas que ya alcanzaron el minimo es un refinamiento del segundo tipo.

## Evaluacion por incremento

| Incremento | Estado recomendado | Lectura de producto |
| --- | --- | --- |
| MVP 1 | Demo funcional, cierre pendiente | La base opera y tiene pruebas locales, pero requiere validacion integrada/online, seguridad de errores y cierre de accesibilidad/responsive. |
| MVP 2 | Avanzado; UX principal unificado | El flujo de usuario visible esta consolidado: Inicio, Disponibilidad, Reservas/Historial y detalle comparten patrones. Persisten brechas de reglas grupales, validacion integral, infraestructura y pruebas de integracion. |
| MVP 3 | Parcial | Hay lectura y resumen administrativo, no gestion institucional completa. |
| MVP 4 | En desarrollo | Documentacion y pruebas iniciales existen; reportes, notificaciones, infracciones y evidencia final siguen incompletos. |

## Brechas y contradicciones que impiden declarar cierre

1. El alcance academico definitivo excluye autenticacion institucional real y despliegue productivo, pero el repositorio documenta Entra ID y una demo Azure ya implementados.
2. La ventana y frecuencia versionadas requieren conservar evidencia de integracion y verificacion en la infraestructura objetivo.
3. El flujo grupal ya soporta `PENDING`, participantes, codigo de invitacion, confirmacion por minimo y condicion `AT_RISK`. La regla fue refinada durante MVP 2: una reserva ya confirmada no regresa a `PENDING` al caer bajo el minimo; conserva `CONFIRMED` y cambia su condicion grupal. El cierre pendiente se concentra en vencimiento, liberacion asociada e integracion con notificaciones.
4. Ante actividad institucional versus reserva particular, la reserva debe cancelarse automaticamente y notificarse al usuario; ante dos actividades, el administrador debe poder cancelar una o mantener ambas. El esquema actual rechaza esos conflictos.
5. Los ocho recursos del seed representan el inventario oficial, pero el administrador aun no puede mantenerlo de forma completa.

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

1. Con la configuracion vigente, un usuario solo puede elegir fechas desde el dia actual hasta el dia anterior al mismo dia de la semana siguiente; por ejemplo, un martes puede reservar hasta el lunes siguiente.
2. Al crear una solicitud, el usuario no puede crear otra hasta el mismo dia de la semana siguiente; por ejemplo, si la crea un martes para el miercoles, vuelve a poder solicitar desde el martes siguiente. La duracion de este periodo debe ser configurable.
3. Una solicitud `PENDING` consume la oportunidad desde su creacion; al pasar a `CANCELLED` deja de consumirla.
4. El minimo de 10 participantes es obligatorio para Cancha 1, Cancha 2 y Cancha 3, que corresponden formalmente a multicancha 1, 2 y 3. El solicitante cuenta y todos los participantes deben tener cuenta.
5. La solicitud `PENDING` bloquea el horario. Las confirmaciones pueden registrarse o retirarse hasta exactamente una hora antes del inicio, inclusive, con plazo configurable. Si una retirada deja menos de 10, vuelve a `PENDING`; si llega al limite bajo el minimo, cambia a `CANCELLED` y libera la oportunidad semanal.
6. Para todos los recursos se aprueban duraciones de 30, 60, 90, 120, 150 y 180 minutos.
7. `OPEN_USE` no requiere confirmacion de integrantes; los recursos grupales indicados se confirman automaticamente al alcanzar el minimo.
8. Un conflicto entre actividad institucional y reserva particular cancela automaticamente la reserva particular y debe notificar al usuario afectado.
9. Un conflicto entre dos actividades debe informarse al administrador, quien puede cancelar una de ellas o mantener ambas.
10. Los ocho recursos actuales constituyen el inventario oficial. Solo usuarios con rol administrador pueden modificar recursos, periodos, plazos o recursos sujetos a la politica.

11. Los cambios administrativos de politicas se aplican prospectivamente: cada solicitud conserva la version vigente al crearse.
12. Excepcionalmente, un administrador puede corregir solicitudes futuras `PENDING` o `CONFIRMED` mediante seleccion explicita, simulacion previa, motivo obligatorio, aplicacion atomica y auditoria. La correccion no edita versiones historicas ni cancela solicitudes implicitamente.
13. El solicitante cuenta como participante y no puede retirar su participacion; para salir debe cancelar la solicitud completa.

## Arquitectura de politicas: estado de implementacion

La politica se versiona y cada solicitud referencia la version aplicable. Publicacion inmediata, snapshot completo, permisos, historial e idempotencia estan implementados y verificados localmente. Participantes/estados, plazo/vencimiento, interfaz administrativa y correcciones excepcionales siguen pendientes. `ADMIN-005` se mantiene para una entrega arquitectonica posterior.

## Evidencia local del corte

Evidencia mas reciente del frontend, 2026-08-20:

- `npm run build`: aprobado.
- `npm test`: 25 pruebas aprobadas.
- `git diff --check`: aprobado.
- Flujo revisado manualmente durante el pulido UX: carrusel manual, filtros de Disponibilidad, detalle compartido, cancelacion con confirmacion inline y transiciones de autenticacion.
- Commit funcional de referencia: `a9a599d` (`feat: unify reservation flows and auth transitions`).

Evidencia historica del backend y de infraestructura debe mantenerse separada y no inferirse a partir de estas pruebas frontend.

No verificado nuevamente en este corte documental: infraestructura Azure en ejecucion, Microsoft Entra ID real, responsive exhaustivo, accesibilidad completa y demo online.
