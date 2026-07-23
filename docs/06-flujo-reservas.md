# Poli-REDI - Flujo de reservas

## Objetivo del documento

Este documento describe el flujo funcional de reservas en Poli-REDI, sus validaciones principales, puntos de experiencia de usuario y pruebas recomendadas.

## Actores

- Usuario normal: consulta disponibilidad, crea reservas propias, cancela reservas propias y revisa historial.
- Administrador: consulta informacion operacional y puede cancelar reservas segun permisos.
- Backend: valida autenticacion, usuario, RUT, recurso, permisos y estado.
- Azure SQL Database: aplica restricciones de integridad y reglas de conflicto.

## Flujo actual de creacion de reserva

1. El usuario inicia sesion.
2. El frontend carga usuario actual, recursos, reservas y actividades.
3. El usuario abre disponibilidad.
4. El sistema muestra recursos, reservas, talleres y actividades institucionales del dia seleccionado.
5. El usuario selecciona un horario libre.
6. Si el usuario normal no tiene RUT, la UI bloquea la accion.
7. El formulario permite ajustar recurso, fecha, hora, duracion, actividad y, para recursos grupales, un objetivo opcional entre el minimo y la capacidad.
8. El frontend valida campos obligatorios.
9. El backend toma el usuario desde la sesion autenticada.
10. El backend rechaza usuarios normales sin RUT.
11. El servicio valida zona horaria de negocio, fecha/hora, duracion permitida, paso de inicio, jornada operativa, ventana reservable y frecuencia desde la creacion de la solicitud activa anterior.
12. La base de datos vuelve a proteger ventana y frecuencia ante concurrencia, ademas de validar conflictos de recurso, usuario, bloqueos y actividades programadas.
13. Si el recurso pertenece al conjunto grupal de la politica, el backend crea la solicitud `PENDING`, registra al propietario como participante confirmado y devuelve un `joinCode` de una sola exposicion; en otro caso crea `CONFIRMED` sin codigo.
14. Si la reserva se crea, la UI refresca disponibilidad y muestra mensaje de exito.
15. Si existe conflicto, el formulario mantiene el error visible.

El backend y la base del flujo grupal estan implementados y verificados localmente. El frontend cubre seleccion del objetivo, progreso agregado y edicion por propietario, pero aun no muestra/conserva/permite compartir `joinCode`, ingresar por codigo, confirmar ni retirar. No fue verificado en Azure SQL; el deadline de confirmaciones/retiros y el vencimiento automatico siguen pendientes.

## Flujo objetivo aprobado de solicitud y confirmacion

1. El usuario autenticado y con RUT selecciona recurso, fecha, hora y duracion.
2. El servidor valida que la fecha elegida este dentro de la ventana configurable; con el valor vigente de siete dias, un martes se puede elegir desde ese martes hasta el lunes siguiente.
3. El servidor valida la frecuencia configurable desde la fecha de creacion de la solicitud anterior. Una solicitud `PENDING` consume la oportunidad desde que se crea; una solicitud `CANCELLED` deja de consumirla.
4. El sistema consulta la politica del recurso.
5. Si el recurso es `OPEN_USE`, no exige confirmacion grupal ni consume la
   frecuencia de solicitudes. Puede reservarse consecutivamente, pero se
   rechaza si el mismo usuario ya tiene cualquier reserva activa solapada.
6. Para multicancha 1, 2 y 3, identificadas en el inventario como Cancha 1, 2 y 3, registra la solicitud como `PENDING` y bloquea el horario.
7. El solicitante cuenta dentro del minimo. El objetivo opcional usa el minimo por defecto, no puede superar la capacidad y limita nuevas altas; alcanzar el minimo, no el objetivo, confirma la solicitud.
8. El propietario puede cambiar el objetivo hasta el limite configurable inclusive, sin bajarlo del minimo ni del conteo confirmado y sin superar la capacidad.
9. Con menos del minimo de confirmaciones vigentes, la solicitud permanece pendiente.
10. Al alcanzar el minimo, el sistema cambia automaticamente la solicitud a `CONFIRMED`.
11. Si una persona retira su confirmacion dentro del plazo y el conteo baja del minimo, vuelve a `PENDING`.
12. Si al alcanzar el limite permanece bajo el minimo, debe cancelarse automaticamente; este ultimo paso sigue pendiente.

Solo un usuario con rol administrador puede modificar los recursos sujetos a confirmacion, el periodo semanal o el plazo previo. Los cambios se aplican prospectivamente a solicitudes creadas despues de la publicacion.

## Flujo objetivo aprobado de prioridad institucional

