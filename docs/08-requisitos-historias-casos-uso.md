# Poli-REDI - Requisitos, historias de usuario y casos de uso

## Objetivo del documento

Este documento consolida los requisitos funcionales, requisitos no funcionales, historias de usuario y casos de uso principales de Poli-REDI.

Debe mantenerse actualizado junto con el backlog, la arquitectura, el flujo de reservas y la documentacion tecnica cada vez que cambie el alcance funcional del sistema.

## Alcance del sistema

Poli-REDI es un sistema web para gestionar reservas deportivas institucionales. Permite a usuarios autenticados consultar disponibilidad, registrar reservas, cancelar reservas propias, revisar historial y recibir notificaciones. Los administradores pueden revisar informacion operacional, usuarios, recursos, reportes y, en iteraciones futuras, gestionar bloqueos, recursos, programacion institucional e infracciones.

## Actores

### Usuario normal

Persona autenticada que usa recursos deportivos institucionales.

Responsabilidades principales:

- Consultar disponibilidad.
- Crear reservas propias.
- Registrar o actualizar RUT cuando corresponda.
- Revisar sus reservas e historial.
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

- Guardar usuarios, recursos, actividades, reservas, notificaciones y datos administrativos.
- Aplicar reglas de integridad y restricciones criticas.
- Proveer vistas de apoyo para reportes y calendario.

## Requisitos funcionales

### RF-001 - Autenticacion de usuarios

El sistema debe permitir ingreso mediante Microsoft Entra ID y, en ambiente de desarrollo, mediante autenticacion local controlada por variables de entorno.

Estado actual: Implementado.

### RF-002 - Sincronizacion de usuario autenticado

El sistema debe obtener o crear el usuario local asociado a la identidad autenticada.

Estado actual: Implementado parcialmente.

Pendiente relacionado: `AUTH-002`.

### RF-003 - Registro y validacion de RUT

El sistema debe permitir que usuarios normales registren o actualicen su RUT, validando formato y digito verificador.

Estado actual: Implementado.

### RF-004 - Bloqueo de reserva sin RUT

El sistema debe impedir que usuarios normales sin RUT creen reservas.

Estado actual: Implementado.

### RF-005 - Consulta de disponibilidad

El sistema debe mostrar disponibilidad de recursos por fecha, considerando reservas existentes.

Estado actual: Implementado parcialmente.

Pendiente relacionado: `RES-004`.

### RF-006 - Creacion de reservas

El sistema debe permitir crear reservas sobre recursos disponibles, asociadas al usuario autenticado y, opcionalmente, a una actividad.

Estado actual: Implementado.

### RF-007 - Validacion de conflictos de reserva

El sistema debe rechazar reservas que se solapen con otras reservas confirmadas o reglas de disponibilidad.

Estado actual: Implementado para reservas.

Pendiente relacionado: `RES-004`, `ADMIN-004`.

### RF-008 - Cancelacion de reservas

El sistema debe permitir cancelar reservas propias y permitir que administradores cancelen reservas segun permisos.

Estado actual: Implementado.

### RF-009 - Mis reservas

El sistema debe mostrar las reservas del usuario autenticado con fecha, recurso, actividad, duracion y estado.

Estado actual: Implementado.

### RF-010 - Historial

El sistema debe mostrar reservas pasadas, canceladas o finalizadas del usuario.

Estado actual: Implementado parcialmente.

Pendiente relacionado: `UI-003`.

### RF-011 - Catalogo de recursos

El sistema debe mostrar recursos deportivos con datos reales.

Estado actual: Implementado parcialmente.

Pendiente relacionado: `UI-001`.

### RF-012 - Panel administrador

El sistema debe mostrar un panel administrativo solo a usuarios administradores.

Estado actual: Implementado.

### RF-013 - Gestion de usuarios

El sistema debe permitir a administradores listar usuarios y, en futuras iteraciones, bloquear o desbloquear cuentas.

Estado actual: Implementado parcialmente.

Pendiente relacionado: `ADMIN-002`.

