# Checklist de cierre MVP2

Fecha de corte: 2026-09-01

Rama de trabajo: `feature/mvp2`

Estado: **EN VALIDACION MANUAL LOCAL**

## Alcance de este cierre

El cierre cubre reservas grupales, participantes autenticados, reglas por
recurso, plazo de confirmacion, solapes personales, expiracion bajo el minimo
y su integracion con PostgreSQL 16 y la interfaz existente.

No agrega administracion general, generadores de notificaciones, sanciones,
lista de espera ni nuevos actores.

## 14A - Modelo de reglas grupales

- [x] Existe `PG16_0010_mvp2_group_resource_rules.sql`.
- [x] El minimo se configura por recurso grupal.
- [x] Cada reserva grupal conserva snapshot de minimo y capacidad.
- [x] Cancha 1, Cancha 2 y Sala Multiuso admiten flujo grupal.
- [x] Los tres recursos parten con minimo 10.
- [x] El plazo vigente de confirmacion es 60 minutos.
- [x] La migracion fue aplicada a la base local existente y termino con codigo 0.
- [x] La consulta posterior mostro los tres recursos, sus minimos y capacidades.

## 14B - Participacion y transiciones

- [x] El solicitante se registra una sola vez y cuenta como participante.
- [x] El owner no puede retirar su propia participacion.
- [x] `PENDING` pasa a `CONFIRMED` al alcanzar el minimo snapshot.
- [x] Una reserva confirmada que baja del minimo conserva `CONFIRMED` y pasa a `AT_RISK`.
- [x] Al recuperar el minimo vuelve a `HEALTHY`.
- [x] El codigo utiliza el mismo confirmation deadline en `PENDING` y `CONFIRMED`.
- [x] El codigo puede utilizarse hasta el limite mientras exista capacidad.
- [x] La capacidad maxima impide nuevas incorporaciones.
- [x] No existe lista de espera.
- [x] Se impide participar en otra reserva activa horariamente solapada.
- [x] Los intervalos adyacentes no se consideran solapados.
- [x] La prueba PostgreSQL real `TestParticipantsIntegrationConfirmedToAtRisk` aprobo.

## 14C - Expiracion controlada

- [x] El housekeeping revisa solicitudes grupales antes de lecturas y escrituras relevantes.
- [x] Una reserva `PENDING` bajo el minimo al alcanzar el deadline pasa a `CANCELLED`.
- [x] La cancelacion guarda `cancellationReason=MINIMUM_NOT_MET`.
- [x] Una ejecucion repetida no vuelve a modificar la reserva.
- [x] El frontend explica el motivo en el modal y en la vista de detalle.
- [x] La prueba PostgreSQL real `TestExpirePendingGroupReservationsIntegration` aprobo.
- [ ] Verificar manualmente que el horario liberado vuelve a ser seleccionable en Disponibilidad.

## 14D - Reproducibilidad y regresion

- [x] Existe verificacion SQL MVP2 independiente.
- [x] Existe un script que crea un contenedor y volumen efimeros sin tocar la base habitual.
- [x] Ejecutar desde cero `PG16_0001` a `PG16_0010` en la base efimera.
- [x] Ejecutar toda la suite Go con `POLIREDI_INTEGRATION=1` contra esa base.
- [x] Ejecutar `go vet ./...` contra el corte.
- [x] Suite Go local sin integracion externa aprobada.
- [x] Suite frontend local aprobada: 27 pruebas.
- [x] Build Vite local aprobado: 1880 modulos.
- [x] `git diff --check` aprobado.
- [x] Registrar el resultado final de `verify-mvp2-ephemeral.sh`.

Evidencia del 2026-09-01: `CIERRE AUTOMATIZADO MVP2: PASS`. La ejecucion
incluyo la cadena limpia de PostgreSQL 16, verificacion SQL, integraciones Go,
`go vet`, 27 pruebas frontend y build Vite en un entorno Linux temporal.

Comando unico de cierre automatizado:

```bash
bash infra/local/quadlet/verify-mvp2-ephemeral.sh
```

El script elimina su contenedor, volumen y archivos temporales incluso si una
prueba falla. No modifica `poliredi-postgres-mvp1` ni su volumen.

## Prueba manual local

- [ ] Crear una reserva grupal en cada uno de los tres recursos habilitados.
- [ ] Comprobar que nace `PENDING` y muestra minimo, capacidad y codigo.
- [ ] Confirmar participantes hasta alcanzar `CONFIRMED + HEALTHY`.
- [ ] Retirar un participante no owner y observar `CONFIRMED + AT_RISK`.
- [ ] Reincorporarlo y observar `HEALTHY`.
- [ ] Intentar retirar al owner y comprobar el rechazo.
- [ ] Intentar confirmar con una reserva propia solapada y comprobar el rechazo.
- [ ] Probar un intervalo adyacente y comprobar que se permite.
- [ ] Probar el limite de capacidad.
- [ ] Consultar una cancelacion `MINIMUM_NOT_MET` y comprobar el mensaje.

## Cierre online MVP2

El desarrollo local no acredita por si solo el ambiente online decidido para
MVP2. Antes de declarar el incremento completamente entregado:

- [ ] PostgreSQL online creado desde la cadena aprobada.
- [ ] Secretos fuera del repositorio y usuario de aplicacion sin privilegios globales.
- [ ] Backend desplegado con `MVP_SCOPE=mvp2` y autenticacion de desarrollo deshabilitada.
- [ ] Frontend compilado con `VITE_MVP_SCOPE=mvp2` y URL real de la API.
- [ ] CORS, HTTPS y autenticacion institucional comprobados.
- [ ] Smoke test online de crear, unir, retirar, cancelar y expirar.
- [ ] Evidencia de logs sin datos personales ni secretos.
- [ ] Procedimiento de rollback probado o documentado.

## Criterio de salida

MVP2 queda **cerrado localmente** cuando 14D automatizado y la prueba manual
local esten completos. Queda **entregado online** solamente cuando tambien se
complete la seccion de cierre online.

No marcar el MVP2 como finalizado usando solo pruebas unitarias o una base ya
evolucionada: la cadena limpia, la experiencia manual y el ambiente online son
evidencias distintas.
