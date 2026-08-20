# Poli-REDI - Checklist de demo tecnica MVP 1

## Objetivo

Validar rapidamente que la base tecnica del MVP 1 sigue operativa antes de una presentacion o revision.

Este checklist no reemplaza pruebas automatizadas. Sirve como prueba de humo manual para confirmar que frontend, backend, Azure SQL, autenticacion y flujo minimo de reservas funcionan juntos.

## Datos de ejecucion

Completar antes de cada validacion.

| Campo | Valor |
| --- | --- |
| Fecha |  |
| Responsable |  |
| Ambiente | Local / Online |
| Frontend |  |
| Backend |  |
| Base de datos | Azure SQL |
| Usuario normal usado |  |
| Usuario admin usado |  |
| Resultado general | Pendiente / Aprobado / Con observaciones / Fallido |

## Preparacion local

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Levantar backend con Go directo desde `backend` | Backend responde en `http://localhost:3000` | Aprobado |<img width="990" height="280" alt="image" src="https://github.com/user-attachments/assets/a688c63b-d3ad-4095-a089-0691ec99b127" />|
| Levantar frontend desde `frontend` | Frontend responde en `http://localhost:5173` | Aprobado |<img width="623" height="224" alt="image" src="https://github.com/user-attachments/assets/6bcc66c4-289a-436c-bdf2-c8185a622203" />|
| Confirmar variables locales del backend | `DEV_AUTH_ENABLED` y CORS corresponden al ambiente de prueba | Aprobado |  |
| Confirmar variables locales del frontend | `VITE_API_BASE_URL` apunta al backend correcto | Aprobado |  |

## Validacion tecnica base

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Abrir `GET http://localhost:3000/api/health` | Respuesta publica saludable | Aprobado | <img width="1327" height="613" alt="image" src="https://github.com/user-attachments/assets/9e4648e4-405e-42eb-a36e-cf45e27fc919" /> |
| Abrir health online del backend | Respuesta publica saludable | Aprobado | <img width="1320" height="562" alt="image" src="https://github.com/user-attachments/assets/c3217ba8-61dc-4ed1-8357-e6790417e5ab" /> |
| Abrir frontend local | La pantalla de login carga sin errores criticos | Aprobado | <img width="1359" height="716" alt="image" src="https://github.com/user-attachments/assets/497a0d5a-7ee4-4c90-af97-5dadda2e567c" /> |
| Abrir frontend online | La pantalla de login carga sin errores criticos | Aprobado |<img width="1365" height="716" alt="image" src="https://github.com/user-attachments/assets/7b8e3ce3-9933-41c0-a0ef-f1b96b02e8e4" /> |
| Revisar consola del navegador | No hay errores bloqueantes durante carga inicial | Aprobado |  |

## Autenticacion y permisos

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Iniciar sesion como usuario normal | Entra a la aplicacion y muestra datos del usuario | Aprobado | <img width="1361" height="678" alt="image" src="https://github.com/user-attachments/assets/7ba5697d-8c37-46e3-bb92-648267474791" /> <img width="1300" height="560" alt="image" src="https://github.com/user-attachments/assets/335e0991-e930-4e95-b0ac-a3d2a096f757" />
 |
| Iniciar sesion como administrador | Entra a la aplicacion y muestra menu administrativo | Pendiente | <img width="1356" height="656" alt="image" src="https://github.com/user-attachments/assets/1a173db6-0a60-49dc-9fb9-8859e055c0a5" /> |
| Probar usuario normal contra ruta admin | No ve menu admin y no accede a vistas protegidas | Pendiente |  |
| Cerrar sesion | Redirige a `/login` | Pendiente |  |
| Si se usa modo local, probar usuario normal | Carga `/api/me` con cabeceras de prueba | Pendiente |  |
| Si se usa modo local, probar admin | Carga permisos admin solo para el usuario admin esperado | Pendiente |  |

