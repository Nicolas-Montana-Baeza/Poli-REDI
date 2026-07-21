# Poli-REDI - Requisitos, historias de usuario y casos de uso

## Objetivo del documento

Este documento consolida los requisitos funcionales, requisitos no funcionales, historias de usuario y casos de uso principales de Poli-REDI.

Debe mantenerse actualizado junto con el backlog, la arquitectura, el flujo de reservas y la documentacion tecnica cada vez que cambie el alcance funcional del sistema.

## Alcance del sistema

Poli-REDI es un sistema web para gestionar reservas deportivas institucionales. Permite a usuarios autenticados consultar disponibilidad, registrar reservas, cancelar reservas propias, revisar historial, inscribirse en talleres deportivos y recibir notificaciones. Los administradores pueden revisar informacion operacional, usuarios, recursos, reportes y, en iteraciones futuras, gestionar bloqueos, recursos, programacion institucional e infracciones.

## Actores

### Usuario normal

Persona autenticada que usa recursos deportivos institucionales.

Responsabilidades principales:

- Consultar disponibilidad.
- Crear reservas propias.
- Registrar o actualizar RUT cuando corresponda.
- Revisar sus reservas e historial.
- Consultar talleres deportivos e inscribirse cuando existan cupos.
- Cancelar reservas permitidas.
- Revisar notificaciones.

### Administrador

Usuario autenticado con permisos administrativos.

Responsabilidades principales:

- Acceder al panel administrativo.
- Revisar usuarios, recursos, reservas y reportes.
- Cancelar reservas cuando corresponda.
- Gestionar usuarios, recursos, bloqueos, programacion e infracciones en futuras iteraciones.

### Sistema de autenticacion Microsoft Entra ID

Proveedor externo de identidad.

Responsabilidades principales:

- Autenticar usuarios.
- Entregar tokens para consumir la API protegida.
- Proveer datos base de identidad, como correo, nombre, tenant y object id.

### Base de datos Azure SQL

Sistema de persistencia.

Responsabilidades principales:

- Guardar usuarios, recursos, actividades, reservas, talleres, inscripciones, notificaciones y datos administrativos.
- Aplicar reglas de integridad y restricciones criticas.
- Proveer vistas de apoyo para reportes y calendario.

## Requisitos funcionales

### RF-001 - Autenticacion de usuarios

El sistema debe permitir ingreso mediante Microsoft Entra ID y, en ambiente de desarrollo, mediante autenticacion local controlada por variables de entorno.

Estado actual: Implementado.

### RF-002 - Sincronizacion de usuario autenticado

El sistema debe obtener o crear el usuario local asociado a la identidad autenticada.

Estado actual: Implementado. La identidad Entra se asocia al usuario local y se sincronizan `oid` y tenant; falta validacion integrada/online del corte actual.

### RF-003 - Registro y validacion de RUT

El sistema debe permitir que usuarios normales registren o actualicen su RUT, validando formato y digito verificador.

Estado actual: Implementado.

### RF-004 - Bloqueo de reserva sin RUT

El sistema debe impedir que usuarios normales sin RUT creen reservas.

Estado actual: Implementado.

### RF-005 - Consulta de disponibilidad

El sistema debe mostrar disponibilidad de recursos por fecha, considerando reservas existentes, talleres activos, actividades institucionales y bloqueos asociados al recurso.

Estado actual: Implementado parcialmente. El endpoint sanitizado combina reservas con actividades institucionales y la UI suma talleres recurrentes; falta incorporar bloqueos y filtros backend por fecha/rango.

Pendiente relacionado: `RES-004`.

### RF-006 - Creacion de reservas

El sistema debe permitir crear solicitudes de reserva sobre recursos disponibles, asociadas al usuario autenticado y, opcionalmente, a una actividad. El servidor debe asignar el estado segun la politica del recurso, aplicar la restriccion semanal y aceptar para todos los recursos solo duraciones de 30, 60, 90, 120, 150 o 180 minutos dentro de la jornada institucional.

Estado actual: Implementado parcialmente. La asociacion al usuario, el estado controlado por servidor, la zona `America/Santiago` y el catalogo de duraciones aprobado tienen pruebas locales; falta la restriccion semanal y la confirmacion condicional por participantes aprobadas el 2026-07-20.

Pendiente relacionado: `RES-008`, `RES-009`, `RES-010`, `RES-011`.

### RF-007 - Validacion de conflictos de reserva

