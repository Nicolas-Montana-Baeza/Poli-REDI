# Poli-REDI - Evolucion y trazabilidad de requisitos

Fecha de corte: 2026-08-20

## 1. Objetivo

Este documento reconstruye la evolucion funcional de Poli-REDI comparando:

1. necesidades y reglas levantadas originalmente;
2. requisitos documentados en distintos momentos;
3. decisiones de producto posteriores;
4. comportamiento implementado;
5. restricciones de persistencia;
6. pruebas automatizadas e integradas disponibles.

Su objetivo no es convertir automaticamente el codigo en requisito.

La ingenieria inversa se utiliza para detectar diferencias y reconstruir su genealogia. Una diferencia solo se convierte en requisito vigente cuando existe una decision de producto que la respalda.

## 2. Modelo de tres verdades

Para cada comportamiento se distinguen tres dimensiones:

- **A - Requisito anterior:** lo que se habia definido o aprobado previamente.
- **B - Requisito vigente:** lo que actualmente se considera correcto para el producto.
- **C - Implementacion observada:** lo que realmente ejecuta el sistema.

Clasificacion resultante:

| Relacion | Clasificacion |
| --- | --- |
| A = B = C | Requisito estable |
| B precisa A sin cambiar su objetivo | Clarificacion |
| B amplia A | Ampliacion |
| B divide A en reglas mas especificas | Descomposicion |
| B sustituye una regla anterior | Reemplazo / evolucion |
| C aporta una capacidad util no formalizada | Requisito emergente candidato |
| A = B pero C es distinto | Defecto o deuda, no evolucion |
| A contradice el alcance actual y no existe decision posterior suficiente | Decision pendiente |

## 3. Jerarquia de evidencia

### Para reconstruir comportamiento implementado

Orden preferente:

1. pruebas de integracion;
2. restricciones y modelo de persistencia;
3. servicios y reglas backend;
4. contrato API;
5. frontend;
6. documentacion de estado.

### Para declarar comportamiento requerido

Orden preferente:

1. decision de producto vigente;
2. requisito funcional o regla de negocio vigente;
3. caso de uso / historia de usuario;
4. comportamiento implementado.

Una implementacion no se considera automaticamente requisito solo por existir.

## 4. Evidencia historica utilizada

Puntos de control principales:

- `2413ab14`: catalogo previo a la gran precision funcional del 2026-07-20; contenia RF-001 a RF-019.
- `dff488f8`: cambio de intervalos de inicio de 30 a 15 minutos y fortalecimiento de reglas de agenda.
- `64219088`: refinamiento explicito de duraciones, restriccion semanal, participantes, estado grupal y notificaciones por conflicto.
- `51903041`: implementacion de administracion e historial de politicas de reserva.
- `backend/internal/services/group_reservation_integration_test.go`: evidencia de creacion grupal inicial.
- `backend/internal/repositories/participants_repository_integration_test.go`: evidencia de transiciones de participantes.
- documentacion vigente `docs/08-requisitos-historias-casos-uso.md` y `docs/13-estado-actual-producto.md`.

## 5. Matriz inicial de evolucion

