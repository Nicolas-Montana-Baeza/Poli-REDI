# Poli-REDI - Frontend, UI/UX y experiencia de reservas

## Objetivo del documento

Este documento registra la estructura frontend actual y las mejoras recomendadas para que Poli-REDI se perciba como una interfaz profesional de reservas deportivas institucionales.

El foco principal es la experiencia del usuario durante disponibilidad, reserva, cancelacion, historial y administracion inicial.

## Estado actual observado

El frontend esta construido con Vue 3, Vite, Pinia, Vue Router, Axios y MSAL Browser.

La aplicacion ya cuenta con una estructura clara:

- `src/views/`: pantallas principales.
- `src/components/availability/`: calendario, timeline, bloques de reserva y modales.
- `src/components/forms/`: formulario de reserva, selector de recurso y selector de fecha/hora.
- `src/components/layout/`: sidebar, header, menu de usuario y campana de notificaciones.
- `src/stores/`: estado global de autenticacion, reservas, recursos, actividades, talleres y notificaciones.
- `src/services/`: comunicacion con la API.
- `src/router/`: proteccion de rutas y redireccion segun rol.

### Estado sincronizado 2026-08-20

La interfaz fue revisada nuevamente durante el cierre de UX de MVP 2.

Cambios ya implementados:

- `ReservationsView.vue` concentra reservas activas e historial; el historial se selecciona mediante `?tab=history`.
- La antigua `HistoryView.vue` fue retirada.
- `ReservationForm.vue` funciona tanto en modo `create` como `detail` y actua como modal compartido para los principales flujos de reserva.
- Inicio, Disponibilidad y Reservas reutilizan el mismo patron de detalle.
- La cancelacion usa confirmacion destructiva inline dentro del propio modal; no se utiliza `window.confirm`.
- La disponibilidad incorpora filtros por recurso, opcion `Todos` y filtro para mostrar recursos con bloques disponibles.
- El carrusel de recursos del inicio es manual y horizontal; no tiene autoplay.
- Seleccionar un recurso en el carrusel dirige a Disponibilidad con el recurso precargado mediante query string.
- La autenticacion dispone de una pantalla intermedia compartida para resolver el estado inicial de sesion y para comunicar el cierre de sesion.
- El login no debe mostrarse mientras el estado de autenticacion aun es desconocido.
- Se retiro por ahora el toast global de exito para evitar mensajes inconsistentes entre acciones.
- Existe una suite automatizada frontend ejecutada mediante `npm test`; la verificacion local del 2026-08-20 completo 25 pruebas correctamente.

## Sistema visual actual

Actualmente existe una base global de estilos importada desde `src/style.css`.

Los archivos `src/assets/styles/main.css` y `src/assets/styles/variables.css` ya contienen tokens y clases reutilizables para colores, superficies, radios, sombras, espacios, botones, tarjetas, estados, campos y badges. Todavia quedan estilos scoped en componentes donde el layout o la grilla lo requieren.

Conclusion:

- Existe una base global funcional.
- La aplicacion ya aplica parte de esa base en pantallas principales.
- Los estilos scoped siguen siendo validos para layout especifico, pero aun existen constantes visuales repetidas y quiebres responsive en componentes compartidos.
- La consistencia pendiente debe resolverse por componentes transversales, no mediante un rediseño completo.

Recomendacion para MVP 1:

- Mantener y ajustar tokens globales en `variables.css`.
- Seguir aplicando clases reutilizables de `main.css` en pantallas criticas.
- Mantener estilos scoped para layout especifico y comportamiento propio de cada componente.

Backlog relacionado:

- `BACK-007`: crear base global de sistema visual MVP 1.
- `BACK-008`: aplicar sistema visual global a pantallas MVP 1.
- `BACK-010`: unificar tarjetas/listado de Mis Reservas e Historial.
- `BACK-011`: unificar estados, filtros y vacios en vistas de reservas.
- `BACK-021`: corregir shell responsive y controles globales.
- `BACK-022`: completar accesibilidad de modales y calendario.

## Coherencia visual entre reservas e historial

Reservas activas e Historial forman actualmente un unico modulo.

`ReservationsView.vue` utiliza la URL como fuente de estado:

- `/reservations`: reservas activas o accionables.
- `/reservations?tab=history`: historial.
- `/history`: redirige al tab de historial por compatibilidad.

La misma vista comparte:

- tarjeta de reserva;
- clasificacion temporal;
- estados de carga/error/vacio;
- filtros de historial;
- apertura del detalle;
- modal `ReservationForm.vue` en modo `detail`.

Las reservas activas se ordenan cronologicamente de forma ascendente. El historial se ordena de forma descendente.

La vista historica ya no mantiene una implementacion separada mediante `HistoryView.vue`, reduciendo duplicacion de comportamiento y estilos.

## Fortalezas UX actuales

