# Poli-REDI - Checklist de demo tecnica MVP 2

## Objetivo

Validar de forma reproducible que el flujo de usuario del MVP 2 queda operativo para una demo tecnica: disponibilidad, reserva, historial, confirmacion grupal por codigo, participacion confirmada y permisos de cancelacion coherentes.

Este checklist no reemplaza pruebas automatizadas. Sirve como prueba de humo manual para confirmar que frontend, backend, base de datos y flujo de usuario trabajan juntos con evidencia observable.

## Datos de ejecucion

| Campo | Valor |
| --- | --- |
| Fecha |  |
| Responsable |  |
| Ambiente | Local / Online |
| Frontend |  |
| Backend |  |
| Base de datos | Azure SQL / SQL Server local |
| Usuario normal usado |  |
| Usuario participante usado |  |
| Usuario admin usado |  |
| Resultado general | Pendiente / Aprobado / Con observaciones / Fallido |

## Preparacion local

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Levantar backend desde `backend/` | El servicio responde en `http://localhost:3000` | Pendiente |  |
| Levantar frontend desde `frontend/` | El frontend responde en `http://localhost:5173` | Pendiente |  |
| Confirmar variables locales | `.env` del backend y `VITE_API_BASE_URL` del frontend apuntan al ambiente correcto | Pendiente |  |
| Configurar llavero de join code | `JOIN_CODE_ENCRYPTION_KEYS` y `JOIN_CODE_KEY_VERSION` estan presentes | Pendiente |  |

## Validacion de acceso y perfil

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Iniciar sesion como usuario normal | La app carga con sesion autenticada | Pendiente |  |
| Iniciar sesion como administrador | El menu administrativo queda visible para el rol adecuado | Pendiente |  |
| Usuario normal intenta acceder a vistas admin | La ruta queda bloqueada o no visible | Pendiente |  |
| Cerrar sesion | Redirige a `/login` | Pendiente |  |

## Validacion de disponibilidad y reserva

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Abrir disponibilidad | La vista carga recursos reales y horarios disponibles | Pendiente |  |
| Seleccionar horario disponible | Se abre formulario de reserva | Pendiente |  |
| Crear reserva propia | La reserva queda persistida con estado esperado | Pendiente |  |
| Intentar reservar en horario ocupado | El backend rechaza conflicto y la UI muestra error claro | Pendiente |  |
| Revisar detalle de reserva | Muestra recurso, fecha, horario, duracion y estado | Pendiente |  |

## Validacion de historial y Mis Reservas

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Abrir `Mis Reservas` | Muestra reservas propias y las confirmadas por codigo como membresia valida | Pendiente |  |
| Abrir `Historial` | Muestra reservas pasadas, canceladas o finalizadas | Pendiente |  |
| Ver reserva confirmada por otro usuario | El propietario ve la reserva en su listado propio y el participante confirmado la ve en su perfil | Pendiente |  |

## Validacion de inscripcion y desinscripcion de talleres

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Inscribirse con RUT en taller activo con cupo | Crea un episodio `CONFIRMED`, descuenta cupo y registra `WORKSHOP_ENROLLMENT_CREATED` | Pendiente |  |
| Cancelar la inscripcion propia sin RUT | Cambia solo el episodio `CONFIRMED` propio a `CANCELLED` | Pendiente |  |
| Repetir la cancelacion | Respuesta idempotente, sin duplicar auditoria ni modificar terceros | Pendiente |  |
| Consultar taller inactivo | Responde `409 WORKSHOP_ENROLLMENT_CLOSED` y no cambia la inscripcion | Pendiente |  |
| Revisar cupo y solapes despues de cancelar | El cupo queda liberado y el episodio cancelado ya no bloquea | Pendiente |  |
| Abrir Historial | La fila permanece visible como `Inscripcion cancelada` | Pendiente |  |
| Reinscribirse | Crea un episodio nuevo `CONFIRMED`; no reactiva el cancelado | Pendiente |  |
| Intentar retirar a otro usuario | No existe parametro ni ruta que permita la operacion | Pendiente |  |
| Cancelar cerca del horario | Se permite mientras el taller este activo; no hay corte hasta definir un periodo formal | Pendiente |  |

## Validacion de flujo grupal por codigo

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Owner recupera codigo de reserva grupal | El codigo queda visible solo para el propietario | Pendiente |  |
| Owner rota el codigo | El codigo anterior queda invalido y el nuevo queda disponible | Pendiente |  |
| Usuario participante entra al codigo manualmente | El flujo de `/join` o ingreso manual funciona | Pendiente |  |
| Confirmar participacion | La reserva muestra un participante confirmado y el progreso avanza | Pendiente |  |
| Retirar confirmacion | El participante se retira y el progreso baja | Pendiente |  |
| Reconfirmar participacion | La cuenta vuelve a quedarse en el corte de participantes vigente | Pendiente |  |
| Confirmar con otro horario solapado activo | El sistema rechaza la confirmacion con `409` y mensaje comprensible | Pendiente |  |

## Validacion de permisos de cancelacion

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Owner cancela su reserva | La reserva cambia a cancelada | Pendiente |  |
| Participante confirmado intenta cancelar | La accion queda bloqueada o no visible en UI | Pendiente |  |
| Usuario ajeno intenta cancelar | La API rechaza la accion con permiso adecuado | Pendiente |  |
| Admin cancela reserva permitida | Admin puede cancelar y la accion queda consistente | Pendiente |  |

## Validacion de estados del flujo grupal

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Alcanzar el minimo de participantes | La reserva cambia a `CONFIRMED` si corresponde | Pendiente |  |
| Bajar de minimo antes del plazo | La reserva vuelve a `PENDING` | Pendiente |  |
| Llegar al limite sin alcanzar minimo | La reserva se cancela y libera oportunidad | Pendiente |  |
| Confirmar o retirar exactamente al limite | La regla de deadline inclusivo se respeta | Pendiente |  |

## Validacion de observabilidad y errores

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Forzar un error de regla de negocio | La UI muestra mensaje claro y seguro | Pendiente |  |
| Probar respuesta de error | No se exponen secretos ni detalles internos | Pendiente |  |
| Revisar consola del backend | No se observan trazas sensibles ni errores irrelevantes | Pendiente |  |

## Resultado final

| Item | Valor |
| --- | --- |
| Fecha de cierre de validacion |  |
| Resultado final | Pendiente / Aprobado / Con observaciones / Fallido |
| Observaciones principales |  |
| Bloqueadores encontrados |  |
| Acciones siguientes |  |

## Evidencia tecnica reciente

Ejecutar antes de cada demo:

```powershell
cd backend
go test ./...

cd ..\frontend
npm test
npm run build
```

| Comando | Fecha | Resultado | Observacion |
| --- | --- | --- | --- |
| `go test ./...` |  |  |  |
| `npm test` |  |  |  |
| `npm run build` |  |  |  |