| ID | Dominio | Regla anterior | Regla vigente reconstruida | Tipo | Confianza | Estado |
| --- | --- | --- | --- | --- | --- | --- |
| EV-001 | Jornada y duracion | Reserva cercana a una hora; granularidad mas rigida | Apertura 08:00, cierre 22:00, inicio cada 15 min, duraciones 30/60/90/120/150/180 | Ampliacion y precision | Alta | Vigente |
| EV-002 | Frecuencia | Una reserva por semana / espera semanal | Ventana reservable y frecuencia son reglas distintas y configurables; `PENDING` y `CONFIRMED` consumen, `CANCELLED` libera | Clarificacion y descomposicion | Alta | Vigente |
| EV-003 | Participantes | Minimo de 10 personas | Diez cuentas unicas; solicitante incluido; persistencia individual; plazo; capacidad y codigo de invitacion | Descomposicion y fortalecimiento | Alta | Vigente parcial |
| EV-004 | Estado grupal | Una confirmada que baja de 10 vuelve a `PENDING` | Una vez confirmada conserva `CONFIRMED`; bajo el minimo pasa a `AT_RISK` | Reemplazo | Alta | Vigente |
| EV-005 | Prioridad institucional | EFI/administracion prioriza o cancela | Conflicto institucional-particular cancela la reserva y notifica; actividad-actividad requiere decision administrativa | Refinamiento | Alta | Aprobado pendiente de implementacion completa |
| EV-006 | Politicas de reserva | Reglas tratadas inicialmente como parametros practicamente fijos | Versiones publicables por administrador, vigencia prospectiva, snapshot por solicitud, historial e idempotencia | Generalizacion | Alta | Vigente parcial |
| EV-007 | Inventario | Recursos identificados como espacios disponibles | Ocho recursos forman linea base oficial y deben poder mantenerse sin perder historial | Formalizacion | Alta | Vigente parcial |
| EV-008 | Tiempo institucional | Zona horaria implicita | Contrato explicito `America/Santiago` compartido entre capas | RNF emergente | Alta | Vigente |
| EV-009 | Politica por tipo de recurso | Tratamiento relativamente uniforme de recursos | `RESERVABLE`, `OPEN_USE`, `INFORMATIVE` y `ADMIN_ONLY` tienen reglas diferentes | Especializacion | Media-alta | Vigente; origen historico exacto por completar |
| EV-010 | Prioridad institucional | Reserva particular en conflicto se cancela automaticamente | El backend actual registra un conflicto y habilita resolucion administrativa `KEEP`/`ALLOW`/`CANCEL`/`RESCHEDULE` | Candidato de reemplazo/refinamiento | Alta evidencia tecnica; decision de producto pendiente | Candidato |
| EV-011 | Motor de persistencia | PostgreSQL inicial -> Azure SQL Database | PostgreSQL 16 + `pgx` + migraciones `PG16_*` | Reevaluacion y reemplazo arquitectonico | Alta | Vigente |

## 6. Evoluciones confirmadas

### EV-001 - Jornada, duracion y granularidad

#### Antes

El levantamiento describia reservas cercanas a una hora. Durante la implementacion inicial tambien existio una granularidad de 30 minutos.

#### Evolucion

El dominio mostro que una reserva no necesita durar exactamente una hora.

Se consolido:

- apertura 08:00;
- inicio anterior a 22:00;
- termino maximo exactamente 22:00;
- inicios en pasos de 15 minutos;
- duraciones de 30 a 180 minutos en incrementos de 30.

#### Motivo

Aumentar flexibilidad sin perder una agenda discreta y validable.

#### Afecta

- RF-006;
- RN-005;
- reglas frontend/backend;
- pruebas de agenda;
- politica administrable.

### EV-002 - De "una reserva por semana" a dos politicas temporales

#### Antes

La regla se expresaba de forma resumida como una reserva por semana o espera de siete dias.

#### Evolucion

Se separaron dos conceptos:

1. **ventana reservable:** hasta que fecha futura puede elegir el usuario;
2. **frecuencia de solicitud:** desde cuando puede crear una nueva solicitud.

Con la configuracion inicial ambos utilizan siete dias, pero no representan la misma regla.

Ademas:

- `PENDING` consume la oportunidad;
- `CONFIRMED` mantiene el consumo;
- `CANCELLED` libera la oportunidad;
- el servidor comunica la proxima fecha permitida;
- el periodo es configurable.

#### Tipo

Clarificacion + descomposicion.

### EV-003 - De "10 participantes" a identidad grupal persistida

#### Antes

La regla principal era reunir al menos diez participantes.

#### Evolucion

El concepto de participante paso a ser una entidad verificable:

- cada participante corresponde a una cuenta autenticada;
- no se permiten duplicados;
- el solicitante cuenta una vez;
- existe owner de la solicitud;
- existe persistencia de participantes;
- existe limite temporal de cambios;
- se respeta capacidad del recurso;
- existe join code;
- el codigo almacenado no necesita conservarse en texto plano;
- el owner y administracion pueden consultar informacion ampliada segun permisos.

#### Tipo

Descomposicion + fortalecimiento de identidad y seguridad.

### EV-004 - Estado persistido vs condicion operacional

#### Version 1 - 2026-07-20

`PENDING -> CONFIRMED -> PENDING`

Una retirada bajo el minimo hacia regresar la reserva a pendiente.

#### Version 2 - 2026-08-20

Se separan dos dimensiones:

`reservation.status`