El sistema debe rechazar reservas que se solapen con otras reservas confirmadas, talleres activos o reglas de disponibilidad. Los recursos `OPEN_USE` permiten concurrencia y no bloquean por reserva existente.

Estado actual: Implementado parcialmente para reservas, talleres activos, bloqueos y actividades programadas. El servidor controla el estado inicial y la base protege conflictos concurrentes; falta mostrar bloqueos en disponibilidad y ampliar pruebas integradas.

Pendiente relacionado: `RES-004`, `RES-010`, `ADMIN-004`.

### RF-008 - Cancelacion de reservas

El sistema debe permitir cancelar reservas propias y permitir que administradores cancelen reservas segun permisos, siempre que el estado admita cancelacion y la reserva no haya finalizado segun la zona horaria institucional.

Estado actual: Implementado parcialmente. Propiedad, rol, estados cancelables, reserva finalizada y zona horaria ya se validan; falta una confirmacion consistente en todos los puntos de acceso y verificacion integrada/online.

Pendiente relacionado: `RES-009`, `RES-010`.

### RF-009 - Reservas por rol

El sistema debe mostrar reservas con fecha, recurso, actividad, duracion y estado. Usuarios normales solo ven sus reservas; administradores pueden ver todas las reservas.

Estado actual: Implementado.

### RF-010 - Historial por rol

El sistema debe mostrar reservas pasadas, canceladas o finalizadas. Usuarios normales solo ven su historial; administradores pueden ver todo el historial.

Estado actual: Implementado. Historial usa datos reales, filtros por estado/fecha y acceso al detalle segun rol.

### RF-011 - Catalogo de recursos

El sistema debe mostrar recursos deportivos con datos reales e imagen configurable cuando exista.

Estado actual: Implementado.

### RF-012 - Panel administrador

El sistema debe mostrar un panel administrativo solo a usuarios administradores.

Estado actual: Implementado.

### RF-013 - Gestion de usuarios

El sistema debe permitir a administradores listar usuarios y, en futuras iteraciones, bloquear o desbloquear cuentas.

Estado actual: Implementado parcialmente.

Pendiente relacionado: `ADMIN-002`.

### RF-014 - Gestion de recursos

El sistema debe permitir que administradores creen, editen, activen o desactiven los recursos del inventario oficial. Los ocho recursos actuales constituyen la linea base aprobada y los cambios deben reflejarse en catalogo y disponibilidad.

Estado actual: Implementado parcialmente. Administradores ya pueden actualizar la imagen del recurso; falta CRUD completo, sede, modo de reserva y activacion/desactivacion.

Pendiente relacionado: `ADMIN-003`.

### RF-015 - Bloqueos de disponibilidad

El sistema debe permitir que administradores creen bloqueos por mantencion, cierre, evento u otro motivo.

Estado actual: Pendiente.

Pendiente relacionado: `ADMIN-004`.

### RF-016 - Programacion institucional

El sistema debe permitir registrar clases, talleres, eventos, campeonatos o entrenamientos institucionales.

Estado actual: Pendiente.

Pendiente relacionado: `ADMIN-005`.

### RF-017 - Notificaciones

El sistema debe mostrar notificaciones reales del usuario.

Estado actual: Implementado parcialmente.

Pendiente relacionado: `NOTIF-001`.

### RF-018 - Reportes

El sistema debe mostrar indicadores de uso, horas punta y datos administrativos.

Estado actual: Implementado parcialmente.

Pendiente relacionado: `REP-001`, `REP-002`, `REP-003`.

### RF-019 - Talleres deportivos

El sistema debe permitir que usuarios autenticados consulten talleres activos e inscribirse cuando tengan RUT y existan cupos.

Estado actual: Implementado.

### RF-020 - Restriccion semanal de reservas particulares

El sistema debe limitar las fechas reservables al periodo institucional configurado y evitar que un usuario cree mas de una solicitud dentro del periodo contado desde la fecha local de creacion de su solicitud anterior. `PENDING` y `CONFIRMED` consumen la oportunidad; `CANCELLED` deja de consumirla. Con siete dias, un martes permite elegir hasta el lunes siguiente y una solicitud creada ese martes permite volver a solicitar desde el martes posterior.

Estado actual: APROBADO e IMPLEMENTADO; VERIFICADO LOCALMENTE el 2026-07-21. Verificacion contra SQL Server/Azure SQL y concurrencia real PENDIENTES.

Pendiente relacionado: `RES-012`.

### RF-021 - Confirmacion de participantes minimos

