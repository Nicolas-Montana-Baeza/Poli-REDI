# Poli-REDI - Estado actual de producto

Fecha de corte: 2026-08-20

### 1. Resumen

- **Problema — HECHO:** el proceso levantado es manual, depende de Google Calendar, concentra la operacion en el encargado y carece de disponibilidad y trazabilidad centralizadas.
- **Objetivo — APROBADO:** validar un prototipo que centralice recursos, disponibilidad, reservas y reglas institucionales basicas sin presentarlo como plataforma productiva completa.
- **Usuarios afectados — HECHO:** usuarios normales, administradores y personal institucional encargado de la operacion deportiva.
- **Resultado esperado — PROPUESTA:** disponer de una demo verificable cuyo alcance implementado, faltante y futuro pueda comunicarse sin confundir codigo, aprobacion y validacion.
- **Estado actual — HECHO:** existe una demo funcional con identidad, reservas y consulta operacional. El flujo grupal ya crea solicitudes `PENDING`, persiste participantes autenticados, confirma al alcanzar el minimo y representa riesgo posterior mediante `AT_RISK`. Permanecen pendientes el cierre del vencimiento automatico, algunas integraciones administrativas, notificaciones y evidencia integral antes de declarar cerrado MVP 2.

### 2. Evidencia revisada

| Fuente | Hallazgo | Clasificacion |
| --- | --- | --- |
| `../Documentos/levantamiento_poli_redi.txt` | Gestion manual; jornada 08:00-22:00; reserva aproximada de una hora; espera semanal; minimo de 10 participantes; prioridad institucional y decisiones administrativas. | HECHO documental; requiere confirmar vigencia institucional |
| `../Documentos/alcance_definitivo_prototipo_poli_redi.txt` | MVP 1 y 2 son el nucleo; participantes, restriccion semanal y bloqueos forman parte del alcance; Entra real y despliegue productivo aparecen fuera de alcance. | APROBADO segun nombre del documento, pero contradicho por documentacion posterior |
| `docs/08-requisitos-historias-casos-uso.md` | Consolida 25 requisitos funcionales, incluidos los acuerdos del 2026-07-20 y el estado parcial verificado localmente de las politicas prospectivas. | ACTUALIZADO |
| `backend/internal/routes/routes.go` | La superficie comun incluye identidad, recursos, disponibilidad y reservas. MVP2 agrega reservas grupales, talleres institucionales, unidades, programacion institucional y resolucion administrativa de conflictos. `FULL` conserva modulos legacy como notificaciones y administracion de politicas. | IMPLEMENTADO por scopes; cierre funcional variable |
| `backend/internal/middleware/auth_middleware.go` | La identidad se valida en servidor; rol, bloqueo y RUT provienen del usuario local; `DEV_AUTH_ENABLED` selecciona el modo local. | IMPLEMENTADO |
| `backend/internal/services/reservations_service.go` y flujo de participantes | La reserva usa el usuario autenticado y aplica estado inicial segun politica. Los recursos grupales pueden comenzar `PENDING`, registrar participantes y alcanzar `CONFIRMED`; la condicion grupal se mantiene separada del ciclo de vida. | IMPLEMENTADO PARCIAL y con pruebas locales/integracion disponibles |
| `database/postgres/migrations/` | PostgreSQL 16 es la linea canonica vigente. `PG16_0001` a `0008` cubren baseline, invariantes, participantes, programacion institucional, disponibilidad y excepciones. | IMPLEMENTADO incremental; provisioning automatico MVP2 PENDIENTE |
| `frontend/src/components/availability/AvailabilitySection.vue` | La interfaz combina reservas, actividades programadas y talleres; no incorpora bloqueos visibles y permite abrir seleccion sobre recursos que despues puede rechazar el servidor. | IMPLEMENTADO PARCIAL |
| `database/seed.sql` | Contiene los ocho recursos confirmados como inventario oficial inicial. | APROBADO como linea base; gestion completa PENDIENTE |
| Decisiones explicitas del usuario del 2026-07-20 | Confirman que `PENDING` consume y `CANCELLED` libera la oportunidad; el solicitante cuenta y todos los participantes tienen cuenta; la solicitud bloquea, admite cambios en el limite exacto y se cancela al vencer bajo el minimo; las tres canchas son multicanchas y solo administradores modifican politicas. | APROBADO |
| Pruebas locales 2026-07-21 | Evidencia historica de la etapa Azure SQL: cuatro rondas QA, pruebas Go y validaciones de contrato; `npm test` y `npm run build` aprobaron. | EVIDENCIA HISTORICA; no define el motor vigente |
| `docs/12-checklist-demo-mvp1.md` | La mayor parte de la prueba manual e integrada sigue pendiente; no hay verificacion online actual registrada. | PENDIENTE |

### 3. Alcance

#### Dentro del alcance

- Mantener una descripcion verificable del comportamiento actual.
- Identidad institucional y modo de desarrollo, roles, bloqueo y RUT.
- Consulta de recursos, actividades, disponibilidad, reservas propias/globales e historial.
- Creacion y cancelacion de reservas conforme a las reglas actualmente implementadas.
- Ventana y frecuencia de reserva configurables, confirmacion de 10 participantes, plazo previo y estado condicionado por recurso.
- Prioridad institucional y resolucion administrativa de conflictos entre actividades.
- Inventario oficial inicial de ocho recursos y su gestion administrativa futura.
- Consulta e inscripcion en talleres.
- Panel y lectura administrativa basica.
- Notificaciones e indicadores en su nivel parcial actual.
- Documentacion del desfase entre alcance academico y prototipo implementado.

#### Fuera del alcance

- Diseñar o implementar cambios de frontend, backend, base de datos o despliegue.
- Declarar aprobadas las reglas que solo aparecen implementadas.
- Declarar el sistema listo para operacion institucional o produccion.
- Implementar la decision aprobada sobre versiones prospectivas y correcciones retroactivas excepcionales controladas.
- Certificar el ambiente online sin una ejecucion verificable.

#### Futuro posible

- Gestion administrativa completa de usuarios, inventario oficial, bloqueos y programacion.
- Cierre de vencimiento, capacidad e integraciones del flujo grupal de punta a punta como trabajo obligatorio de MVP 2.
- Infracciones, notificaciones completas, reportes institucionales y consulta de auditoria.
- Filtros de servidor, detalle individual de reserva y disponibilidad por rango.
- Cierre de accesibilidad, responsive, seguridad de errores y pruebas integradas.

### 4. Actores