- `PENDING`
- `CONFIRMED`
- `CANCELLED`
- `REJECTED`
- `EXPIRED`

`groupCondition`

- `PENDING_MINIMUM`
- `HEALTHY`
- `AT_RISK`
- `INACTIVE`

Cadena principal:

`PENDING + PENDING_MINIMUM`
-> minimo alcanzado
-> `CONFIRMED + HEALTHY`
-> perdida posterior del minimo
-> `CONFIRMED + AT_RISK`

Si recupera el minimo:

`CONFIRMED + AT_RISK`
-> `CONFIRMED + HEALTHY`

#### Motivo

Una reserva que ya alcanzo el minimo fue efectivamente confirmada. Perder integrantes despues constituye una condicion de riesgo y no borra ese hecho historico.

Esta separacion tambien genera eventos de dominio utilizables por notificaciones.

### EV-005 - Prioridad institucional operacionalizada

#### Antes

La institucion tenia prioridad sobre reservas particulares y administracion podia intervenir.

#### Evolucion

Se definieron dos conflictos diferentes:

**Actividad institucional vs reserva particular**

- prevalece la actividad;
- se cancela la reserva particular;
- se informa el efecto;
- se notifica al usuario afectado.

**Actividad institucional vs actividad institucional**

- el sistema informa el conflicto;
- administracion puede cancelar una;
- o mantener ambas cuando el uso compartido sea valido;
- la decision debe quedar trazable.

#### Tipo

Refinamiento del significado de prioridad.

### EV-006 - Politicas versionadas

#### Antes

Ventanas, duraciones, plazos y recursos sujetos se comportaban conceptualmente como reglas globales.

#### Evolucion

La regla paso a ser un objeto institucional administrable:

- politica vigente;
- historial de politicas;
- vigencia temporal;
- duraciones permitidas;
- recursos sujetos;
- periodo reservable;
- frecuencia;
- minimo;
- plazo de confirmacion;
- jornada;
- intervalo;
- publicacion autorizada;
- idempotencia;
- snapshot/version asociada a la solicitud.

#### Consecuencia

Los requisitos pueden cambiar sin reinterpretar retroactivamente cada reserva historica.

Esta evolucion es especialmente importante para la estrategia incremental de MVP.

### EV-007 - Inventario oficial

#### Antes

Los recursos eran principalmente datos de catalogo.

#### Evolucion

Los ocho recursos actuales pasan a constituir una linea base oficial administrable.

El objetivo deja de ser solamente "mostrar recursos" y pasa a incluir:

- mantener datos;
- modo;
- estado operativo;
- activacion/desactivacion;
- preservacion del historial.

### EV-008 - Zona horaria institucional

#### Antes

El reloj podia depender implicitamente del ambiente de ejecucion.

#### Evolucion

`America/Santiago` pasa a ser una regla transversal explicita.

Frontend, backend y persistencia deben interpretar consistentemente:

- hora institucional;
- pasado/presente/futuro;
- cancelacion;
- ventanas;
- horario de verano.

#### Tipo

Requisito no funcional emergente descubierto durante implementacion/despliegue.

### EV-009 - Modos de recurso

El dominio deja de asumir que todos los recursos se reservan igual.

Se consolidan comportamientos diferenciados:

- `RESERVABLE`;
- `OPEN_USE`;
- `INFORMATIVE`;
- `ADMIN_ONLY`.

Ejemplo:

`OPEN_USE` admite concurrencia y no participa del flujo de confirmacion grupal.

El origen exacto de esta evolucion debe continuar rastreandose en commits anteriores a la linea base del 2026-07-20.

### EV-011 - Evolucion del motor de persistencia

La persistencia de Poli-REDI ha pasado por tres etapas verificables.

#### Version A - PostgreSQL inicial

El backend inicial fue creado con PostgreSQL.

Evidencia:

- `76876d11` - 2026-05-23
- `feat: initialize backend with Fiber framework and PostgreSQL connection`

#### Version B - Azure SQL Database

El 2026-07-03 se realizo una migracion deliberada hacia Azure SQL Database.

Evidencia:

- `ae4837f2` - 2026-07-03
- `Refactor database implementation to migrate from PostgreSQL to Azure SQL Database`

Esta etapa introdujo:

- `go-mssqldb`;
- T-SQL;
- esquema `dbo`;
- patrones de concurrencia SQL Server;
- documentacion orientada a Azure SQL.