El sistema debe registrar confirmaciones de usuarios unicos y exigir al menos 10, incluido el solicitante, para multicancha 1, 2 y 3, identificadas en el inventario como Cancha 1, 2 y 3. Todos los participantes deben tener cuenta. Las confirmaciones pueden registrarse o retirarse hasta exactamente una hora antes inclusive, plazo configurable.

Estado actual: APROBADO el 2026-07-20; no implementado.

Pendiente relacionado: `RES-008`.

### RF-022 - Estado condicionado por politica del recurso

El sistema debe mantener `PENDING` y bloquear el horario hasta alcanzar el minimo, cambiar automaticamente a `CONFIRMED` al cumplirlo y devolverla a `PENDING` si una retirada valida reduce el conteo. Si llega al limite con menos de 10, debe cambiar a `CANCELLED`, liberar el horario y dejar de consumir la oportunidad semanal. Los recursos `OPEN_USE` no requieren confirmacion grupal.

Estado actual: APROBADO el 2026-07-20; no implementado. El flujo actual confirma todas las reservas al crearlas.

Pendiente relacionado: `RES-008`, `RES-010`.

### RF-023 - Resolucion de conflictos institucionales

El sistema debe cancelar automaticamente una reserva particular cuando entra en conflicto con una actividad institucional y notificar al usuario afectado. Si el conflicto es entre dos actividades, debe informar al administrador y permitirle cancelar cualquiera de ellas o mantener ambas.

Estado actual: APROBADO el 2026-07-20; no implementado. Las restricciones actuales rechazan el solapamiento antes de permitir la decision.

Pendiente relacionado: `ADMIN-005`, `RES-004`, `NOTIF-001`.

### RF-024 - Inventario oficial administrable

El sistema debe permitir que el administrador mantenga el inventario oficial inicial de ocho recursos en el MVP administrativo correspondiente, conservando el historial aplicable.

Estado actual: APROBADO el 2026-07-20; implementado parcialmente mediante lectura y cambio de imagen.

Pendiente relacionado: `ADMIN-003`.

### RF-025 - Configuracion administrativa de politicas de reserva

El sistema debe permitir exclusivamente a usuarios con rol administrador publicar nuevas versiones del periodo de reserva, el plazo previo de confirmacion y los recursos sujetos a confirmacion grupal. Los cambios normales son prospectivos y cada solicitud conserva la version vigente al crearse. Excepcionalmente, el administrador puede migrar solicitudes futuras `PENDING` o `CONFIRMED` seleccionadas a otra version mediante simulacion, motivo obligatorio, confirmacion, aplicacion atomica y auditoria; la operacion no edita versiones historicas ni cancela solicitudes implicitamente.

Estado actual: APROBADO e IMPLEMENTADO PARCIAL; VERIFICADO LOCALMENTE el 2026-07-21 para publicacion prospectiva de condiciones y recursos permitidos, historial administrativo, DTO publico minimo, permisos e idempotencia. La clasificacion de recursos sujetos a confirmacion grupal, participantes, minimo, transiciones, interfaz administrativa y correcciones excepcionales siguen PENDIENTES; no se verifico contra SQL Server/Azure SQL real.

Pendiente relacionado: `ADMIN-006`.

## Requisitos no funcionales

### RNF-001 - Seguridad

Las rutas protegidas deben requerir token Bearer o mecanismo local de desarrollo explicitamente habilitado.

### RNF-002 - No exposicion de secretos

Las credenciales reales no deben guardarse en archivos versionados ni en documentacion.

### RNF-003 - Trazabilidad

Las tareas completadas, parciales o nuevas deben reflejarse en `docs/07-backlog.md`.

### RNF-004 - Documentacion viva

La carpeta `docs/` debe actualizarse cuando cambien arquitectura, instalacion, base de datos, frontend, backend, flujo de reservas, requisitos o backlog.

### RNF-005 - Validacion tecnica minima

Antes de cerrar cambios relevantes se debe ejecutar, cuando corresponda:

```bash
cd backend
go test ./...

cd ../frontend
npm test
npm run build
```

Las pruebas actuales cubren reglas temporales y de agenda, pero no constituyen cobertura completa de componentes, permisos ni integracion. Si no se pueden ejecutar pruebas, debe registrarse el motivo y no declararse el requisito como verificado.

### RNF-006 - Compatibilidad local

El proyecto debe poder ejecutarse localmente con backend Go/Fiber, frontend Vue/Vite y Azure SQL Database configurada.