### RF-014 - Gestion de recursos

El sistema debe permitir crear, editar, activar o desactivar recursos deportivos.

Estado actual: Pendiente.

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

Pendiente relacionado: `REP-001`, `REP-002`.

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
npm run build
```

Si no se pueden ejecutar pruebas, debe registrarse el motivo.

### RNF-006 - Compatibilidad local

El proyecto debe poder ejecutarse localmente con backend Go/Fiber, frontend Vue/Vite y Azure SQL Database configurada.

### RNF-007 - Usabilidad

La interfaz debe mostrar estados de carga, error y vacio en flujos principales.

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

- Si no tengo RUT, el sistema me dirige a configuracion.
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
- El sistema rechaza conflictos de horario.
- El sistema muestra exito o error.

### HU-005 - Cancelar reserva propia

Como usuario normal, quiero cancelar una reserva propia cuando ya no la usare.

Criterios de aceptacion:

- Solo puedo cancelar reservas propias.
- El sistema pide confirmacion.
- La lista o grilla se actualiza luego de cancelar.

### HU-006 - Revisar mis reservas

Como usuario normal, quiero ver mis reservas actuales y su detalle para hacer seguimiento.

Criterios de aceptacion:

- Se muestran reservas reales del usuario autenticado.
- Se muestra recurso, fecha, hora, duracion, actividad y estado.
- Existe acceso al detalle.

### HU-007 - Revisar historial

Como usuario normal, quiero revisar reservas pasadas o canceladas para consultar mi actividad anterior.

Criterios de aceptacion:

- Se muestran reservas historicas reales.
- Existen estados de carga, error y vacio.
- En una iteracion futura, se podra filtrar por fecha y estado.

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

- Puedo crear, editar, activar o desactivar recursos.
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
- En una iteracion futura, se integraran infracciones y vistas SQL dedicadas.

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
3. El sistema redirige a configuracion.
4. El usuario ingresa RUT.
5. El sistema valida y guarda el RUT.
6. El sistema redirige al destino original.

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
3. El usuario confirma la reserva.
4. El backend valida usuario, recurso y conflicto horario.
5. El sistema crea la reserva.
6. La disponibilidad se actualiza.

Postcondiciones:

- La reserva queda asociada al usuario autenticado.

Flujos alternativos:

- Si hay conflicto horario, el sistema rechaza la reserva.
- Si falta RUT, el sistema bloquea la creacion.
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
4. El backend valida permisos.
5. El sistema cambia el estado a cancelada.
6. La interfaz actualiza la lista o grilla.

Postcondiciones:

- La reserva queda cancelada.

Flujos alternativos:

- Si el usuario normal intenta cancelar reserva ajena, el sistema rechaza la accion.

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

## Matriz de trazabilidad inicial

| Requisito | Historias relacionadas | Backlog relacionado |
| --- | --- | --- |
| RF-001 | HU-001 | AUTH-001, API-003 |
| RF-002 | HU-001 | AUTH-002 |
| RF-003 | HU-002 | AUTH-004 |
| RF-004 | HU-002, HU-004 | AUTH-004 |
| RF-005 | HU-003 | RES-003, RES-004 |
| RF-006 | HU-004 | RES-001, RES-002, RES-005 |
| RF-007 | HU-004, HU-012 | RES-004, ADMIN-004, QA-001 |
| RF-008 | HU-005 | RES-007, API-003 |
| RF-009 | HU-006 | RES-006 |
| RF-010 | HU-007 | UI-003 |
| RF-011 | HU-003, HU-011 | UI-001, API-001 |
| RF-012 | HU-009 | ADMIN-001 |
| RF-013 | HU-010 | ADMIN-002 |
| RF-014 | HU-011 | ADMIN-003 |
| RF-015 | HU-012 | ADMIN-004 |
| RF-016 | HU-012 | ADMIN-005 |
| RF-017 | HU-008 | NOTIF-001 |
| RF-018 | HU-013 | REP-001, REP-002 |

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
- Documentos tecnicos afectados dentro de `docs/`