| Actor | Necesidad | Permisos o restricciones |
| --- | --- | --- |
| Usuario normal | Consultar disponibilidad, reservar, cancelar lo propio, revisar historial e inscribirse en talleres. | Debe autenticarse; el servidor determina su identidad; necesita RUT; no accede a datos personales ajenos ni acciones administrativas. |
| Participante | Confirmar o retirar participacion en una solicitud grupal. | Debe tener cuenta autenticada; el solicitante cuenta una vez; no puede duplicar su confirmacion. |
| Administrador | Supervisar reservas, mantener inventario y politicas, programar actividades y resolver conflictos. | Debe tener rol validado por el servidor; es el unico actor que puede modificar periodos, plazos o recursos sujetos a confirmacion. |
| Personal institucional | Mantener programacion, bloqueos y criterios operativos. | Actor levantado, pero sin flujo propio completo en el sistema actual. |
| Microsoft Entra ID | Acreditar identidad institucional. | Entrega identidad; Poli-REDI conserva el control de roles, bloqueo y RUT. |
| Soporte u operacion | Preparar y verificar ambientes de demo. | Debe impedir `DEV_AUTH_ENABLED` en el ambiente publico y no exponer secretos. |

### 5. Requisitos

#### Requisitos funcionales

**ID:** RF-001  
**Titulo:** Acceso autenticado  
**Descripcion:** El sistema debe permitir acceso mediante identidad Microsoft y un modo local explicitamente limitado a desarrollo.  
**Actor:** Usuario normal o administrador.  
**Precondiciones:** Configuracion de identidad disponible.  
**Comportamiento esperado:** Validar la identidad antes de habilitar funciones internas.  
**Resultado:** Sesion asociada a un usuario local activo o rechazo seguro.  
**Prioridad:** MUST.  
**Fuente:** `docs/08-requisitos-historias-casos-uso.md`, middleware y rutas.  
**Estado:** IMPLEMENTADO; verificacion online PENDIENTE.  
**Dependencias:** Microsoft Entra ID o modo local de desarrollo.

**ID:** RF-002  
**Titulo:** Perfil local y autorizacion  
**Descripcion:** El sistema debe obtener o crear el perfil asociado a la identidad validada y aplicar rol, bloqueo y RUT desde datos controlados por el servidor.  
**Actor:** Sistema de identidad y usuario autenticado.  
**Precondiciones:** Identidad valida.  
**Comportamiento esperado:** Sincronizar identidad y rechazar cuentas bloqueadas.  
**Resultado:** Perfil local vigente disponible para autorizar operaciones.  
**Prioridad:** MUST.  
**Fuente:** `AUTH-002`, middleware y repositorio de usuarios.  
**Estado:** IMPLEMENTADO.  
**Dependencias:** RF-001.

**ID:** RF-003  
**Titulo:** Consulta de disponibilidad util  
**Descripcion:** El sistema debe mostrar ocupacion por recurso y fecha considerando todas las fuentes que realmente impiden reservar, sin exponer datos personales ajenos.  
**Actor:** Usuario normal y administrador.  
**Precondiciones:** Sesion activa y datos cargables.  
**Comportamiento esperado:** Mostrar reservas, programacion, talleres y bloqueos aplicables con comportamiento diferenciado por modo de recurso.  
**Resultado:** El usuario puede distinguir un horario reservable antes de enviar una solicitud.  
**Prioridad:** MUST.  
**Fuente:** RF-005, alcance definitivo y flujo actual.  
**Estado:** IMPLEMENTADO PARCIAL; bloqueos visibles y filtros de rango PENDIENTES.  
**Dependencias:** Catalogo de recursos y fuentes de ocupacion.

**ID:** RF-004  
**Titulo:** Crear reserva propia  
**Descripcion:** El sistema debe crear una reserva para el usuario autenticado sobre un recurso elegible y un horario permitido.  
**Actor:** Usuario normal o administrador.  
**Precondiciones:** Usuario activo; RUT para usuario normal; recurso existente.  
**Comportamiento esperado:** Ignorar identidad y estado decididos por el cliente, validar fecha, jornada, duracion y conflictos, y asignar el estado inicial.  
**Resultado:** Solicitud creada con el estado asignado por su politica (`PENDING` o `CONFIRMED`) o rechazo explicativo sin efectos parciales.<br>
**Prioridad:** MUST.  
**Fuente:** RF-006, RF-007, codigo y esquema.  
**Estado:** IMPLEMENTADO para las reglas actuales; alcance institucional en contradiccion.  
**Dependencias:** RF-001, RF-002, RF-003, RF-020, RF-021 y RF-022.

**ID:** RF-005  
**Titulo:** Cancelar reserva permitida  
**Descripcion:** El propietario o un administrador debe poder cancelar una reserva activa que aun no haya finalizado.  
**Actor:** Usuario normal o administrador.  
**Precondiciones:** Reserva existente y sesion autorizada.  
**Comportamiento esperado:** Confirmar intencion, validar propiedad/rol, estado y termino institucional.  
**Resultado:** Estado `CANCELLED` o rechazo sin modificar la reserva.  
**Prioridad:** MUST.  
**Fuente:** RF-008 y flujo actual.  
**Estado:** IMPLEMENTADO; la confirmacion destructiva inline se utiliza en el flujo compartido. Verificacion integrada/online PENDIENTE como evidencia separada.<br>
**Dependencias:** RF-001, RF-002 y RN-008.

**ID:** RF-006  
**Titulo:** Inscribirse en taller  
**Descripcion:** Un usuario autenticado con RUT debe poder inscribirse una sola vez en un taller activo con cupo.  
**Actor:** Usuario normal.  
**Precondiciones:** Taller activo, cupo y RUT.  
**Comportamiento esperado:** Validar duplicado y capacidad antes de confirmar.  
**Resultado:** Inscripcion confirmada o rechazo recuperable.  
**Prioridad:** SHOULD.  
**Fuente:** RF-019 y codigo.  
**Estado:** IMPLEMENTADO.  
**Dependencias:** RF-002.

**ID:** RF-007  
**Titulo:** Administracion basica  
**Descripcion:** El administrador debe poder consultar usuarios, recursos, reservas e indicadores sin exponer estas vistas a usuarios normales.  
**Actor:** Administrador.  
**Precondiciones:** Rol administrador vigente.  
**Comportamiento esperado:** Autorizar en servidor y mostrar datos operacionales.  
**Resultado:** Lectura administrativa disponible.  
**Prioridad:** SHOULD.  
**Fuente:** RF-012 a RF-014 y RF-018.  
**Estado:** IMPLEMENTADO PARCIAL; gestion de datos PENDIENTE.  
**Dependencias:** RF-002.

