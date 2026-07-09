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

## Sistema visual actual

Actualmente existe una base global de estilos importada desde `src/style.css`.

Los archivos `src/assets/styles/main.css` y `src/assets/styles/variables.css` ya contienen tokens y clases reutilizables para colores, superficies, radios, sombras, espacios, botones, tarjetas, estados, campos y badges. Todavia quedan estilos scoped en componentes donde el layout o la grilla lo requieren.

Conclusion:

- Existe una base global funcional.
- La aplicacion ya aplica parte de esa base en pantallas principales.
- Todavia falta validacion visual final y completar consistencia en pantallas secundarias.

Recomendacion para MVP 1:

- Mantener y ajustar tokens globales en `variables.css`.
- Seguir aplicando clases reutilizables de `main.css` en pantallas criticas.
- Mantener estilos scoped para layout especifico y comportamiento propio de cada componente.

Backlog relacionado:

- `BACK-007`: crear base global de sistema visual MVP 1.
- `BACK-008`: aplicar sistema visual global a pantallas MVP 1.
- `BACK-010`: unificar tarjetas/listado de Mis Reservas e Historial.
- `BACK-011`: unificar estados, filtros y vacios en vistas de reservas.

## Coherencia visual entre reservas e historial

Mis Reservas e Historial deben sentirse como vistas hermanas del mismo modulo. La diferencia principal es conceptual:

- Mis Reservas muestra reservas accionables o vigentes.
- Historial muestra reservas pasadas o canceladas.

La tarjeta, los badges, los estados de carga/error/vacio y el acceso al detalle deben compartir el mismo patron visual. Las acciones cambian por contexto: Mis Reservas puede mostrar `Cancelar`, mientras Historial prioriza `Ver detalle`.

Recomendacion:

- Crear un componente compartido de tarjeta de reserva.
- Usarlo en `ReservationsView.vue` y `HistoryView.vue`.
- Mantener filtros solo en Historial, pero con el mismo estilo global de formularios.
- Usar siempre `getReservationDisplayStatus` para la etiqueta visual.

## Fortalezas UX actuales

- La navegacion separa rutas publicas, autenticadas y administrativas.
- La vista de disponibilidad permite elegir fecha, recurso y horario desde una grilla visual.
- El formulario de reserva muestra validaciones por campo.
- La UI mantiene estados de carga, error y exito en flujos principales.
- Los usuarios normales sin RUT son bloqueados antes de crear reservas.
- La vista de talleres permite buscar oferta disponible e inscribirse con control de cupos.
- Las reservas existentes aparecen como bloques visuales sobre la linea de tiempo.
- La cancelacion queda limitada visualmente al propietario de la reserva o a administradores.

## Hallazgos UX de segunda pasada

### Confirmacion de acciones criticas

La cancelacion muestra una advertencia, pero no existe una confirmacion fuerte adicional antes de cambiar el estado de la reserva.

Mejora recomendada:

- Agregar dialogo de confirmacion antes de cancelar.
- Mostrar resumen completo: recurso, fecha, horario, actividad y estado.
- Usar texto directo: `Cancelar esta reserva` y `Mantener reserva`.

### Seleccion de horario

La linea de tiempo permite seleccionar cualquier minuto. Esto entrega flexibilidad, pero puede generar horarios poco institucionales y errores de seleccion fina.

Mejora recomendada:

- Redondear seleccion a bloques de 15 o 30 minutos.
- Mostrar una previsualizacion del rango antes de abrir el formulario.
- Validar visualmente si la duracion elegida se cruza con otra reserva.

### Claridad de disponibilidad

La grilla distingue reservas existentes, pero aun falta un calendario unificado para reservas, bloqueos administrativos y actividades programadas.

Mejora recomendada:

- Incorporar leyenda visual: disponible, reservado, bloqueado, mantencion, actividad institucional.
- Evitar mostrar etiquetas tecnicas como `RESERVABLE`, `ADMIN_ONLY` o `INFORMATIVE` al usuario final.
- Traducir modos de reserva a textos como `Reservable`, `Solo administracion`, `Informativo`.

### Formulario de reserva

El formulario ya valida campos obligatorios, pero puede mejorar su percepcion de seguridad y control.

Mejora recomendada:

- Mostrar capacidad del recurso cuando exista.
- Validar participantes contra capacidad.
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

## Pruebas frontend recomendadas

La configuracion actual de `frontend/package.json` no incluye suite de pruebas automatizadas. Para el MVP se recomienda agregar una capa gradual.

### Pruebas unitarias o de componentes

- Render de `AvailabilitySection` con recursos y reservas.
- Estado vacio cuando no hay recursos.
- Estado de error cuando falla disponibilidad.
- Validaciones de `ReservationForm`.
- Cierre y reapertura de modal sin conservar errores antiguos.
- Router: usuario no autenticado vuelve a `/login`.
- Router: usuario normal no entra a rutas administrativas.

### Pruebas end-to-end recomendadas

- Usuario normal inicia sesion local, registra RUT y crea reserva.
- Usuario normal sin RUT no puede reservar.
- Usuario cancela una reserva propia con confirmacion.
- Usuario con RUT se inscribe en un taller con cupos.
- Usuario sin RUT no puede inscribirse en talleres.
- Usuario normal no ve panel admin.
- Admin entra al panel y revisa usuarios/reportes.
- Conflicto de horario muestra error y mantiene el formulario abierto.

## Prioridades sugeridas

1. Confirmacion fuerte para cancelacion.
2. Disponibilidad por bloques de 15 o 30 minutos.
3. Endpoint de disponibilidad sanitizado para no exponer datos innecesarios.
4. Validacion de capacidad y participantes.
5. Suite minima de pruebas frontend.
6. Mejoras de accesibilidad en modales.
