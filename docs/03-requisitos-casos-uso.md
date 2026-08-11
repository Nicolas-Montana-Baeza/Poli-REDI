# Poli-REDI - Requisitos, Historias de Usuario, Casos de Uso y Roadmap MVP

> **Revisión UX 2026-07-30:** los criterios comunes de detalle, accesibilidad y
> privacidad se aplican sin ampliar permisos.

> **Fecha de consolidación:** 2026-07-23  
> **Propósito:** Agrupar el catálogo funcional, historias de usuario, casos de uso formales y la hoja de ruta de incrementos de software (MVP 1 a MVP 4).

---

## 1. Roadmap por Incrementos (MVPs)

```mermaid
gantt
    title Roadmap Poli-REDI
    dateFormat  YYYY-MM-DD
    section MVP 1 Base Operativa
    Consulta Disponibilidad y Reservas Básicas :done, mvp1, 2026-01-01, 2026-03-31
    section MVP 2 Reglas y Grupal
    Quorum 10 Personas y Frecuencia Semanal   :active, mvp2, 2026-04-01, 2026-06-30
    section MVP 3 Administración
    Bloqueos e Infracciones                   :active, mvp3, 2026-07-01, 2026-08-31
    section MVP 4 Trazabilidad
    Reportes Avanzados                        :mvp4, 2026-09-01, 2026-10-31
```

* **MVP 1 (Base Operativa):** Demo funcional. Inicios de sesión, disponibilidad en agenda, reservas simples, historial básico de reservas propias o participadas, captura de RUT y panel admin inicial.
* **MVP 2 (Reglas y Flujo Grupal):** `ACCEPTED LOCALLY`. Frecuencia semanal, mínimo y objetivo, código grupal cifrado, `/join`, participación, deadline y expiración. Como ampliación controlada incorpora la consulta de talleres e inscripciones propias. Migración 004 y concurrencia real en Azure SQL pendientes. La prioridad institucional no forma parte del cierre aceptado.
* **MVP 3 (Administración Extendida):** Parcial. Bloqueo con motivos, notificaciones internas, pantalla informativa pública e historial institucional de clases, actividades programadas y otros eventos. La participación personal en estos eventos requerirá una relación explícita futura.
* **MVP 4 (Trazabilidad y Futuro):** En desarrollo/Trabajo futuro. Reportes avanzados e integración con sistemas académicos institucionales.

---

## 2. Historias de Usuario (HU)

### Criterios UX comunes

1. Consultar una reserva desde Disponibilidad, Mis Reservas, Historial o un
   código utiliza el mismo detalle visual y semántico.
2. Cada acción se habilita por capacidades explícitas del actor y estado.
3. En Mis Reservas, toda la tarjeta abre el detalle mediante puntero, Enter o
   Espacio; no depende de un botón interno.
4. El diálogo gestiona foco, Escape y devuelve el foco al elemento que lo abrió.
5. El código se obtiene bajo demanda; no se incluye en listados ni se expone a
   terceros o en estados terminales.
6. La agenda personal rechaza solapes reales por reservas propias o
   participaciones confirmadas; los extremos contiguos se permiten.
7. Un skeleton solo sustituye contenido durante la primera carga y cuando no
   existen datos. En un refresh se conservan los datos con un indicador
   discreto; en una mutación se usa un spinner local.
8. Los skeletons conservan la geometría de la superficie, reservan 16:9 para
   medios, exponen el estado de carga mediante una región accesible y respetan
   la preferencia de movimiento reducido.
9. Si Historial obtiene una de sus fuentes y falla la otra, muestra los datos
   disponibles junto con una advertencia parcial.
10. Join y los modales que ya poseen el objeto de detalle no reemplazan el
    contenido por skeleton. Las acciones de confirmar, retirar, cancelar,
    copiar, rotar o guardar mantienen el contexto visible.
11. Cada bloque comunica por separado su tipo u origen y su estado. Los tipos
    admitidos son Reserva, Reserva grupal, Uso libre, Taller, Clase,
    Entrenamiento, Campeonato, Evento e Institucional; un chip de tipo no
    sustituye los estados Pendiente, Confirmada, Cancelada o Programada.