**ID:** RF-020  
**Titulo:** Restriccion semanal de reserva particular  
**Descripcion:** El sistema debe limitar las fechas reservables al periodo institucional configurado e impedir mas de una solicitud vigente del mismo usuario dentro del periodo contado desde la fecha local de creacion de la solicitud anterior. `PENDING` y `CONFIRMED` consumen la oportunidad; `CANCELLED` deja de consumirla.<br>
**Actor:** Usuario normal.  
**Precondiciones:** Usuario identificado e historial consultable.  
**Comportamiento esperado:** Evaluar la fecha institucional, la configuracion vigente y el historial aplicable; rechazar fechas fuera de ventana o solicitudes prematuras y comunicar la proxima fecha permitida.<br>
**Resultado:** Reserva aceptada o rechazada de forma verificable.  
**Prioridad:** MUST.  
**Fuente:** Alcance definitivo, levantamiento y decision explicita del usuario del 2026-07-20.  
**Estado:** APROBADO e IMPLEMENTADO. Existe evidencia local y PostgreSQL vigente; falta cierre integral en el ambiente objetivo.<br>
**Dependencias:** RF-025 para modificar el periodo institucional.

**ID:** RF-021  
**Titulo:** Confirmacion de participantes minimos  
**Descripcion:** Para multicancha 1, 2 y 3, registradas como Cancha 1, 2 y 3, el sistema debe exigir al menos 10 usuarios unicos con cuenta, incluido el solicitante, y aceptar cambios hasta exactamente una hora antes inclusive, plazo configurable.<br>
**Actor:** Solicitante y participantes.  
**Precondiciones:** Reserva grupal identificable y participantes autenticados.<br>
**Comportamiento esperado:** Contabilizar confirmaciones validas sin duplicados, respetar capacidad y comparar con el minimo.<br>
**Resultado:** Antes del primer cumplimiento del minimo permanece `PENDING`; al alcanzar 10 pasa a `CONFIRMED`; una perdida posterior del minimo se representa mediante `AT_RISK` sin borrar la confirmacion alcanzada.<br>
**Prioridad:** MUST.  
**Fuente:** Alcance definitivo, levantamiento, decision del 2026-07-20 y refinamiento del 2026-08-20.<br>
**Estado:** APROBADO e IMPLEMENTADO PARCIAL; persistencia, owner, join, conteo y transicion inicial cuentan con implementacion y pruebas. Vencimiento e integraciones posteriores permanecen pendientes.<br>
**Dependencias:** Identidad autenticada de cada participante y RF-025 para modificar el plazo o los recursos sujetos.

**ID:** RF-022  
**Titulo:** Ciclo de vida y condicion grupal segun recurso<br>
**Descripcion:** El sistema debe determinar el estado inicial segun la politica del recurso y mantener separadas la etapa de ciclo de vida (`status`) y la condicion operacional del grupo (`groupCondition`).<br>
**Actor:** Usuario normal.  
**Precondiciones:** Recurso activo con politica de reserva definida.  
**Comportamiento esperado:** `OPEN_USE` no exige integrantes. Las multicanchas comienzan `PENDING + PENDING_MINIMUM`, alcanzan `CONFIRMED + HEALTHY` al cumplir el minimo y, si posteriormente bajan del minimo, conservan `CONFIRMED` con `AT_RISK`. Recuperar el minimo devuelve la condicion a `HEALTHY`.<br>
**Resultado:** El ciclo de vida conserva que la reserva fue confirmada y la condicion grupal expresa riesgo o recuperacion. El vencimiento bajo el minimo puede llevar a `CANCELLED` segun la politica.<br>
**Prioridad:** MUST.  
**Fuente:** Decision del 2026-07-20 refinada durante MVP 2 el 2026-08-20.<br>
**Estado:** APROBADO e IMPLEMENTADO PARCIAL; la transicion inicial y `AT_RISK` existen. Vencimiento e integracion de notificaciones permanecen pendientes.<br>
**Dependencias:** RF-020, RF-021, RF-025 y notificaciones.

**ID:** RF-023  
**Titulo:** Resolver conflictos de actividades institucionales  
**Descripcion:** Al registrar o modificar una actividad institucional, el sistema debe detectar conflictos y aplicar el resultado aprobado segun el tipo de ocupacion existente.  
**Actor:** Administrador y usuario normal afectado.  
**Precondiciones:** Actividad institucional y ocupacion solapada sobre el mismo recurso.  
**Comportamiento esperado:** Si el conflicto es con una reserva particular, cancelarla automaticamente, informar el efecto al administrador y notificar al usuario afectado; si es entre dos actividades, permitir al administrador cancelar una de las dos o conservar ambas.<br>
**Resultado:** La agenda conserva la prioridad institucional, el usuario conoce la cancelacion y se registra la decision administrativa cuando compartir espacio es valido.<br>
**Prioridad:** MUST.  
**Fuente:** Decision explicita del usuario del 2026-07-20.  
**Estado:** APROBADO e IMPLEMENTADO PARCIAL. La deteccion N-elementos, consulta y resolucion administrativa cuentan con backend y pruebas de integracion. La cancelacion automatica de reservas particulares no coincide con el modelo implementado, que exige una resolucion administrativa explicita; esta diferencia queda bajo evaluacion como `EV-010`. La notificacion asociada permanece pendiente.<br>
**Dependencias:** Programacion institucional, auditoria y notificaciones.

**ID:** RF-024  
**Titulo:** Mantener inventario oficial de recursos  
**Descripcion:** El administrador debe poder modificar el inventario oficial inicial de ocho recursos y sus datos operativos en el MVP administrativo correspondiente.  
**Actor:** Administrador.  
**Precondiciones:** Rol administrador vigente y recurso identificable.  
**Comportamiento esperado:** Permitir altas, modificaciones, activacion o desactivacion y conservar el historial cuando corresponda.  
**Resultado:** Catalogo y disponibilidad reflejan la configuracion institucional vigente.  
**Prioridad:** MUST para cierre del MVP administrativo.  
**Fuente:** Decision explicita del usuario del 2026-07-20 y RF-014.  
**Estado:** APROBADO; lectura e imagen IMPLEMENTADAS, gestion completa PENDIENTE.  
**Dependencias:** Politica de conservacion de recursos con historial.