- La navegacion separa rutas publicas, autenticadas y administrativas.
- La vista de disponibilidad permite elegir fecha, recurso y horario desde una grilla visual.
- Disponibilidad incorpora filtro por recurso con `Todos` como valor por defecto.
- Puede filtrarse la vista para mostrar recursos con bloques disponibles.
- Los enlaces desde el carrusel del inicio pueden precargar el recurso mediante `?resource=<id>`.
- El carrusel de instalaciones es de desplazamiento horizontal manual y no utiliza autoplay.
- El formulario de reserva muestra validaciones por campo y conserva errores recuperables dentro del modal.
- `ReservationForm.vue` concentra creacion y detalle, evitando modales divergentes para la misma entidad.
- La cancelacion exige una confirmacion inline antes de emitir la accion destructiva.
- Los usuarios normales sin RUT son bloqueados antes de crear reservas.
- La vista de talleres permite buscar oferta disponible e inscribirse con control de cupos.
- Los talleres activos se proyectan como bloques de ocupacion en la disponibilidad del recurso.
- Los recursos de uso libre se muestran como `Uso libre` y usan segmentos de intensidad en vez de bloquear completamente el horario.
- La disponibilidad consume un endpoint sanitizado, separado de la consulta administrativa de reservas.
- El catalogo y dashboard muestran imagenes de recursos cuando existen, con fallback visual cuando faltan.
- Los administradores pueden cambiar la imagen de un recurso desde la vista de recursos.
- Las reservas existentes aparecen como bloques visuales sobre la linea de tiempo.
- El detalle de reservas grupales puede mostrar participantes, condicion del grupo y gestion del codigo de invitacion cuando dichos datos son entregados por la API.
- La cancelacion queda limitada visualmente al propietario de la reserva o a administradores.
- El bootstrap de autenticacion diferencia entre sesion aun no resuelta, sesion autenticada y sesion anonima.
- La entrada y salida de sesion reutilizan una pantalla de transicion para evitar flashes del login.

## Hallazgos UX de segunda pasada

### Confirmacion de acciones criticas

Estado implementado 2026-08-20.

La cancelacion ya no utiliza la confirmacion nativa del navegador.

`ReservationForm.vue` muestra dentro del mismo modal una confirmacion destructiva con acciones equivalentes a:

- volver sin cancelar;
- confirmar explicitamente la cancelacion.

El componente de presentacion controla la confirmacion visual y la vista padre conserva la responsabilidad de ejecutar la mutacion en el store.

Este patron evita:

- `window.confirm`;
- modal sobre modal;
- confirmaciones duplicadas entre componente y vista padre.

### Seleccion de horario

La linea de tiempo permite seleccionar cualquier minuto. Esto entrega flexibilidad, pero puede generar horarios poco institucionales y errores de seleccion fina.

Mejora recomendada:

- Redondear seleccion a bloques de 15 o 30 minutos.
- Mostrar una previsualizacion del rango antes de abrir el formulario.
- Validar visualmente si la duracion elegida se cruza con otra reserva.

### Claridad de disponibilidad

La grilla distingue reservas, talleres y actividades institucionales programadas. Aun falta incorporar bloqueos administrativos al mismo calendario.

Mejora recomendada:

- Incorporar leyenda visual: disponible, reservado, bloqueado, mantencion, actividad institucional.
- Evitar mostrar etiquetas tecnicas como `RESERVABLE`, `ADMIN_ONLY` o `INFORMATIVE` al usuario final.
- Traducir modos de reserva a textos como `Reservable`, `Solo administracion`, `Informativo`.

### Formulario de reserva

El formulario ya valida campos obligatorios, pero puede mejorar su percepcion de seguridad y control.

Mejora recomendada:

- Mostrar capacidad del recurso cuando exista.
- No solicitar participantes en MVP 1. Solo volver a incorporarlo si `RES-008` persiste y valida el dato de punta a punta.
- Mostrar rango final calculado: inicio y termino.
- Confirmar que la reserva esta asociada al usuario autenticado, sin mostrar IDs internos.
- Mantener el error de conflicto dentro del modal, con una accion clara para elegir otro horario.

### Experiencia movil

La vista usa una grilla horizontal por recurso. En movil esto puede exigir bastante desplazamiento.

Mejora recomendada:

- Agregar selector de recurso arriba de la linea de tiempo en movil.
- Mostrar una sola linea de tiempo por recurso seleccionado.
- Mantener el calendario compacto y evitar que la seleccion actual ocupe demasiado alto.

### Accesibilidad ligera

Los modales y botones funcionan, pero conviene reforzar accesibilidad para una entrega mas profesional.

Mejora recomendada:

- Asegurar `aria-label` en botones iconicos.
- Cerrar modales con tecla Escape.
- Devolver foco al elemento que abrio el modal.
- Usar `aria-live` para errores y mensajes de exito.
- Evitar depender solo del color para comunicar estado.

## Hallazgos de revision exhaustiva 2026-07-14

### Shell responsive compartido

En 360 px el saludo puede invadir la campana, el dropdown de notificaciones puede salir del viewport y el sidebar cerrado conserva enlaces enfocables fuera de pantalla. La correccion se concentra en `BACK-021` para beneficiar todas las vistas.