1. El administrador registra o modifica una actividad institucional.
2. El sistema detecta ocupaciones solapadas y las presenta al administrador.
3. Si existe una reserva particular, el sistema la cancela automaticamente al confirmar la actividad institucional.
4. Si existe otra actividad institucional, el administrador puede cancelar cualquiera de las dos o mantener ambas cuando compartan validamente el espacio.
5. El sistema notifica al usuario cuya reserva particular fue cancelada automaticamente.
6. El sistema registra la decision y actualiza la agenda.

## Flujo actual de cancelacion

1. El usuario selecciona una reserva existente.
2. La UI abre el detalle de reserva.
3. La UI muestra accion de cancelacion solo si el usuario puede cancelar.
4. El frontend solicita cancelacion enviando el ID de reserva.
5. El backend obtiene el usuario autenticado desde middleware.
6. El servicio valida que el usuario sea propietario o administrador.
7. Si la reserva existe, no esta cancelada y su termino no ha pasado, cambia su estado a `CANCELLED`.
8. La UI refresca disponibilidad y muestra mensaje de exito.

Estado de integridad MVP 1:

- El servidor permite la transicion solo desde estados activos y no confia en estados enviados por el cliente (`RES-010`, en revision por despliegue).
- Todas las comparaciones de inicio/termino usan la zona de negocio definida por `APP_TIMEZONE` (`RES-009`, en revision por despliegue).

## Mejora requerida en cancelacion

La interfaz actual muestra una advertencia, pero debe agregarse una confirmacion fuerte antes de ejecutar la cancelacion.

Confirmacion recomendada:

- Titulo: `Cancelar reserva`.
- Resumen: recurso, fecha, horario y actividad.
- Acciones: `Mantener reserva` y `Cancelar reserva`.
- Mensaje: `Esta accion cambiara la reserva a estado cancelada.`

## Estado real y clasificacion temporal

Para evitar ambiguedades, Poli-REDI diferencia entre estado persistido y clasificacion temporal.

Estado real persistido:

- `PENDING`: pendiente.
- `CONFIRMED`: confirmada.
- `CANCELLED`: cancelada.
- `REJECTED`: rechazada.
- `EXPIRED`: expirada, si se usa en futuras iteraciones.

Regla de integridad MVP 1:

- El endpoint publico de creacion no acepta un estado decidido por el cliente.
- El backend asigna el estado inicial.
- Solo estados activos pueden transicionar a `CANCELLED`.

Clasificacion temporal derivada:

- Futura: la reserva aun no comienza.
- En curso: la hora actual esta dentro del rango de la reserva.
- Pasada: la reserva ya termino.

Regla UX:

- `CONFIRMED` futura se muestra como `Confirmada`.
- `CONFIRMED` en curso se muestra como `En curso`.
- `CONFIRMED` pasada se muestra como `Finalizada`.
- `CANCELLED` siempre se muestra como `Cancelada`.

`Mis Reservas` muestra reservas accionables para el usuario normal. `Historial` muestra reservas pasadas o canceladas. Estas vistas no representan estados de base de datos, sino formas de consultar la misma entidad segun contexto.

## Validaciones funcionales

### Frontend

- Recurso obligatorio.
- Fecha obligatoria.
- Hora obligatoria.
- Duracion mayor a cero.
- Bloqueo de creacion si no se pudo cargar disponibilidad.
- Bloqueo de creacion si usuario normal no tiene RUT.
- Instalaciones inactivas, informativas o no permitidas para el rol deben mostrarse deshabilitadas (`BACK-020`).

### Backend

- Usuario autenticado obligatorio.
- Usuario normal con RUT obligatorio para crear reservas.
- `resourceId` obligatorio.
- `startTime` obligatorio y parseable.
- `durationMinutes` en `30, 60, 90, 120, 150, 180`.
- Inicio entre 08:00 y 21:30, en intervalos de 15 minutos.
- Termino completo de la reserva a mas tardar a las 22:00.
- El payload publico no acepta `status`; el servidor asigna `PENDING` a recursos grupales de la politica y `CONFIRMED` al resto.
- Cancelacion solo por propietario o administrador.
- Cancelacion permitida unicamente desde `CONFIRMED` o `PENDING`.

Validaciones backend pendientes antes de cerrar MVP 1:

- Verificacion desplegada de la zona horaria institucional compartida (`RES-009`, en revision).
- Verificacion desplegada del estado inicial y transiciones controlados por servidor (`RES-010`, en revision).
- Verificacion desplegada de duraciones, paso de inicio y jornada operativa (`RES-011`, en revision).

### Base de datos

- Rechaza usuario bloqueado.
- Rechaza recurso inactivo.
- Rechaza recurso informativo.
- Rechaza recurso solo admin para usuario normal.
- Rechaza choque con reserva confirmada del mismo recurso.
- Rechaza choque con reserva confirmada del mismo usuario.
- Rechaza cruce con bloqueo activo.
- Rechaza cruce con actividad programada.