### RNF-007 - Usabilidad

La interfaz debe mostrar estados de carga, error y vacio en flujos principales.

### RNF-008 - Minimizacion de datos visibles

La disponibilidad para usuarios normales debe exponer solo la informacion necesaria para identificar horarios ocupados, evitando entregar datos internos o personales de reservas ajenas.

### RNF-009 - Accesibilidad basica

Los flujos criticos deben poder operarse con controles claros, mensajes de estado comprensibles y modales con comportamiento accesible basico.

### RNF-010 - Consistencia temporal

Frontend, backend y base de datos deben aplicar una unica zona horaria de negocio y un contrato explicito de serializacion. Las reglas de pasado, en curso, futuro y cancelacion no deben depender de la zona local del servidor.

Estado actual: Implementado y verificado localmente; pendiente validacion integrada/online.

### RNF-011 - Errores publicos seguros

Las respuestas HTTP deben entregar codigos y mensajes publicos estables sin incluir errores internos de SQL, JWT, JWKS o configuracion. El diagnostico tecnico debe permanecer en logs sanitizados.

Pendiente relacionado: `SEC-005`.

### RNF-012 - Responsive transversal

El shell, header, sidebar, dropdowns y modales deben mantenerse operables sin superposiciones desde 320 px y no dejar controles ocultos dentro del orden de tabulacion.

Pendiente relacionado: `BACK-021`, `BACK-022`.

## Historias de usuario

### HU-001 - Iniciar sesion

Como usuario institucional, quiero iniciar sesion con mi cuenta Microsoft para acceder al sistema de reservas.

Criterios de aceptacion:

- El usuario puede entrar con Entra ID.
- Las rutas internas requieren autenticacion.
- Si la sesion no existe, el usuario vuelve a `/login`.

### HU-002 - Completar RUT

Como usuario normal, quiero registrar mi RUT cuando el sistema lo solicite para poder crear reservas.

Criterios de aceptacion:

- Si no tengo RUT, el sistema me solicita registrarlo en un modal.
- El RUT se valida antes de guardar.
- Despues de guardar, puedo continuar al flujo solicitado.

### HU-003 - Consultar disponibilidad

Como usuario normal, quiero revisar la disponibilidad por fecha para elegir un horario libre.

Criterios de aceptacion:

- La grilla muestra recursos reales.
- La grilla muestra reservas del dia seleccionado.
- Cambiar de fecha actualiza la informacion.

### HU-004 - Crear reserva

Como usuario normal, quiero crear una reserva sobre un recurso disponible para asegurar mi horario.

Criterios de aceptacion:

- La reserva se asocia a mi usuario autenticado.
- El servidor asigna el estado segun la politica del recurso.
- Fecha, hora y duracion cumplen la jornada institucional.
- El sistema rechaza conflictos de horario.
- Si el recurso es grupal, la solicitud permanece pendiente hasta reunir 10 participantes confirmados.
- Si el recurso es `OPEN_USE`, no se exige confirmacion grupal.
- La restriccion semanal se aplica antes de aceptar una nueva solicitud.
- El sistema muestra exito o error.

### HU-005 - Cancelar reserva propia

Como usuario normal, quiero cancelar una reserva propia cuando ya no la usare.

Criterios de aceptacion:

- Solo puedo cancelar reservas propias.
- Solo puedo cancelar una reserva en estado cancelable y no finalizada.
- El sistema pide confirmacion fuerte antes de ejecutar la cancelacion.
- La lista o grilla se actualiza luego de cancelar.

### HU-006 - Revisar mis reservas

Como usuario normal, quiero ver mis reservas actuales y su detalle para hacer seguimiento.

Criterios de aceptacion:

- Se muestran reservas reales del usuario autenticado.
- Se muestra recurso, fecha, hora, duracion, actividad y estado.
- Existe acceso al detalle.

### HU-006A - Revisar reservas como administrador

Como administrador, quiero ver todas las reservas del sistema para apoyar la supervision operacional.

Criterios de aceptacion:

- Se muestran reservas reales de todos los usuarios.
- Usuarios normales siguen viendo solo sus propias reservas.
- Existe acceso al detalle de cada reserva visible.

### HU-007 - Revisar historial

Como usuario normal, quiero revisar reservas pasadas o canceladas para consultar mi actividad anterior.

Criterios de aceptacion:

- Se muestran reservas historicas reales.
- Existen estados de carga, error y vacio.
- Se puede filtrar por fecha y estado dentro del historial visible para el rol.