#### Version C - PostgreSQL 16 vigente

En agosto se reevaluo la arquitectura de persistencia.

Evidencia:

- `1c1f4561` - 2026-08-14
- `feat(database): initialize MVP1 baseline schema and seed data`

y:

- `7bb8c4e9` - 2026-08-17
- `merge: integrar MVP1 estable con PostgreSQL`

El runtime vigente:

- usa `github.com/jackc/pgx/v5`;
- abre la conexion con driver `pgx`;
- acepta `postgres://` y `postgresql://`;
- utiliza variables `PG*`;
- posee migraciones PostgreSQL MVP1 y MVP2;
- posee pruebas de integracion PostgreSQL.

#### Clasificacion

Tipo: **reevaluacion y reemplazo arquitectonico**.

La etapa Azure SQL no desaparece de la historia del proyecto. Se conserva como version arquitectonica previa.

#### Fuente de verdad vigente

- `backend/internal/database/database.go`;
- `backend/.env.example`;
- `database/README.md`;
- `database/postgres/`.

#### Deudas detectadas

La migracion no esta cerrada completamente:

1. talleres legacy y `scheduled_activities_repository.go` fueron retirados tras comprobar que la infraestructura institucional PostgreSQL los reemplaza;
2. `go-mssqldb` fue retirado como dependencia directa;
3. notificaciones fueron migradas a PostgreSQL mediante `PG16_0009_full_notifications.sql`;
4. lectura, historial y publicacion administrativa de politicas utilizan PostgreSQL;
5. el instalador Quadlet se actualiza para inicializar el esquema vigente `PG16_0001` a `PG16_0009`;
6. los scripts T-SQL se conservan unicamente como legado historico.

El backend compilable ya no contiene consultas SQL Server activas.

Estas deudas no cambian el motor vigente: PostgreSQL 16.


#### Regla documental para referencias Azure SQL

Las menciones posteriores a la formalizacion de EV-011 deben clasificarse asi:

- **Vigente incorrecto:** una frase afirma que Azure SQL es el motor actual -> debe corregirse.
- **Historia correcta:** commit, QA, checklist o backlog de julio de 2026 -> se conserva y se etiqueta como historico cuando exista riesgo de confusion.
- **Deuda correcta:** descripcion de T-SQL, `go-mssqldb`, `dbo`, `UPDLOCK/HOLDLOCK` u otro resto ejecutable -> se conserva hasta completar la migracion.
- **Despliegue historico:** configuracion de la antigua demo Azure SQL -> no debe presentarse como receta vigente para el backend PostgreSQL.

No se realiza reemplazo masivo de la palabra `Azure`; Azure continua siendo relevante para Entra ID, Static Web Apps/App Service historicos o futuros y otros servicios que no dependen del motor SQL.

## 7. Evoluciones de interaccion

Estas decisiones son relevantes y deben conservar trazabilidad, aunque no necesariamente requieren un RF independiente.

| ID | Decision | Clasificacion |
| --- | --- | --- |
| UXE-001 | Reservas activas e historial pasan de vistas separadas a un mismo modulo con tabs | Refinamiento UX |
| UXE-002 | El detalle de reserva se unifica en `ReservationForm` | Simplificacion de interaccion |
| UXE-003 | Cancelacion utiliza confirmacion destructiva inline y evita `window.confirm` | Fortalecimiento UX |
| UXE-004 | Carrusel de recursos pasa a desplazamiento exclusivamente manual | Decision UX |
| UXE-005 | Inicio puede transferir el recurso seleccionado a Disponibilidad | Mejora de continuidad |
| UXE-006 | Disponibilidad incorpora filtros por recurso y recursos con bloques | Mejora de descubrimiento |
| UXE-007 | Toasts de exito locales se retiran hasta disponer de un patron transversal | Decision de consistencia |

## 8. Requisitos emergentes candidatos

Estos comportamientos tienen evidencia o utilidad clara, pero no deben promoverse automaticamente a requisito vigente sin decision explicita.

### CAND-000 - EV-010: prioridad automatica vs resolucion administrativa

La regla aprobada el 2026-07-20 establece que una actividad institucional en conflicto con una reserva particular cancela automaticamente la reserva.

La implementacion actual utiliza un modelo mas general:

`ocupaciones solapadas`
-> `scheduling_conflict`
-> `scheduling_conflict_items`
-> resolucion administrativa

Las resoluciones disponibles incluyen:

- `KEEP`;
- `ALLOW`;
- `CANCEL`;
- `RESCHEDULE`.

Las pruebas de integracion demuestran:

- conflictos conectados de mas de dos elementos;
- cancelacion persistida de reservas;
- cancelacion y reprogramacion de actividades;
- coexistencia autorizada;
- rechazo atomico de planes incompatibles;
- cierre trazable del conflicto.

Esta implementacion puede representar una evolucion superior del requisito porque evita efectos destructivos automaticos y permite resolver casos institucionales complejos.

Sin embargo, no se promueve aun a requisito vigente porque cambiaria una decision de negocio aprobada explicitamente.

Decision requerida:

- mantener cancelacion automatica para reserva particular y usar resolucion administrativa solo en otros conflictos; o
- adoptar resolucion administrativa como comportamiento general y versionar RF-023.

Estado: **candidato de evolucion de alta prioridad**.

### CAND-001 - Eventos grupales para notificaciones

`AT_RISK`, recuperacion del minimo y cancelacion por vencimiento permiten generar eventos como:

- grupo bajo minimo;
- grupo recuperado;
- proximidad del limite;
- cancelacion por minimo insuficiente.

Encaja naturalmente con MVP posteriores de notificaciones.

Estado: candidato fuerte.

### CAND-002 - Autenticacion real y despliegue como alcance

La documentacion academica original situa autenticacion institucional real y despliegue productivo fuera del alcance, mientras el prototipo incorpora Entra ID y demo Azure.

Esto no debe considerarse automaticamente una evolucion aprobada.

Clasificacion: conflicto de alcance pendiente de decision.

### CAND-003 - Filtros persistentes en URL

La seleccion de recurso y filtro de disponibilidad ya puede representarse en la navegacion.

Podria formalizarse como requisito de recuperabilidad/compartibilidad del estado de consulta.

Estado: candidato UX.

### CAND-004 - Estado global de operaciones

La eliminacion de toasts locales muestra la necesidad de una estrategia transversal de feedback.

Podria evolucionar hacia un requisito no funcional de consistencia de notificaciones de interfaz.

Estado: candidato RNF.

## 9. Diferencias que NO deben confundirse con evolucion

Ejemplos actuales:

- un bloqueo que impide reservar pero no aparece visualmente;
- una ruta 404 no gestionada;
- llamadas de sesion redundantes;
- accesibilidad incompleta;
- vencimiento automatico aun no cerrado;
- una funcionalidad aprobada que no esta implementada.

En estos casos:

`requisito vigente != implementacion`

La clasificacion correcta es deuda, defecto o incremento pendiente.

No se debe modificar el requisito solo para hacer coincidir la documentacion con el codigo.

## 10. Versionado conceptual de requisitos

No es necesario cambiar los IDs existentes.

Se propone conservar:

`RF-022`

y registrar versiones funcionales dentro de la trazabilidad:

- `RF-022/v1` - 2026-07-20: bajo minimo posterior retorna a `PENDING`;
- `RF-022/v2` - 2026-08-20: reserva confirmada conserva `CONFIRMED` y pasa a `AT_RISK`.

El mismo modelo puede aplicarse a otras reglas cuando una decision nueva sustituya efectivamente la anterior.

## 11. Procedimiento para futuras mutaciones

Cuando un MVP produzca una posible evolucion:

1. registrar la observacion;
2. identificar requisito afectado;
3. comparar requisito anterior con comportamiento observado;
4. decidir si es bug, deuda, clarificacion o evolucion;
5. si se aprueba la evolucion, registrar motivo y fecha;
6. actualizar requisito, regla, HU/CU y criterios de aceptacion;
7. actualizar pruebas;
8. actualizar backlog y roadmap;
9. conservar la version anterior como trazabilidad, no como regla vigente.

## 12. Fosiles documentales detectados en esta auditoria

Durante la primera ingenieria inversa se detectaron referencias antiguas que sobrevivieron a la evolucion de RF-022:

- `HU-015` todavia indicaba regreso a `PENDING`;
- `CA-022` todavia esperaba regreso a `PENDING`;
- partes de `docs/13-estado-actual-producto.md` describian participantes como no implementados;
- el flujo grupal de `docs/13` conservaba la transicion antigua;
- algunas contradicciones historicas seguian describiendo que toda reserva era creada `CONFIRMED`;
- `docs/02-arquitectura.md` seguia presentando los endpoints de participantes como contratos futuros aunque la superficie MVP2 ya los expone;
- RF-025 en el estado de producto todavia clasificaba participantes, minimo y transiciones como pendientes.

Estos casos demuestran por que la trazabilidad de evolucion debe mantenerse como artefacto propio.

## 13. Segundo hallazgo: programacion institucional

La auditoria de arquitectura detecto una segunda zona con posible desfase importante.

La superficie MVP2 actual ya contiene rutas para:

- unidades institucionales;
- membresias y gestores;
- actividades institucionales;
- consulta de conflictos de programacion;
- detalle de conflicto;
- resolucion administrativa por item.

Ademas existen pruebas de integracion dedicadas a unidades, actividades, disponibilidad y calendario institucional.

Esto contradice documentacion anterior que presentaba `ADMIN-005` como una entrega arquitectonica completamente futura.

La existencia de rutas y pruebas demuestra implementacion tecnica, pero no basta por si sola para declarar cumplido RF-023 o ADMIN-005. Debe realizarse una auditoria especifica de:

1. deteccion de reserva particular vs actividad institucional;
2. cancelacion automatica de la reserva afectada;
3. generacion de notificacion al usuario;
4. actividad institucional vs actividad institucional;
5. capacidad de mantener ambas;
6. autorizacion de administradores y managers;
7. trazabilidad de la resolucion.

Clasificacion actual: **implementacion emergente detectada; cierre funcional por verificar**.

## 14. Resultado de la primera pasada

La primera auditoria concluye:

- `AT_RISK` no es la unica evolucion de requisito;
- existen varias reglas que fueron ampliadas, descompuestas o generalizadas durante los MVP;
- algunas decisiones tecnicas se transformaron en requisitos no funcionales;
- existen decisiones de UX que conviene versionar aunque no sean reglas de negocio;
- existen conflictos de alcance que requieren aprobacion y no deben resolverse mediante ingenieria inversa automatica;
- la documentacion puede conservar simultaneamente versiones distintas de una misma regla si no existe una matriz de evolucion.

Este documento pasa a ser el indice genealogico de requisitos del proyecto.

### CAND-005 - Configuracion administrativa completa de politica grupal

La ingenieria inversa de `reservation_policies` detecto que el modelo persistido
MVP2 contiene configuracion grupal que todavia no forma parte del contrato
administrativo de publicacion.

La persistencia vigente maneja:

- `reservation_policy_group_resources`;
- `late_withdrawal_minutes`;
- `group_recovery_deadline_minutes`.

Sin embargo, `PublishReservationPolicyRequest` solo permite administrar:

- ventana reservable;
- frecuencia de solicitud;
- deadline de confirmacion;
- minimo de participantes;
- jornada;
- intervalo;
- duraciones;
- recursos permitidos.

Por lo tanto, una publicacion administrativa no puede decidir explicitamente:

1. que recursos utilizan flujo grupal;
2. cuanto antes del inicio un retiro se considera tardio;
3. hasta cuando un grupo `CONFIRMED + AT_RISK` puede recuperar participantes.

#### Comportamiento transitorio seguro

Mientras producto no cierre esta decision, una nueva politica:

- hereda `late_withdrawal_minutes` desde la politica vigente;
- hereda `group_recovery_deadline_minutes` desde la politica vigente;
- hereda los recursos grupales que continuan incluidos en el nuevo scope;
- no inventa valores nuevos;
- no convierte esta configuracion en administrable implicitamente.

#### Clasificacion

Tipo: **requisito emergente por validar**.

No bloquea el cierre arquitectonico de `EV-011`, pero si bloquea declarar
completamente cerrada la administracion MVP2 de politicas.

#### Decision de producto futura

Alternativas:

1. ampliar `PublishReservationPolicyRequest` con:
   - `groupResourceIds`;
   - `lateWithdrawalMinutes`;
   - `groupRecoveryDeadlineMinutes`;

2. mover alguna de estas reglas fuera de la politica si producto determina que
   pertenece estructuralmente al recurso o a otro agregado del dominio.