### Elegibilidad de recursos

Los recursos inactivos, informativos o restringidos pueden parecer seleccionables hasta que backend rechaza el submit. `BACK-020` exige deshabilitarlos con motivo visible sin retirar la validacion de servidor.

### Foco y operacion por teclado

Los modales de reserva y detalle no mueven ni contienen el foco. Inputs y botones iconicos requieren nombres accesibles, y la seleccion sobre timeline depende del puntero. `BACK-022` agrupa esta correccion del flujo critico.

### Navegacion y sesion

Estado actualizado 2026-08-20.

La autenticacion incorpora un bootstrap global basado en un estado `initialized` para diferenciar una sesion aun no resuelta de una sesion realmente anonima.

`AuthLoadingScreen.vue` se reutiliza para:

- verificar la sesion al iniciar la aplicacion;
- procesar el callback institucional;
- comunicar el cierre de sesion.

El router evita renderizar `/login` hasta resolver el estado de autenticacion y mantiene una transicion especifica durante logout.

La coordinacion global redujo los flashes visuales entre login y aplicacion.

Todavia puede revisarse en una iteracion posterior la eliminacion de llamadas redundantes a `loadAuthUser()` desde vistas individuales y la estrategia definitiva para rutas 404 dentro de `BACK-023`.

### Dashboard

Estado actualizado 2026-08-20.

El carrusel principal dejo de utilizar desplazamiento automatico.

Comportamiento vigente:

- scroll horizontal manual;
- soporte para mouse, touch y trackpad;
- controles laterales cuando corresponde;
- la tarjeta completa funciona como acceso al recurso;
- al seleccionar una instalacion se navega a `/availability?resource=<id>`;
- Disponibilidad interpreta ese parametro y deja seleccionado el recurso correspondiente.

Con esto se evita movimiento forzado y se conecta directamente el descubrimiento del recurso con la consulta de disponibilidad.

### Instalacion reproducible

Una dependencia visual esta declarada en la raiz y no en `frontend/package.json`, aunque Azure compila `./frontend`. `BACK-024` debe validar `npm ci` y build desde una instalacion limpia del subproyecto.

## Pruebas frontend recomendadas

`frontend/package.json` incluye actualmente una suite automatizada ejecutada con `node --test`.

La suite cubre utilidades de tiempo de negocio, reglas de reserva, reglas de disponibilidad, foco/clasificacion de reservas y configuracion de alcance MVP.

Verificacion local registrada el 2026-08-20:

- `npm run build`: correcto.
- `npm test`: 25 pruebas correctas.
- `git diff --check`: sin errores.

La cobertura aun debe crecer hacia componentes Vue, router y pruebas end-to-end.

### Pruebas unitarias o de componentes

- Render de `AvailabilitySection` con recursos y reservas.
- Estado vacio cuando no hay recursos.
- Estado de error cuando falla disponibilidad.
- Validaciones de `ReservationForm`.
- Cierre y reapertura de modal sin conservar errores antiguos.
- Router: usuario no autenticado vuelve a `/login`.
- Router: usuario normal no entra a rutas administrativas.
- Ruta desconocida muestra Not Found.
- Dos consumidores simultaneos de sesion producen una sola llamada a `/api/me`.
- Recursos no reservables no abren el formulario.
- Helpers temporales respetan el contrato definido por `RES-009`.

### Pruebas end-to-end recomendadas

- Usuario normal inicia sesion local, registra RUT y crea reserva.
- Usuario normal sin RUT no puede reservar.
- Usuario cancela una reserva propia con confirmacion.
- Usuario con RUT se inscribe en un taller con cupos.
- Usuario sin RUT no puede inscribirse en talleres.
- Usuario normal no ve panel admin.
- Admin entra al panel y revisa usuarios/reportes.
- Conflicto de horario muestra error y mantiene el formulario abierto.
- Modal mueve, atrapa y devuelve el foco; Escape lo cierra cuando corresponde.
- Header, campana y sidebar no se superponen ni dejan foco fuera de pantalla en mobile.

## Prioridades sugeridas

1. `BACK-020`: mantener validacion preventiva para recursos no reservables antes del submit.
2. `BACK-021`: completar revision responsive del shell compartido.
3. `BACK-022`: cerrar accesibilidad de foco, Escape, labels y operacion por teclado.
4. `BACK-023`: revisar 404 y eliminar recargas de sesion redundantes que aun puedan existir en vistas individuales.
5. `BACK-024`: validar instalacion limpia y retirar codigo o dependencias frontend obsoletas.
6. `QA-002`: ampliar la regresion automatizada desde utilidades hacia componentes, router y flujos integrados.
7. Mantener `ReservationForm.vue` como unica fuente visual para creacion/detalle siempre que el contexto corresponda.
8. Mantener las notificaciones de exito fuera de la UI hasta definir un patron transversal consistente para toda la aplicacion.