### HU-007A - Revisar historial global como administrador

Como administrador, quiero revisar todo el historial de reservas para auditar actividad del sistema.

Criterios de aceptacion:

- Se muestran reservas historicas reales de todos los usuarios.
- Usuarios normales siguen viendo solo su historial.
- Los filtros de estado y fecha funcionan sobre el conjunto visible segun rol.

### HU-008 - Revisar notificaciones

Como usuario normal, quiero ver notificaciones asociadas a mis reservas o eventos relevantes.

Criterios de aceptacion:

- La campana muestra contador real.
- No se consulta la API si no hay sesion.
- En una iteracion futura, podre marcar notificaciones como leidas.

### HU-009 - Acceder al panel administrador

Como administrador, quiero acceder a un panel con resumen de reservas, recursos y actividad para monitorear el sistema.

Criterios de aceptacion:

- Solo administradores ven accesos administrativos.
- Usuarios normales no ven la seccion de administracion.
- El panel carga datos reales.

### HU-010 - Gestionar usuarios

Como administrador, quiero listar usuarios y bloquear o desbloquear cuentas para controlar el acceso al sistema.

Criterios de aceptacion:

- Puedo ver usuarios reales.
- En una iteracion futura, podre bloquear o desbloquear usuarios.
- El sistema no debe permitir autobloqueo accidental.

### HU-011 - Gestionar recursos

Como administrador, quiero mantener el catalogo de recursos deportivos para que las reservas usen informacion vigente.

Criterios de aceptacion:

- Puedo actualizar la imagen de un recurso existente.
- En el MVP administrativo podre crear, editar, activar o desactivar los ocho recursos oficiales y sus futuras modificaciones.
- No se eliminan recursos con historial sin criterio definido.
- Los cambios se reflejan en disponibilidad.

### HU-012 - Crear bloqueos

Como administrador, quiero bloquear horarios por mantencion, cierre o eventos para evitar reservas no disponibles.

Criterios de aceptacion:

- Puedo seleccionar recurso, fecha y rango horario.
- El sistema valida solapamientos.
- Los bloqueos aparecen en el calendario.

### HU-013 - Consultar reportes

Como administrador, quiero revisar reportes de uso para apoyar decisiones operativas.

Criterios de aceptacion:

- Se muestran indicadores de uso.
- Se muestran horas punta.
- La definicion de reserva activa excluye finalizadas, canceladas, rechazadas y expiradas.
- Las horas se presentan sin redondeos que alteren el total real.
- En una iteracion futura, se integraran infracciones y vistas SQL dedicadas.

### HU-014 - Inscribirse en taller deportivo

Como usuario normal, quiero revisar talleres deportivos disponibles e inscribirme en uno con cupos para participar en actividades institucionales.

Criterios de aceptacion:

- Se muestran talleres activos con dia, horario, lugar, capacidad e inscritos.
- Puedo buscar talleres por nombre, dia o lugar.
- Si ya estoy inscrito, el sistema lo indica y no duplica la inscripcion.
- Si no tengo RUT, el backend rechaza la inscripcion.
- Si no hay cupos, el sistema impide la inscripcion.

### HU-015 - Reunir participantes para una reserva grupal

Como solicitante de un recurso grupal, quiero que los integrantes confirmen su participacion para que la solicitud se confirme al alcanzar el minimo institucional.

Criterios de aceptacion:

- La solicitud comienza en estado `PENDING`.
- El solicitante cuenta una vez y todos los participantes se identifican mediante una cuenta.
- La solicitud `PENDING` bloquea el horario para usos incompatibles.
- Con menos de 10 confirmaciones no se presenta como reserva confirmada.
- La decima confirmacion cambia automaticamente la solicitud a `CONFIRMED` si las demas reglas siguen vigentes.
- Si una retirada valida reduce el conteo por debajo de 10, la reserva vuelve a `PENDING`.
- Las confirmaciones y retiradas se aceptan hasta exactamente una hora antes inclusive y se rechazan despues.
- Si vence bajo el minimo, cambia a `CANCELLED`, libera el horario y la oportunidad semanal.
- El cliente no puede forzar el estado ni el conteo.

### HU-016 - Resolver conflictos institucionales

Como administrador, quiero conocer los conflictos de una actividad institucional para aplicar la prioridad institucional sin impedir usos compartidos validos.

Criterios de aceptacion:

- Un conflicto con reserva particular cancela automaticamente esa reserva y muestra el efecto al administrador.
- El usuario afectado recibe una notificacion de la cancelacion automatica.
- Un conflicto entre actividades permite cancelar cualquiera de ellas o mantener ambas.
- Mantener ambas requiere una decision administrativa explicita.
- La decision queda trazable.

### HU-017 - Mantener inventario oficial

Como administrador, quiero modificar el inventario oficial de recursos para mantener el sistema alineado con los espacios disponibles.

Criterios de aceptacion:

- Los ocho recursos actuales forman la linea base.
- El administrador puede mantener datos, modo y estado operativo en el MVP correspondiente.
- Los cambios se reflejan en catalogo y disponibilidad.
- El historial no se pierde por desactivar o modificar un recurso.

### HU-018 - Configurar politicas de reserva

Como administrador, quiero modificar el periodo de reserva, el plazo previo y los recursos sujetos a confirmacion para ajustar las reglas institucionales sin otorgar esta facultad a usuarios normales.

Criterios de aceptacion:

- Solo usuarios con rol administrador pueden modificar las politicas.
- Los valores iniciales son siete dias y una hora antes del inicio.
- Cancha 1, 2 y 3 corresponden a multicancha 1, 2 y 3 y comienzan sujetas a confirmacion grupal.
- El cambio informa desde cuando rige.
- Los cambios normales se aplican prospectivamente y las solicitudes conservan su version.
- Una correccion excepcional exige seleccionar solicitudes futuras activas, previsualizar el resultado, informar un motivo y confirmar la aplicacion.
- Si una solicitud es incompatible, el lote completo se rechaza sin cancelaciones implicitas.
- La correccion y cualquier reversion quedan auditadas.

## Casos de uso

### CU-001 - Autenticarse en el sistema

Actor principal: Usuario normal o administrador.

Precondiciones:

- El usuario tiene cuenta valida.
- La configuracion de Entra ID o modo local esta disponible.

Flujo principal:

1. El usuario abre la aplicacion.
2. El sistema muestra login.
3. El usuario inicia sesion.
4. El sistema obtiene datos del usuario local.
5. El sistema muestra la pantalla inicial segun permisos y estado del perfil.

Postcondiciones:

- El usuario queda autenticado.
- El sistema puede consumir rutas protegidas.

### CU-002 - Registrar RUT obligatorio

Actor principal: Usuario normal.

Precondiciones:

- El usuario esta autenticado.
- El usuario no tiene RUT registrado.

Flujo principal:

1. El usuario intenta entrar a una ruta interna.
2. El sistema detecta que falta RUT.
3. El sistema muestra un modal obligatorio para registrar RUT.
4. El usuario ingresa RUT en el modal.
5. El sistema valida y guarda el RUT.
6. El sistema permite continuar en la aplicacion.

Postcondiciones:

- El usuario queda habilitado para crear reservas.

### CU-003 - Crear una reserva

Actor principal: Usuario normal.

Precondiciones:

- El usuario esta autenticado.
- El usuario tiene RUT valido.
- Existen recursos disponibles.

Flujo principal:

1. El usuario abre disponibilidad.
2. El usuario selecciona fecha, recurso, hora, duracion y actividad.
3. El usuario envia la solicitud.
4. El backend valida usuario, recurso, restriccion semanal, zona horaria, jornada, duracion y conflicto horario.
5. El sistema asigna el estado segun la politica del recurso.
6. Si el recurso requiere grupo, la solicitud queda `PENDING` hasta alcanzar 10 participantes confirmados.
7. Si el recurso no requiere confirmacion grupal, se aplica directamente su comportamiento aprobado.
8. La disponibilidad se actualiza.

Postcondiciones:

- La solicitud queda asociada al usuario autenticado con el estado controlado por servidor.

Flujos alternativos:

- Si hay conflicto horario, el sistema rechaza la reserva.
- Si el horario o duracion estan fuera de regla, el sistema rechaza la reserva con mensaje accionable.
- Si un cliente intenta enviar un estado, no puede modificar el estado inicial ni evitar validaciones.
- Si falta RUT, el sistema bloquea la creacion.
- Si no se cumple la restriccion semanal, el sistema rechaza la solicitud y comunica la proxima fecha permitida.
- Si falta un dato obligatorio, el frontend muestra error.

### CU-004 - Cancelar una reserva

Actor principal: Usuario normal o administrador.

Precondiciones:

- El usuario esta autenticado.
- Existe una reserva cancelable.

Flujo principal:

1. El usuario abre el detalle de una reserva.
2. El usuario solicita cancelar.
3. El sistema pide confirmacion.
4. El backend valida permisos, estado cancelable y termino segun la zona institucional.
5. El sistema cambia el estado a cancelada.
6. La interfaz actualiza la lista o grilla.

Postcondiciones:

- La reserva queda cancelada.

Flujos alternativos:

- Si el usuario normal intenta cancelar reserva ajena, el sistema rechaza la accion.
- Si la reserva esta cancelada, rechazada, expirada o finalizada, el sistema rechaza la accion.

### CU-005 - Consultar mis reservas

Actor principal: Usuario normal.

Precondiciones:

- El usuario esta autenticado.

Flujo principal:

1. El usuario abre Mis Reservas.
2. El frontend solicita reservas del usuario autenticado.
3. El sistema muestra listado y estados.
4. El usuario abre un detalle si lo necesita.

Postcondiciones:

- El usuario visualiza sus reservas vigentes.

### CU-006 - Revisar panel administrador

Actor principal: Administrador.

Precondiciones:

- El usuario esta autenticado.
- El usuario tiene rol administrador.

Flujo principal:

1. El administrador abre el panel.
2. El sistema carga recursos y reservas.
3. El sistema muestra indicadores y accesos administrativos.

Postcondiciones:

- El administrador visualiza informacion operacional.

Flujos alternativos:

- Si un usuario normal intenta acceder, el sistema bloquea la ruta y no muestra el menu administrativo.

### CU-007 - Consultar notificaciones

Actor principal: Usuario normal o administrador.

Precondiciones:

- El usuario esta autenticado.

Flujo principal:

1. El usuario abre la aplicacion.
2. La campana consulta notificaciones del usuario.
3. El sistema muestra contador y listado.

Postcondiciones:

- El usuario visualiza sus notificaciones.

Flujos alternativos:

- Si no hay sesion, la campana no consulta la API.

### CU-008 - Inscribirse en taller

Actor principal: Usuario normal.

Precondiciones:

- El usuario esta autenticado.
- El usuario tiene RUT registrado.
- Existe al menos un taller activo con cupos.

Flujo principal:

1. El usuario abre Talleres.
2. El frontend solicita talleres activos.
3. El sistema muestra cupos, horario, lugar y estado de inscripcion.
4. El usuario solicita inscribirse.
5. El backend valida usuario, RUT, taller activo, cupo disponible e inscripcion duplicada.
6. El sistema registra la inscripcion y actualiza el taller.

Postcondiciones:

- El usuario queda inscrito en el taller seleccionado.

Flujos alternativos:

- Si falta RUT, el backend rechaza la inscripcion.
- Si el taller esta lleno, el sistema muestra error.
- Si el usuario ya estaba inscrito, el sistema evita duplicar la inscripcion.

### CU-009 - Confirmar participantes de una solicitud grupal

Actor principal: Participante.

Precondiciones:

- Existe una solicitud `PENDING` sobre un recurso grupal.
- El participante tiene cuenta, puede ser identificado y aun no ha confirmado; el solicitante ya cuenta una vez.
- La solicitud pendiente bloquea el horario.

Flujo principal:

1. El participante revisa la solicitud correspondiente.
2. El participante confirma su participacion.
3. El sistema valida identidad y duplicado.
4. El sistema actualiza el conteo de confirmaciones vigentes.
5. Al alcanzar 10 confirmaciones, vuelve a validar las demas reglas y cambia la solicitud a `CONFIRMED`.
6. Hasta exactamente el limite configurable inclusive, una persona puede retirar su confirmacion.
7. Si el conteo vigente baja de 10, el sistema devuelve la reserva a `PENDING`.
8. Si la solicitud llega al limite bajo el minimo, el sistema la cambia a `CANCELLED`, libera el horario y la oportunidad semanal.

Postcondiciones:

- La confirmacion queda registrada una sola vez.
- La reserva solo queda confirmada al alcanzar el minimo.

Flujos alternativos:

- Con menos de 10 confirmaciones, permanece `PENDING`.
- Una confirmacion o retirada exactamente una hora antes se acepta; una posterior se rechaza sin cambiar el conteo.
- Si la solicitud deja de ser valida antes de alcanzar el minimo, no se confirma automaticamente.

### CU-010 - Resolver conflicto de actividad institucional

Actor principal: Administrador.

Precondiciones:

- Existe una actividad institucional nueva o modificada.
- El mismo recurso presenta al menos una ocupacion solapada.