## Reglas UX recomendadas

- Mostrar siempre el rango completo de la reserva antes de confirmar.
- Redondear seleccion de horario a intervalos institucionales.
- Completar la UI para mostrar, conservar y compartir el codigo, ingresar por codigo y ejecutar confirmar/retirar; la API ya existe (`RES-008`).
- Distinguir siempre minimo institucional, objetivo de convocatoria y capacidad maxima. El mensaje de frecuencia debe aclarar que el periodo se cuenta desde la creacion de la solicitud anterior.
- Diferenciar visualmente reserva, bloqueo, mantencion y actividad institucional.
- No exponer informacion personal de reservas ajenas en disponibilidad.
- Mantener errores recuperables dentro del mismo contexto visual.

## Diagrama resumido

```mermaid
sequenceDiagram
  actor Usuario
  participant UI as Frontend
  participant API as Backend
  participant DB as Azure SQL

  Usuario->>UI: Selecciona horario disponible
  UI->>UI: Valida datos visibles
  UI->>API: POST /api/reservations
  API->>API: Obtiene usuario autenticado
  API->>API: Valida RUT y datos minimos
  API->>DB: Inserta reserva
  DB->>DB: Valida conflictos
  DB-->>API: Reserva creada o error
  API-->>UI: Resultado
  UI-->>Usuario: Exito o error accionable
```

## Pruebas recomendadas para cierre del flujo

- Crear reserva valida.
- Crear reserva sin RUT debe fallar para usuario normal.
- Elegir una fecha posterior al ultimo dia de la ventana configurable debe fallar; con siete dias, un martes permite hasta el lunes siguiente inclusive.
- Crear una nueva solicitud antes del mismo dia de la semana siguiente debe fallar; al llegar a ese dia debe permitirse, salvo otro impedimento.
- Una solicitud grupal con 9 participantes confirmados debe permanecer `PENDING`.
- La decima confirmacion valida debe cambiarla a `CONFIRMED` si las demas reglas siguen vigentes.
- Retirar una confirmacion y bajar de 10 antes del limite debe devolverla a `PENDING` sin liberar el horario.
- El solicitante cuenta una vez y no puede retirar su participacion; si desea abandonar debe cancelar la solicitud completa.
- Confirmar o retirar exactamente una hora antes debe permitirse; cualquier intento posterior debe rechazarse.
- Una solicitud que llega al limite con menos de 10 confirmaciones debe cancelarse, liberar el horario y liberar la oportunidad semanal.
- Un recurso `OPEN_USE` no debe exigir confirmaciones grupales.
- Crear reserva con conflicto debe fallar y mantener modal abierto.
- Enviar un estado manipulado no debe alterar el estado inicial ni evitar conflictos.
- Crear fuera de horario o con duracion no permitida debe fallar en backend.
- La misma reserva debe conservar hora y categoria temporal en local y Azure.
- Cancelar reserva propia debe pedir confirmacion y actualizar grilla.
- Cancelar reserva ajena debe fallar para usuario normal.
- Cancelar una reserva rechazada, expirada, cancelada o finalizada debe fallar.
- Admin puede cancelar reserva ajena.
- Cambio de fecha no debe mostrar reservas de otro dia.
- Reservas canceladas no deben bloquear disponibilidad activa.
- Una actividad institucional en conflicto debe cancelar automaticamente la reserva particular.
- Dos actividades institucionales en conflicto deben permitir al administrador cancelar una o mantener ambas.

## Cambios administrativos de politica

- Una publicacion administrativa crea un snapshot completo e inmutable con vigencia inmediata; las solicitudes existentes conservan su `policy_id`.
- La publicacion exige `Idempotency-Key`: un replay identico devuelve la misma version y uno divergente se rechaza.
- La reserva selecciona la politica dentro de la transaccion; entre operaciones concurrentes, la primera que obtiene el bloqueo determina la version aplicable.
- El usuario autenticado consulta solo condiciones operativas; el historial con identificadores, autoria y vigencias exige rol administrador.
- La correccion excepcional no esta implementada en este incremento. El diseno aprobado exige vista previa temporal de un solo uso vinculada al administrador, seleccion de solicitudes futuras activas, motivo y aplicacion atomica auditada, sin cancelaciones implicitas.

## Politica temporal vigente

Poli-REDI usa `APP_TIMEZONE=America/Santiago` como zona de negocio y un reloj inyectable en pruebas. La estrategia vigente mantiene `DATETIME2` como hora local institucional: un valor `2026-07-14 10:30:00` significa 10:30 de Chile. Requests con offset se convierten a Santiago antes de persistir; requests sin offset se interpretan directamente en la zona institucional. La API responde fechas RFC 3339 con el offset real de Chile para evitar que frontend y backend clasifiquen reservas con horas distintas.