## Datos reales

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Abrir Disponibilidad | Carga recursos reales desde Azure SQL | Pendiente |  |
| Abrir formulario de reserva | Carga actividades reales aprobadas | Pendiente |  |
| Abrir Recursos | Lista instalaciones reales | Pendiente |  |
| Abrir Dashboard | Muestra informacion calculada desde API/stores | Pendiente |  |
| Abrir Mis Reservas como usuario normal | Muestra solo reservas propias accionables | Pendiente |  |
| Abrir Historial como usuario normal | Muestra solo historial propio | Pendiente |  |
| Abrir Reservas como admin | Muestra reservas globales | Pendiente |  |
| Abrir Historial como admin | Muestra historial global | Pendiente |  |

## Flujo minimo de reserva

Usar una fecha y recurso de prueba. Si la demo online apunta a datos reales, coordinar antes para evitar afectar operacion.

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Usuario normal con RUT selecciona horario disponible | Se abre formulario de reserva | Pendiente |  |
| Usuario normal sin RUT intenta reservar | Se solicita RUT antes de continuar | Pendiente |  |
| Crear reserva de prueba | La reserva queda confirmada o en el estado esperado | Pendiente |  |
| Intentar crear reserva en horario ocupado | Backend rechaza conflicto y UI muestra error claro | Pendiente |  |
| Abrir detalle de reserva | Muestra actividad, recurso, fecha, horario, duracion y usuario | Pendiente |  |
| Admin abre detalle de reserva | Muestra nombre y RUT del usuario | Pendiente |  |
| Cancelar reserva de prueba como propietario | Cambia a cancelada y desaparece de accionables si corresponde | Pendiente |  |
| Cancelar reserva como admin | Admin puede cancelar una reserva permitida | Pendiente |  |

## Seguridad minima de demo

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Revisar que no se muestren secretos en pantalla | No aparecen passwords, tokens ni connection strings | Pendiente |  |
| Revisar consola del backend durante prueba | No imprime tokens ni datos sensibles innecesarios | Pendiente |  |
| Confirmar CORS local | Permite solo origenes esperados para desarrollo | Pendiente |  |
| Confirmar CORS online | Permite solo frontend online autorizado | Pendiente |  |
| Confirmar modo desarrollo online | `DEV_AUTH_ENABLED` no esta activo en ambiente publico | Pendiente |  |

## Regresion agregada tras revision exhaustiva

Estas validaciones son obligatorias para el cierre definitivo reabierto del MVP 1.

| Paso | Resultado esperado | Estado | Backlog |
| --- | --- | --- | --- |
| Crear una reserva con hora conocida en local y online | Inicio, termino y categoria temporal coinciden en Chile | En revision; falta prueba online | `RES-009` |
| Intentar enviar `status` manualmente al crear | El servidor lo rechaza sin alterar el estado inicial ni conflictos | En revision; falta prueba desplegada | `RES-010` |
| Crear antes de apertura, despues de cierre o con duracion no permitida | Backend rechaza cada caso con mensaje seguro | En revision; falta prueba desplegada | `RES-011` |
| Seleccionar recurso inactivo, informativo o solo admin como usuario normal | No abre formulario y explica el motivo | Pendiente | `BACK-020` |
| Revisar header, campana y sidebar en 320/360 px | No hay superposicion, recorte ni foco fuera de pantalla | Pendiente | `BACK-021` |
| Abrir/cerrar formulario y detalle con teclado | Foco entra, queda dentro, Escape cierra y foco vuelve al origen | Pendiente | `BACK-022` |
| Abrir una URL inexistente | Muestra Not Found y permite volver | Pendiente | `BACK-023` |
| Navegar entre tres vistas protegidas | No se repiten solicitudes simultaneas a `/api/me` | Pendiente | `BACK-023` |
| Abrir Historial y luego Detalle | El detalle carga y volver regresa a Historial | Verificado 2026-07-14 | `UI-002`, `UI-003` |
| Forzar un error interno controlado en ambiente local | La respuesta no expone SQL, JWT, JWKS ni `err.Error()` | Pendiente | `SEC-005` |
| Activar reduccion de movimiento | El carrusel queda usable sin animacion continua | Pendiente | `BACK-018` |

## Validaciones aprobadas para MVP 2 y MVP 3

Estas filas registran decisiones del 2026-07-20. No forman parte de la evidencia de cierre del MVP 1 y permanecen pendientes de implementacion.