**ID:** RF-025<br>
**Titulo:** Configurar politicas institucionales de reserva<br>
**Descripcion:** Solo usuarios con rol administrador deben poder modificar el periodo de reserva, el plazo previo y los recursos sujetos a confirmacion grupal.<br>
**Actor:** Administrador.<br>
**Precondiciones:** Cuenta vigente con rol administrador.<br>
**Comportamiento esperado:** Mostrar valores vigentes, aceptar cambios autorizados e indicar desde cuando rigen.<br>
**Resultado:** Politica institucional actualizada sin facultades equivalentes para usuarios normales.<br>
**Prioridad:** SHOULD para MVP 3.<br>
**Fuente:** Decision explicita del usuario del 2026-07-20.<br>
**Estado:** APROBADO e IMPLEMENTADO PARCIAL. La publicacion prospectiva, historial, permisos, idempotencia, clasificacion de recursos grupales, participantes persistidos, minimo y transiciones principales cuentan con implementacion incremental. Permanecen pendientes la interfaz administrativa completa, el vencimiento automatico, las correcciones excepcionales y la evidencia integral del ambiente objetivo.<br>
**Dependencias:** RF-002, RF-020 a RF-022 y arquitectura aprobada de versionado prospectivo con correcciones excepcionales.

#### Requisitos no funcionales

**ID:** RNF-001  
**Titulo:** Zona horaria institucional  
**Descripcion:** Fechas y horas de agenda deben interpretarse y mostrarse en `America/Santiago`, incluyendo cambios de horario de verano.  
**Actor:** Todos.  
**Precondiciones:** Fecha/hora valida.  
**Comportamiento esperado:** Valores con offset se convierten a Santiago; valores sin offset se interpretan como hora institucional.  
**Resultado:** La misma reserva conserva fecha, hora y categoria temporal entre cliente y servidor.  
**Prioridad:** MUST.  
**Fuente:** RNF-010 y codigo de reloj.  
**Estado:** IMPLEMENTADO y VERIFICADO local; verificacion online PENDIENTE.  
**Dependencias:** Ninguna de producto.

**ID:** RNF-002  
**Titulo:** Privacidad por rol  
**Descripcion:** Un usuario normal no debe recibir identidad, correo o RUT de reservas ajenas al consultar disponibilidad.  
**Actor:** Usuario normal.  
**Precondiciones:** Consulta autenticada.  
**Comportamiento esperado:** Reducir la respuesta al dato necesario para decidir disponibilidad.  
**Resultado:** Ocupacion visible sin datos personales ajenos.  
**Prioridad:** MUST.  
**Fuente:** RNF-008 y handler de disponibilidad.  
**Estado:** IMPLEMENTADO; prueba automatizada especifica PENDIENTE.  
**Dependencias:** RF-003.

**ID:** RNF-003  
**Titulo:** Accesibilidad y adaptacion  
**Descripcion:** Los flujos criticos deben ser operables con teclado, foco visible, etiquetas, mensajes anunciables y anchos desde 320 px.  
**Actor:** Todos.  
**Precondiciones:** Interfaz cargada.  
**Comportamiento esperado:** Mantener controles visibles, foco dentro de dialogos, cierre con Escape y retorno al disparador.  
**Resultado:** Reserva y cancelacion utilizables sin puntero y en movil.  
**Prioridad:** MUST para cierre de demo.  
**Fuente:** RNF-009, RNF-012, `BACK-021` y `BACK-022`.  
**Estado:** PENDIENTE.  
**Dependencias:** Evaluacion UI/UX.

**ID:** RNF-004  
**Titulo:** Evidencia de calidad  
**Descripcion:** Cada declaracion de cierre debe respaldarse con pruebas automatizadas y validacion integrada proporcional al flujo.  
**Actor:** Equipo del proyecto.  
**Precondiciones:** Incremento candidato a cierre.  
**Comportamiento esperado:** Registrar comando, fecha, resultado y limitaciones.  
**Resultado:** Estado `VERIFICADO` diferenciable de `IMPLEMENTADO`.  
**Prioridad:** MUST.  
**Fuente:** RNF-005 y checklist de demo.  
**Estado:** IMPLEMENTADO PARCIAL.  
**Dependencias:** Ambiente local y, para cierre, ambiente integrado/online.

### 6. Reglas de negocio

**RN-001 — Identidad confiable**  
**Regla:** La identidad de una operacion proviene del token validado o del modo local explicitamente habilitado; el cliente no decide el usuario propietario.  
**Justificacion:** Evita suplantacion.  
**Fuente:** Middleware y RF-001.  
**Estado:** IMPLEMENTADO.  
**Excepciones:** Modo local solo con `DEV_AUTH_ENABLED=true`.

**RN-002 — Autorizacion administrativa**  
**Regla:** Toda accion administrativa debe validar el rol en el servidor.  
**Justificacion:** Ocultar una vista no constituye autorizacion.  
**Fuente:** Middleware y rutas.  
**Estado:** IMPLEMENTADO para las rutas existentes.  
**Excepciones:** Ninguna.

**RN-003 — RUT previo**  
**Regla:** Un usuario normal sin RUT no puede crear reservas ni inscribirse en talleres; un administrador no queda sujeto a este bloqueo.  
**Justificacion:** Identificacion minima vigente del solicitante.  
**Fuente:** Handlers de reservas y talleres.  
**Estado:** IMPLEMENTADO.  
**Excepciones:** Administrador.

**RN-004 — Tiempo institucional**  
**Regla:** Las fechas de agenda se interpretan en `America/Santiago`; un instante con offset se convierte y uno sin offset se entiende como hora local institucional.  
**Justificacion:** Evita diferencias por servidor, navegador y horario de verano.  
**Fuente:** Codigo de reloj y RNF-010.  
**Estado:** IMPLEMENTADO; VERIFICADO local.  
**Excepciones:** Ninguna definida.

**RN-005 — Jornada y duracion aprobadas**<br>
**Regla:** El inicio debe estar entre 08:00 inclusive y 22:00 exclusiva, en pasos de 15 minutos; el termino puede ser exactamente 22:00; las duraciones admitidas son 30 a 180 minutos en incrementos de 30.  
**Justificacion:** Es el contrato observable actual y respeta la decision de que una reserva no tiene que durar exactamente una hora.  
**Fuente:** Reglas backend/frontend y decision explicita del usuario del 2026-07-20.  
**Estado:** APROBADO para todos los recursos, IMPLEMENTADO y VERIFICADO local; verificacion integrada/online PENDIENTE.<br>
**Excepciones:** Ninguna definida.