12. La leyenda `Tipos de bloque` y los chips usan el mismo helper de
    clasificación en Por recurso y Agenda del día. `OPEN_USE` mantiene el
    heatmap y explica textualmente que la intensidad representa reservas
    simultáneas.
13. La información de tipo y estado debe entenderse sin depender solo del
    color. Los chips no reciben foco y cada bloque interactivo expone un nombre
    accesible completo con tipo, título seguro, estado, horario y acción.
14. Un usuario normal que consulta una reserva ajena solo ve ocupación: título
    `Reserva`, horario, recurso y el tipo grupal seguro si corresponde. No ve
    PII, actividad, participantes, mínimo, objetivo, capacidad ni plazo.
15. La reserva propia conserva su información y acciones autorizadas; el
    administrador recibe detalle operacional conforme a su rol. Una actividad
    programada conserva su categoría mediante `activityType`, con
    `Institucional` como alternativa genérica.
16. Este incremento no agrega bloqueos futuros entre talleres, clases,
    entrenamientos, campeonatos, eventos u otras actividades institucionales.

* **HU-01 (Estudiante - Consultar Disponibilidad):** *Como estudiante, quiero consultar la disponibilidad de las canchas por fecha y hora para saber cuándo puedo reservar.*
* **HU-02 (Estudiante - Crear Reserva Particular):** *Como estudiante, quiero reservar una cancha ingresando mis datos para asegurar un bloque horario de juego.*
* **HU-03 (Estudiante - Flujo Grupal):** *Como estudiante organizador, quiero invitar a 9 compañeros mediante un código de unión para alcanzar el quorum mínimo requerido.*
* **HU-04 (Administrador - Cancelación Excepcional):** *Como administrador, quiero cancelar una reserva particular ante un evento institucional prioritario para reasignar la cancha.*
* **HU-05 (Estudiante - Gestionar Inscripción a Talleres):** *Como estudiante, quiero inscribirme y cancelar mi propia inscripción a talleres deportivos para administrar mi participación sin perder su trazabilidad.*
* **HU-07 (Estudiante - Consultar Actividad Personal):** *Como estudiante, quiero consultar mis reservas e inscripciones a talleres para reconocer mi actividad registrada sin que el sistema infiera asistencia a clases o eventos.*
* **HU-08 (Administrador - Consultar Historial Institucional):** *Como administrador, quiero consultar clases, actividades programadas y otros eventos históricos para revisar la operación institucional.*

### Criterios de aceptación del historial ampliado

1. En MVP 1, el usuario consulta reservas propias o reservas en las que figura
   como participante confirmado, pasadas o canceladas.
2. La ampliación de MVP 2 muestra solo las inscripciones a talleres del usuario
   autenticado; un usuario no puede consultar inscripciones ajenas.
3. Una inscripción a taller se identifica como tal y no se presenta como reserva
   ni como evidencia de asistencia.
4. En MVP 3, las clases, actividades programadas y otros eventos se muestran en
   el historial institucional conforme a los permisos del rol.
5. Una clase o evento no aparece como actividad personal hasta que exista una
   relación explícita usuario–actividad; la agenda por sí sola no demuestra
   participación.
6. Cada elemento informa su tipo, título, ubicación o recurso, fecha y horario
   aplicables y estado semántico, sin mezclar estados incompatibles entre
   reservas, talleres y eventos.
7. Los filtros permiten distinguir tipo, estado y rango de fechas, mantienen
   orden cronológico descendente y usan la zona horaria `America/Santiago`.
8. Los detalles y autorizaciones se resuelven según el dominio del elemento; no
   se reutiliza obligatoriamente el detalle de reserva para talleres o eventos.
9. Los estados inactivos o cancelados que formen parte de la trazabilidad siguen
   siendo consultables por el actor autorizado.
10. Una inscripción de taller `CANCELLED` aparece como `Inscripción cancelada`,
    deja de consumir cupo y no bloquea solapes; una reinscripción posterior se
    registra como un episodio nuevo `CONFIRMED`.