| Paso | Resultado esperado | Estado | Backlog |
| --- | --- | --- | --- |
| Un martes, elegir una fecha posterior al lunes siguiente con periodo de siete dias | El servidor rechaza por quedar fuera de la ventana configurable | Pendiente | `RES-012` |
| Crear una solicitud `PENDING` un martes y volver a solicitar antes del martes siguiente | El servidor rechaza; si la primera pasa a `CANCELLED`, libera la oportunidad | Pendiente | `RES-012` |
| Confirmar 9 participantes unicos en un recurso grupal | La solicitud permanece `PENDING` | Pendiente | `RES-008` |
| Registrar la decima confirmacion valida | La solicitud cambia una sola vez a `CONFIRMED` si sigue siendo valida | Pendiente | `RES-008`, `RES-010` |
| Retirar una confirmacion vigente y bajar de 10 antes del limite | La reserva vuelve a `PENDING` | Pendiente | `RES-008`, `RES-010` |
| Confirmar o retirar exactamente una hora antes | La operacion se acepta; cualquier intento posterior se rechaza | Pendiente | `RES-008` |
| Llegar al limite con 9 confirmaciones | La solicitud se cancela y libera horario y oportunidad semanal | Pendiente | `RES-008`, `RES-012` |
| Intentar confirmar con una identidad sin cuenta | La operacion se rechaza | Pendiente | `RES-008` |
| Consultar disponibilidad mientras existe una solicitud grupal `PENDING` | El horario aparece ocupado para usos incompatibles | Pendiente | `RES-008`, `API-004` |
| Usar un recurso `OPEN_USE` | No se exige confirmacion grupal | Pendiente | `RES-008` |
| Programar actividad institucional sobre reserva particular | La reserva se cancela, el administrador ve el efecto y el usuario recibe una notificacion | Pendiente | `ADMIN-005`, `NOTIF-001` |
| Programar actividad sobre otra actividad | El administrador puede cancelar una o mantener ambas | Pendiente | `ADMIN-005` |
| Modificar un recurso oficial como administrador | Catalogo y disponibilidad reflejan el cambio sin perder historial | Pendiente | `ADMIN-003` |
| Modificar periodo, plazo o recurso sujeto a confirmacion | Solo un administrador puede hacerlo y el sistema informa desde cuando rige | Pendiente | `ADMIN-006` |

## Resultado final

| Item | Valor |
| --- | --- |
| Fecha de cierre de validacion |  |
| Resultado final | Pendiente / Aprobado / Con observaciones / Fallido |
| Observaciones principales |  |
| Bloqueadores encontrados |  |
| Acciones siguientes |  |

## Evidencia tecnica reciente

Registrar comandos ejecutados antes de una demo.

```powershell
cd backend
go test ./...

cd ..\frontend
npm test
npm run build
```

| Comando | Fecha | Resultado | Observacion |
| --- | --- | --- | --- |
| `go test ./...` | 2026-07-14 | Aprobado local | Incluye pruebas iniciales de reloj de negocio y reservas; pendiente ampliar `QA-001` |
| `npm test` | 2026-07-14 | Aprobado local | Tres casos de zona horaria: invierno, verano y cruce de medianoche |
| `npm run build` | 2026-07-14 | Aprobado | Vite completa build de produccion |
| `go test ./...` | 2026-07-20 | Aprobado local | Pruebas en reloj, JSON, reglas de horario y servicio de reservas; no incluye Azure SQL integrada |
| `npm test` | 2026-07-20 | Aprobado local | 9 pruebas aprobadas de zona horaria y reglas de agenda |
| `npm run build` | 2026-07-20 | Aprobado | Vite 8 completa build de produccion |
| `npm test` | 2026-08-20 | Aprobado local | 25 pruebas frontend aprobadas |
| `npm run build` | 2026-08-20 | Aprobado | Build de produccion completado |
| `git diff --check` | 2026-08-20 | Aprobado | Sin errores de whitespace en el bloque validado |

La evidencia automatizada acumulada no sustituye la validacion de navegacion real, responsive, accesibilidad, Microsoft Entra ID, base de datos integrada ni despliegue online.