**RN-006 — Ciclo de vida y condicion de reservas grupales**  
**Regla:** El cliente no decide el estado. `reservation.status` representa el ciclo de vida y `groupCondition` representa la condicion operacional del grupo. `OPEN_USE` no requiere confirmacion de participantes. Las Canchas 1, 2 y 3 comienzan `PENDING + PENDING_MINIMUM`; al alcanzar el minimo cambian a `CONFIRMED + HEALTHY`. Si posteriormente bajan del minimo conservan `CONFIRMED` y cambian a `AT_RISK`; si recuperan el minimo vuelven a `HEALTHY`. El vencimiento bajo el minimo puede llevar a `CANCELLED` conforme a la politica.<br>
**Justificacion:** Evita hacer oscilar el estado persistido de una reserva ya confirmada, conserva el hecho de que el minimo fue alcanzado y permite que riesgo, recuperacion y vencimiento funcionen como eventos diferenciados para notificaciones y automatizaciones posteriores.  
**Fuente:** Decision inicial del 2026-07-20, refinada durante MVP 2 el 2026-08-20 y respaldada por la implementacion/pruebas del flujo grupal.  
**Estado:** APROBADO e IMPLEMENTADO PARCIAL. `PENDING`, participantes persistidos, transicion a `CONFIRMED` y `AT_RISK` estan implementados; vencimiento automatico e integraciones posteriores permanecen incrementales.  
**Excepciones:** Recursos cuya politica oficial no exija participantes.

**RN-007 — Conflictos y modos de recurso**  
**Regla:** Recursos `RESERVABLE` no admiten solapamientos confirmados; `INFORMATIVE` no admite reserva; `ADMIN_ONLY` exige administrador; `OPEN_USE` permite concurrencia de reservas y no queda bloqueado por programacion o talleres, aunque un bloqueo activo si impide reservar. Un usuario no puede tener dos reservas confirmadas solapadas.  
**Justificacion:** Refleja el comportamiento de base de datos y servicio.  
**Fuente:** `database/schema.sql` y servicio de reservas.  
**Estado:** IMPLEMENTADO en parte y APROBADO para que `OPEN_USE` no requiera confirmacion grupal.  
**Excepciones:** Las indicadas por modo.

**RN-008 — Cancelacion**  
**Regla:** Solo el propietario o un administrador puede cancelar una reserva `PENDING` o `CONFIRMED` cuyo termino sea posterior al momento actual.  
**Justificacion:** Protege propiedad e historial.  
**Fuente:** Servicio y repositorio de reservas.  
**Estado:** IMPLEMENTADO y parcialmente VERIFICADO.  
**Excepciones:** Ninguna.

**RN-009 — Frecuencia semanal**  
**Regla:** La longitud del periodo es configurable y actualmente corresponde a siete dias calendario en `America/Santiago`. `PENDING` consume la oportunidad desde su creacion; `CONFIRMED` la mantiene consumida; al pasar a `CANCELLED` deja de consumirla y se recalcula la proxima fecha permitida.<br>
**Justificacion:** Regla levantada para distribuir acceso.  
**Fuente:** Levantamiento, alcance definitivo y decision explicita del usuario del 2026-07-20.  
**Estado:** APROBADO e IMPLEMENTADO. Existe evidencia local y PostgreSQL vigente; falta cierre integral en el ambiente objetivo.<br>
**Excepciones:** Una solicitud rechazada que no llega a crearse no consume la frecuencia.

**RN-010 — Participantes minimos**  
**Regla:** Multicancha 1, 2 y 3, identificadas como Cancha 1, 2 y 3, requieren al menos 10 usuarios unicos con cuenta y confirmacion vigente. El solicitante cuenta una vez dentro del minimo.<br>
**Justificacion:** Regla levantada para justificar el uso del espacio.  
**Fuente:** Levantamiento, alcance definitivo y decision explicita del usuario del 2026-07-20.  
**Estado:** APROBADO e IMPLEMENTADO PARCIAL; participantes persistidos, solicitante incluido, conteo y transiciones principales cuentan con implementacion. El cierre temporal completo sigue pendiente.<br>
**Excepciones:** `OPEN_USE` y los demas recursos no clasificados para confirmacion grupal.

**RN-011 — Prioridad institucional**  
**Regla:** RF-023/v1 establece cancelacion automatica de una reserva particular ante prioridad institucional. La implementacion MVP 2 utiliza conflictos administrables con resoluciones `KEEP`, `ALLOW`, `CANCEL` y `RESCHEDULE`. `EV-010` debe decidir cual de ambos modelos queda vigente para reservas particulares. Entre actividades institucionales ya existe resolucion administrativa explicita.<br>
**Justificacion:** Mantener disponibilidad para clases y actividades oficiales.  
**Fuente:** Levantamiento y decision explicita del usuario del 2026-07-20.  
**Estado:** APROBADO; no IMPLEMENTADO.  
**Excepciones:** Dos actividades pueden compartir el mismo espacio por decision administrativa.

**RN-012 — Privacidad de disponibilidad**  
**Regla:** Usuarios normales solo reciben ocupacion y datos institucionales genericos; administradores pueden ver detalle operacional.  
**Justificacion:** Minimiza exposicion de RUT, correo y nombre.  
**Fuente:** Handler de disponibilidad y RNF-008.  
**Estado:** IMPLEMENTADO.  
**Excepciones:** Datos propios visibles en las vistas personales.

**RN-013 — Inventario oficial administrable**  
**Regla:** Los ocho recursos actuales constituyen el inventario oficial inicial y el administrador debe poder modificarlo en un MVP posterior sin requerir cambios de codigo.  
**Justificacion:** Mantener el catalogo alineado con la operacion real.  
**Fuente:** Decision explicita del usuario del 2026-07-20 y `database/seed.sql`.  
**Estado:** APROBADO; lectura IMPLEMENTADA y gestion completa PENDIENTE.  
**Excepciones:** Eliminacion de recursos con historial requiere politica de conservacion.

**RN-014 — Bloqueo y vencimiento de solicitud grupal**<br>
**Regla:** Una solicitud grupal `PENDING` bloquea el horario. Se admiten confirmaciones o retiradas hasta exactamente el limite, inicialmente una hora antes inclusive. Si llega al limite bajo 10 confirmaciones, cambia a `CANCELLED` y libera horario y oportunidad semanal.<br>
**Justificacion:** Evita doble asignacion y solicitudes pendientes sin salida.<br>
**Fuente:** Decision explicita del usuario del 2026-07-20.<br>
**Estado:** APROBADO; no IMPLEMENTADO.<br>
**Excepciones:** `OPEN_USE` no aplica confirmacion grupal.

**RN-015 — Administracion exclusiva de politicas**<br>
**Regla:** Solo usuarios con rol administrador pueden modificar el periodo de reserva, el plazo previo o los recursos sujetos a confirmacion grupal.<br>
**Justificacion:** Evita que usuarios normales alteren reglas institucionales.<br>
**Fuente:** Decision explicita del usuario del 2026-07-20.<br>
**Estado:** APROBADO; no IMPLEMENTADO.<br>
**Excepciones:** Ninguna.

### 7. Flujos

#### Flujo principal: crear reserva actual