Flujo principal:

1. El sistema detecta y muestra todas las ocupaciones en conflicto.
2. Para cada reserva particular, el sistema aplica cancelacion automatica por prioridad institucional.
3. Para cada actividad institucional, el administrador elige cancelar una de las dos o mantener ambas.
4. El sistema registra la decision y actualiza la agenda.

Postcondiciones:

- Las reservas particulares en conflicto quedan canceladas.
- Las actividades conservadas reflejan la decision administrativa.

Flujos alternativos:

- Si el administrador cancela la nueva actividad, las otras actividades se mantienen.
- Si decide mantener ambas, el sistema conserva el solapamiento institucional de forma explicita.

### CU-011 - Configurar politicas de reserva

Actor principal: Administrador.

Precondiciones:

- El usuario tiene una cuenta vigente con rol administrador.

Flujo principal:

1. El administrador consulta las politicas institucionales vigentes.
2. Modifica el periodo de reserva, el plazo previo o los recursos sujetos a confirmacion grupal.
3. El sistema valida su rol y los nuevos valores.
4. El sistema muestra el valor anterior, el nuevo valor y desde cuando rige.
5. El administrador confirma el cambio.

Postcondiciones:

- La politica queda disponible para las solicitudes a las que corresponda.
- El cambio queda atribuible al administrador.

Flujos alternativos:

- Un usuario normal no puede modificar las politicas.
- Un valor invalido se rechaza sin alterar la politica vigente.

## Matriz de trazabilidad inicial

| Requisito | Historias relacionadas | Backlog relacionado |
| --- | --- | --- |
| RF-001 | HU-001 | AUTH-001, API-003 |
| RF-002 | HU-001 | AUTH-002 |
| RF-003 | HU-002 | AUTH-004 |
| RF-004 | HU-002, HU-004 | AUTH-004 |
| RF-005 | HU-003 | RES-003, RES-004 |
| RF-006 | HU-004, HU-015 | RES-001, RES-002, RES-005, RES-008, RES-009, RES-010, RES-011, RES-012 |
| RF-007 | HU-004, HU-012, HU-016 | RES-004, RES-010, ADMIN-004, ADMIN-005, QA-001 |
| RF-008 | HU-005 | RES-007, RES-009, RES-010, API-003 |
| RF-009 | HU-006, HU-006A | RES-006 |
| RF-010 | HU-007, HU-007A | UI-003 |
| RF-011 | HU-003, HU-011 | UI-001, API-001 |
| RF-012 | HU-009 | ADMIN-001 |
| RF-013 | HU-010 | ADMIN-002 |
| RF-014 | HU-011 | ADMIN-003 |
| RF-015 | HU-012 | ADMIN-004 |
| RF-016 | HU-012 | ADMIN-005 |
| RF-017 | HU-008 | NOTIF-001 |
| RF-018 | HU-013 | REP-001, REP-002, REP-003 |
| RF-019 | HU-014 | UI-006 |
| RF-020 | HU-004 | RES-012 |
| RF-021 | HU-015 | RES-008 |
| RF-022 | HU-004, HU-015 | RES-008, RES-010 |
| RF-023 | HU-016 | ADMIN-005, RES-004, NOTIF-001 |
| RF-024 | HU-011, HU-017 | ADMIN-003 |
| RF-025 | HU-018 | ADMIN-006 |
| RNF-008 | HU-003, HU-004 | API-004 |
| RNF-009 | HU-003, HU-004, HU-005 | UX-001, UX-002 |

## Relacion con MVPs

La agrupacion oficial de estos requisitos por incremento se mantiene en `docs/09-mvps-roadmap.md`.

Resumen:

- MVP 1: base tecnica funcional.
- MVP 2: flujo usuario completo.
- MVP 3: administracion institucional.
- MVP 4: entrega, calidad y soporte.

## Protocolo de mantenimiento

Este documento debe actualizarse cuando:

- Se agregue, quite o cambie una funcionalidad.
- Cambie el comportamiento de un actor.
- Se complete una tarea del backlog que afecte requisitos.
- Se agregue una pantalla, endpoint o regla de negocio.
- Se detecte un nuevo caso de uso o una excepcion relevante.

Cada cambio funcional debe mantener coherencia entre:

- `docs/07-backlog.md`
- `docs/08-requisitos-historias-casos-uso.md`
- `docs/09-mvps-roadmap.md`
- Documentos tecnicos afectados dentro de `docs/`