* **HU-06 (Perfil - Registrar RUT):** *Como usuario normal, quiero registrar una vez mi RUT para habilitar operaciones personales sin que el dato pueda ser reemplazado posteriormente.*

---

## 3. Casos de Uso Formales (CU)

### CU-01: Solicitar Reserva Particular
* **Actor:** Estudiante Autenticado.
* **Precondición:** El usuario debe tener su RUT registrado y cumplir la frecuencia configurada para solicitudes normales activas. `OPEN_USE` no consume frecuencia, pero tampoco permite solapes activos del mismo usuario.
* **Flujo Principal:**
  1. El estudiante ingresa a la vista `/disponibilidad`.
  2. Selecciona la cancha, fecha y horario deseado.
  3. El sistema valida disponibilidad y restricción semanal en servidor.
  4. Si la cancha requiere grupo, el sistema crea la reserva en estado `PENDING`; el propietario recupera bajo demanda el código cifrado y puede rotarlo.
  5. Los participantes ingresan manualmente o mediante `/join/:code`, consultan progreso y pueden confirmar, retirar o reconfirmar antes del deadline inclusivo.
  6. Al alcanzar el mínimo, la reserva pasa automáticamente a `CONFIRMED`; si vence bajo el mínimo, cambia a `CANCELLED`.

### CU-02: Gestionar Conflicto Institucional (Administrador)
* **Actor:** Administrador.
* **Precondición:** Existen reservas particulares `CONFIRMED` en un horario donde se programará una clase/EFI.
* **Flujo Principal:**
  1. El administrador ingresa la actividad institucional en la agenda.
  2. El sistema detecta el solapamiento de horarios.
  3. El servidor cancela automáticamente la reserva particular afectada.
  4. El sistema emite una notificación interna explicativa al alumno desasociado.

### CU-03: Gestionar una Inscripción a Taller
* **Actor:** Usuario autenticado.
* **Precondición:** Usuario normal con RUT válido, o administrador; taller activo con ocurrencias válidas y cupo.
* **Flujo Principal:**
  1. El sistema normaliza las ocurrencias de todos los días del taller.
  2. Compara solo talleres activos en los que el usuario mantiene una inscripción `CONFIRMED`.
  3. Usa intervalos semiabiertos; los horarios contiguos se permiten.
  4. Si no existe solape, crea la inscripción (`201`) o devuelve la inscripción ya vigente de forma idempotente (`200`).
* **Alternativa:** Si existe solape, responde `409` con código y detalle estructurado del taller en conflicto. Inscripciones `CANCELLED` y talleres inactivos no bloquean.
* **Cancelación propia:**
  1. El usuario autenticado solicita `DELETE /api/workshops/:id/enrollment`.
  2. El sistema identifica al actor por la sesión y busca únicamente su episodio
     `CONFIRMED`; no admite indicar ni retirar a un tercero.
  3. No se exige RUT y no se aplica corte horario mientras el taller carezca de
     período formal.
  4. Si el taller está activo, el episodio cambia a `CANCELLED`, libera cupo,
     deja de bloquear solapes y registra `WORKSHOP_ENROLLMENT_CANCELLED`.
  5. Repetir la operación es idempotente. Si el taller está inactivo, responde
     `409` con `WORKSHOP_ENROLLMENT_CLOSED`.
  6. Una reinscripción posterior crea un episodio nuevo `CONFIRMED`, registra
     `WORKSHOP_ENROLLMENT_CREATED` y no reactiva el episodio cancelado.
* **Resultado:** El historial conserva ambos episodios y presenta el cancelado
  como `Inscripción cancelada`, sin convertirlo en evidencia de asistencia.

### CU-04: Registrar RUT
* **Actor:** Usuario normal autenticado.
* **Precondición:** `/api/me` cargado y ausencia de RUT válido.
* **Resultado:** El RUT queda normalizado, único y write-once. Repetir el mismo valor es idempotente; cambiarlo o duplicarlo responde `409`. Administradores no reciben el modal.