1. El usuario autenticado consulta la disponibilidad de una fecha.
2. El sistema carga recursos, reservas sanitizadas, actividades programadas y talleres.
3. El usuario selecciona recurso, inicio, duracion y actividad opcional.
4. La interfaz valida datos y conflictos conocidos.
5. El servidor obtiene al usuario real, exige RUT al usuario normal y asigna el estado inicial.
6. El servidor valida pasado, zona, jornada, duracion, modo de recurso y talleres.
7. La base rechaza conflictos concurrentes con reservas, bloqueos y actividades programadas.
8. El sistema confirma la reserva o mantiene un error recuperable.

Alternativas y errores:

- Sin RUT: rechazo antes de crear.
- Recurso inactivo, informativo o `ADMIN_ONLY` para usuario normal: el servidor rechaza; la prevencion en interfaz sigue pendiente.
- Horario pasado o fuera de jornada: rechazo.
- Conflicto no visible, por ejemplo un bloqueo: rechazo al confirmar.
- Recurso `OPEN_USE`: permite concurrencia conforme a la regla actual.

#### Flujo vigente: confirmacion por participantes

1. El usuario solicita una de las tres multicanchas dentro de la ventana configurable.
2. El sistema registra la solicitud como `PENDING + PENDING_MINIMUM`; desde ese momento consume la oportunidad semanal y bloquea el horario.
3. El solicitante cuenta una vez. Los demas participantes se autentican con sus cuentas y confirman sin duplicados hasta el limite aplicable.
4. Mientras la solicitud nunca haya alcanzado el minimo y existan menos de 10 confirmaciones validas, permanece `PENDING`.
5. Al alcanzar 10 confirmaciones, cambia a `CONFIRMED + HEALTHY`, siempre que las demas reglas sigan cumpliendose.
6. Si posteriormente una retirada deja menos de 10, conserva `CONFIRMED` y cambia a `AT_RISK`.
7. Si recupera el minimo antes del vencimiento, conserva `CONFIRMED` y vuelve a `HEALTHY`.
8. Si llega al limite aplicable bajo el minimo, el flujo de vencimiento debe llevarla a `CANCELLED`, liberar el horario y dejar de consumir la oportunidad semanal.
9. Para `OPEN_USE`, el sistema no exige este proceso de confirmacion grupal.

Alternativa: un intento de confirmacion posterior al limite se rechaza sin alterar el conteo ni el estado ya resuelto.

#### Flujo objetivo aprobado: conflicto institucional

1. El administrador registra o modifica una actividad institucional.
2. El sistema detecta todas las ocupaciones solapadas sobre el recurso.
3. Si existe una reserva particular, el sistema la incorpora al conflicto. La cancelacion automatica corresponde a RF-023/v1; la resolucion administrativa implementada queda bajo decision `EV-010`. La notificacion al usuario afectado permanece requerida cuando exista cancelacion.
4. Si existe otra actividad institucional, el sistema informa el conflicto.
5. El administrador elige cancelar una de las actividades o mantener ambas cuando puedan compartir el espacio.
6. La decision y sus efectos quedan trazables.

#### Flujo principal: cancelar reserva

1. El usuario selecciona una reserva visible.
2. El sistema valida propiedad o rol administrador.
3. El sistema solicita confirmacion cuando el acceso la implementa.
4. El servidor valida estado y termino en hora institucional.
5. La reserva cambia a `CANCELLED` y la vista se actualiza.

Errores:

- Reserva ajena para usuario normal, inexistente, ya cancelada, rechazada, expirada o finalizada: no se modifica.

### 8. Criterios de aceptacion

**CA-001**  
Dado un usuario sin sesion valida, cuando intenta acceder a una funcion interna, entonces no obtiene datos ni ejecuta la operacion y se le solicita autenticarse.

**CA-002**  
Dado un usuario normal, cuando intenta una ruta administrativa, entonces el servidor responde sin autorizar aunque la solicitud no provenga de la interfaz.

**CA-003**  
Dado un usuario normal sin RUT, cuando intenta crear una reserva o inscribirse en un taller, entonces la operacion es rechazada sin persistir cambios.

**CA-004**  
Dada una reserva valida a las 08:00 o una que termina exactamente a las 22:00, cuando se confirma, entonces supera la validacion horaria; un inicio anterior a 08:00, igual o posterior a 22:00 o un termino posterior a 22:00 es rechazado.

**CA-005**  
Dado un inicio que no coincide con un intervalo de 15 minutos o una duracion fuera de 30, 60, 90, 120, 150 y 180 minutos, cuando se envia, entonces el servidor lo rechaza incluso si el cliente fue manipulado.

**CA-006**  
Dado un payload de creacion con campos de usuario o estado no permitidos, cuando se procesa, entonces se rechaza el contrato desconocido o se utiliza exclusivamente el usuario y estado controlados por servidor.

**CA-007**  
Dada una reserva confirmada que se solapa por recurso o por usuario, cuando se intenta crear otra reserva confirmada incompatible, entonces la base rechaza la operacion completa.

**CA-008**  
Dado un bloqueo activo o actividad programada que se cruza con una reserva sobre un recurso bloqueable, cuando se intenta reservar, entonces la operacion se rechaza y el mensaje no revela datos personales ni internos.

**CA-009**  
Dada una consulta de disponibilidad de usuario normal, cuando existen reservas ajenas, entonces la respuesta no contiene su nombre, correo, RUT ni identificador de usuario.

**CA-010**  
Dada una reserva propia activa y no finalizada, cuando el usuario confirma la cancelacion, entonces cambia a `CANCELLED`; para una reserva ajena, no cancelable o finalizada, no cambia.

**CA-011**  
Dado un taller activo con cupo y un usuario con RUT no inscrito, cuando solicita inscripcion, entonces queda inscrito una vez; sin cupo, sin RUT o duplicado, se rechaza.

**CA-012**  
Dado un dialogo critico abierto en un viewport de 320 px, cuando se usa solo teclado, entonces el foco entra al dialogo, permanece dentro, Escape cierra cuando no hay proceso activo y el foco vuelve al control de origen.

**CA-013**  
Dado un candidato a cierre, cuando se informa como `VERIFICADO`, entonces existe evidencia fechada de pruebas automatizadas y del ambiente integrado aplicable.

**CA-014**  
Dado un periodo vigente de siete dias y una solicitud creada un martes en `America/Santiago`, cuando el mismo usuario intenta crear otra antes del martes siguiente, entonces se rechaza y se comunica el martes siguiente como proxima fecha permitida; desde ese martes puede continuar si cumple las demas reglas.

**CA-015**  
Dada una solicitud nueva sobre Cancha 1, Cancha 2 o Cancha 3 que aun nunca ha alcanzado el minimo y tiene menos de 10 participantes unicos confirmados, cuando se consulta su estado, entonces permanece `PENDING + PENDING_MINIMUM`.

**CA-016**  
Dada una solicitud grupal con 9 participantes confirmados, cuando se registra la decima confirmacion valida y las demas reglas siguen cumpliendose, entonces cambia automaticamente a `CONFIRMED` una sola vez.

**CA-017**  
Dado un recurso `OPEN_USE`, cuando un usuario cumple sus condiciones de acceso, entonces no se le exige reunir confirmaciones de participantes.

**CA-018**  
Dada una nueva actividad institucional que se solapa con una reserva particular, cuando se confirma la programacion institucional, entonces la reserva particular cambia automaticamente a `CANCELLED`, el administrador ve el efecto y el usuario afectado recibe una notificacion sin datos de terceros.

**CA-019**  
Dadas dos actividades institucionales solapadas, cuando el administrador revisa el conflicto, entonces puede cancelar cualquiera de ellas o mantener ambas; el sistema no decide automaticamente entre actividades.

**CA-020**  
Dado el inventario oficial de ocho recursos, cuando un administrador autorizado modifica un recurso en el MVP correspondiente, entonces el catalogo y la disponibilidad reflejan el cambio y el historial previo se conserva conforme a la politica aprobada.

**CA-021**<br>
Dado un periodo vigente de siete dias y que hoy es martes en `America/Santiago`, cuando el usuario elige una fecha, entonces puede elegir desde ese martes hasta el lunes siguiente inclusive y se rechaza el martes posterior por quedar fuera de la ventana.

**CA-022**<br>
Dada una reserva de Cancha 1, 2 o 3 que ya alcanzo 10 confirmaciones y se encuentra `CONFIRMED + HEALTHY`, cuando una persona retira su confirmacion y quedan 9 vigentes antes del limite, entonces conserva `CONFIRMED`, cambia a `AT_RISK` y el conteo visible se actualiza.

**CA-023**<br>
Dado que el limite vigente es una hora antes del inicio, cuando una persona confirma o retira exactamente en ese instante, entonces el cambio se acepta y se evalua inmediatamente el minimo; si queda bajo 10, se aplica la cancelacion de vencimiento. Cualquier intento posterior se rechaza.

**CA-024**<br>
Dada una solicitud `PENDING`, cuando se crea, entonces consume la oportunidad semanal; cuando cambia a `CANCELLED`, deja de consumirla y se recalcula la proxima fecha permitida.

**CA-025**<br>
Dada una solicitud de multicancha, cuando se inicia el conteo, entonces el solicitante cuenta una vez y solo cuentas autenticadas pueden agregar confirmaciones unicas.

**CA-026**<br>
Dada una solicitud grupal `PENDING`, cuando otro usuario consulta el mismo horario para un uso incompatible, entonces aparece ocupado; si llega al limite bajo 10 confirmaciones, cambia a `CANCELLED` y el horario queda disponible.

**CA-027**<br>
Dado un usuario normal y un administrador, cuando intentan modificar periodo, plazo o recursos sujetos a confirmacion, entonces solo el administrador puede completar el cambio y el sistema informa desde cuando rige.

### 9. Casos limite

- Inicio exactamente a 08:00; termino exactamente a 22:00; inicio a 22:00.
- Fecha/hora sin offset, con offset de verano, con offset de invierno y cruce de medianoche UTC.
- Reserva cuyo inicio ya paso pero cuyo termino no; reserva que termina exactamente en el presente.
- Dos solicitudes concurrentes para el mismo recurso o el mismo usuario.
- Reserva `OPEN_USE` concurrente con otra reserva, taller, actividad programada o bloqueo.
- Recurso activo que cambia a inactivo entre consulta y confirmacion.
- Bloqueo existente que no aparece en la interfaz pero impide confirmar.
- Reserva cancelada mostrada en historial que no debe bloquear disponibilidad.
- Usuario bloqueado durante una sesion ya iniciada.
- Taller que alcanza el ultimo cupo con solicitudes concurrentes.
- Ventana semanal un martes, limite entre lunes y martes siguiente, cambio autorizado del periodo y horario de verano.
- Solicitud creada un dia y reserva fijada para otro; cancelacion posterior y recalculo de la siguiente fecha permitida.
- Solicitante contado una sola vez; participante repetido, sin cuenta o que retira su confirmacion antes y despues del limite.
- Retirada que reduce el conteo a 9 exactamente una hora antes del inicio y cancelacion inmediata por llegar al limite bajo el minimo.
- Solicitud grupal que alcanza el minimo despues de que aparece otro conflicto.
- Actividad institucional contra varias reservas particulares simultaneas.
- Dos actividades que comparten solo una parte del horario o del espacio.

### 10. Trazabilidad

| Requisito | Regla relacionada | Criterios de aceptacion | Fuente |
| --- | --- | --- | --- |
| RF-001 | RN-001 | CA-001 | Middleware, rutas, RF-001 |
| RF-002 | RN-002, RN-003 | CA-002, CA-003 | Middleware, usuarios, RF-002 a RF-004 |
| RF-003 | RN-007, RN-012 | CA-007, CA-008, CA-009 | RF-005, disponibilidad, esquema |
| RF-004 | RN-001, RN-004 a RN-007, RN-009, RN-010, RN-014 | CA-003 a CA-008, CA-014 a CA-017, CA-021 a CA-026 | RF-006, RF-007, servicio, esquema y decisiones del 2026-07-20 |
| RF-005 | RN-008 | CA-010 | RF-008 y servicio |
| RF-006 | RN-003 | CA-011 | RF-019 y talleres |
| RF-007 | RN-002, RN-012, RN-013 | CA-002, CA-009, CA-020 | RF-012 a RF-018 y decision del 2026-07-20 |
| RF-020 | RN-009 | CA-014, CA-021, CA-024 | Levantamiento, alcance definitivo y decisiones del 2026-07-20 |
| RF-021 | RN-010, RN-014 | CA-015, CA-016, CA-022, CA-023, CA-025, CA-026 | Levantamiento, alcance definitivo y decisiones del 2026-07-20 |
| RF-022 | RN-006, RN-007, RN-010, RN-014 | CA-015 a CA-017, CA-022 a CA-026 | Decisiones del 2026-07-20 |
| RF-023 | RN-011 | CA-018, CA-019 | Decision del 2026-07-20 |
| RF-024 | RN-013 | CA-020 | RF-014 y decision del 2026-07-20 |
| RF-025 | RN-002, RN-015 | CA-027 | Decision del 2026-07-20 |
| RNF-001 | RN-004 | CA-004, CA-013 | Reloj, pruebas y RNF-010 |
| RNF-002 | RN-012 | CA-009 | Handler de disponibilidad |
| RNF-003 | — | CA-012 | RNF-009, RNF-012 |
| RNF-004 | — | CA-013 | RNF-005 y checklist |

### 11. Contradicciones

| ID | Fuentes en conflicto | Diferencia | Efecto |
| --- | --- | --- | --- |
| C-01 | Alcance definitivo vs repositorio | Entra ID real y demo Azure estaban fuera del alcance original y hoy estan implementados/documentados. | El informe puede describir un alcance distinto del prototipo entregado. |
| C-03 | Reglas implementadas vs evidencia de cierre | La ventana/frecuencia y parte del flujo grupal ya se aplican localmente, pero faltan verificaciones integradas y cierre del vencimiento. | MVP 2 no debe declararse completamente verificado todavia. |
| C-04 | Evolucion documental del flujo grupal | La regla anterior hacia regresar una reserva confirmada a `PENDING`; la decision vigente utiliza `CONFIRMED + AT_RISK`. | Historias, criterios y documentos antiguos deben conservarse como trazabilidad, no como regla vigente. |
| C-05 | Prioridad institucional aprobada vs esquema actual | Debe cancelarse y notificarse la reserva particular y permitirse decision entre actividades, pero el esquema rechaza el solapamiento al registrar la actividad. | El comportamiento aprobado no puede ejecutarse. |
| C-06 | Inventario y politicas aprobadas vs administracion actual | Los ocho recursos son oficiales; la API administrativa ya publica politicas prospectivas, pero faltan interfaz y gestion completa del inventario. | La operacion institucional aun no puede mantenerse completamente desde la interfaz. |
| C-07 | Disponibilidad visible vs regla de base | Bloqueos impiden reservar, pero no se muestran en el endpoint actual. | Un horario puede parecer libre y fallar al confirmar. |

### 12. Preguntas y decisiones pendientes

No quedan preguntas bloqueantes para el incremento analizado.

**DECISION APROBADA el 2026-07-20 y detallada el 2026-07-21:** los cambios administrativos normales crean snapshots completos con vigencia inmediata y cada solicitud conserva la version obtenida por su transaccion al adquirir primero el bloqueo. El historial completo es administrativo y el usuario recibe solo condiciones operativas. La correccion excepcional permanece fuera de este incremento y requerira vista previa temporal de un solo uso vinculada al administrador, motivo, aplicacion atomica y auditoria, sin editar versiones historicas ni cancelar solicitudes implicitamente. La retencion institucional definitiva sigue pendiente; no se elimina auditoria mientras existan reservas relacionadas.

### 13. Riesgos de producto

| Riesgo | Probabilidad | Impacto | Mitigacion propuesta |
| --- | --- | --- | --- |
| Defender como cumplidas reglas aprobadas pero no verificadas | Alta | Alto | Separar evidencia historica Azure SQL de validacion PostgreSQL vigente; presentar cada RF como implementado, parcial o pendiente segun pruebas actuales y no segun el motor anterior. |
| Mostrar disponibilidad incompleta por omitir bloqueos | Alta | Alto | Mantener el rechazo servidor y priorizar la visibilidad de bloqueos antes de operacion real. |
| Cancelar automaticamente sin informar al usuario afectado | Media | Alto | Tratar el aviso ya aprobado como criterio obligatorio y verificar su entrega antes de habilitar prioridad institucional. |
| Solicitudes pendientes que acaparen horarios | Media | Alto | Aplicar el bloqueo aprobado y cancelacion automatica al vencer bajo el minimo. |
| Cambiar politicas de forma retroactiva y sorprender a usuarios | Media | Alto | Aplicar versiones prospectivas; limitar excepciones a solicitudes seleccionadas con simulacion, motivo, atomicidad y auditoria. |
| Confundir demo Azure con sistema productivo | Media | Alto | Etiquetar el ambiente como demo y exigir checklist online antes de presentarlo como disponible. |
| Exponer datos o detalles internos en errores | Media | Alto | Completar criterios de minimizacion y mensajes seguros antes de usuarios reales. |
| Experiencia inconsistente en movil o teclado | Alta | Medio | Evaluacion UI/UX y validacion manual en anchos y flujos criticos definidos. |
| Indicadores interpretados como reportes oficiales | Media | Medio | Etiquetarlos como indicadores iniciales hasta aprobar definiciones y fuentes. |
| Documentos academicos y tecnicos divergen nuevamente | Alta | Alto | Designar fuentes canonicas y actualizar trazabilidad al aprobar cada decision. |

### 14. Recomendacion al Orquestador

```text
Estado de la entrega:
APROBABLE

Siguiente rol recomendado:
ORQUESTADOR / REVISION DE PRODUCTO

Motivo:
La base funcional, el versionado de politicas y una parte sustantiva del flujo grupal ya estan implementados. El siguiente trabajo debe validar la trazabilidad reconstruida, cerrar vencimiento e integraciones pendientes y continuar luego con administracion y notificaciones.

Contexto que debería recibir:
docs/00-resumen-proyecto.md, este informe, docs/08-requisitos-historias-casos-uso.md, docs/09-mvps-roadmap.md, docs/12-checklist-demo-mvp1.md y los documentos historicos de levantamiento/alcance.

Decisiones aprobadas:
Aplicacion prospectiva, correcciones retroactivas excepcionales controladas, cuatro incrementos tecnicos y prohibicion de retiro del solicitante. `ADMIN-005` ya posee una implementacion backend parcial y su diferencia entre cancelacion automatica y resolucion administrativa se mantiene como decision de producto pendiente (`EV-010`).
```

### 15. Cambios propuestos a fuentes de verdad

| Documento | Seccion | Cambio propuesto | Motivo |
| --- | --- | --- | --- |
| `../Documentos/alcance_definitivo_prototipo_poli_redi.txt` | Alcance y reglas | Incorporar estados que consumen oportunidad, participantes con cuenta, bloqueo `PENDING`, vencimiento, nombres oficiales y administracion exclusiva. | Mantener el alcance academico alineado con el producto. |
| Informe de tesis | Problema, alcance, requisitos y resultados | Incorporar ventana configurable, confirmacion condicional reversible, prioridad institucional con aviso e inventario oficial, distinguiendo aprobado de implementado. | Mantener una defensa coherente y comprobable. |
| Documento institucional de operacion | Frecuencia, participantes y conflictos | Formalizar valores iniciales, administracion exclusiva, cancelacion por vencimiento, versionado prospectivo y correcciones excepcionales auditadas. | Evitar que reglas criticas dependan solo del codigo o del informe. |
