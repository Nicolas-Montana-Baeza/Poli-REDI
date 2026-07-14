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
| Levantar backend con Go directo desde `backend` | Backend responde en `http://localhost:3000` | Pendiente |  |
| Levantar frontend desde `frontend` | Frontend responde en `http://localhost:5173` | Pendiente |  |
| Confirmar variables locales del backend | `DEV_AUTH_ENABLED` y CORS corresponden al ambiente de prueba | Pendiente |  |
| Confirmar variables locales del frontend | `VITE_API_BASE_URL` apunta al backend correcto | Pendiente |  |

## Validacion tecnica base

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Abrir `GET http://localhost:3000/api/health` | Respuesta publica saludable | Pendiente |  |
| Abrir health online del backend | Respuesta publica saludable | Pendiente |  |
| Abrir frontend local | La pantalla de login carga sin errores criticos | Pendiente |  |
| Abrir frontend online | La pantalla de login carga sin errores criticos | Pendiente |  |
| Revisar consola del navegador | No hay errores bloqueantes durante carga inicial | Pendiente |  |

## Autenticacion y permisos

| Paso | Resultado esperado | Estado | Evidencia / observacion |
| --- | --- | --- | --- |
| Iniciar sesion como usuario normal | Entra a la aplicacion y muestra datos del usuario | Pendiente |  |
| Iniciar sesion como administrador | Entra a la aplicacion y muestra menu administrativo | Pendiente |  |
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
| Intentar enviar `status` manualmente al crear | El servidor lo rechaza o ignora sin alterar el estado inicial ni conflictos | Pendiente | `RES-010` |
| Crear antes de apertura, despues de cierre o con duracion no permitida | Backend rechaza cada caso con mensaje seguro | Pendiente | `RES-011` |
| Seleccionar recurso inactivo, informativo o solo admin como usuario normal | No abre formulario y explica el motivo | Pendiente | `BACK-020` |
| Revisar header, campana y sidebar en 320/360 px | No hay superposicion, recorte ni foco fuera de pantalla | Pendiente | `BACK-021` |
| Abrir/cerrar formulario y detalle con teclado | Foco entra, queda dentro, Escape cierra y foco vuelve al origen | Pendiente | `BACK-022` |
| Abrir una URL inexistente | Muestra Not Found y permite volver | Pendiente | `BACK-023` |
| Navegar entre tres vistas protegidas | No se repiten solicitudes simultaneas a `/api/me` | Pendiente | `BACK-023` |
| Abrir Historial y luego Detalle | El detalle carga y volver regresa a Historial | Verificado 2026-07-14 | `UI-002`, `UI-003` |
| Forzar un error interno controlado en ambiente local | La respuesta no expone SQL, JWT, JWKS ni `err.Error()` | Pendiente | `SEC-005` |
| Activar reduccion de movimiento | El carrusel queda usable sin animacion continua | Pendiente | `BACK-018` |

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
npm run test:run
npm run build
```

| Comando | Fecha | Resultado | Observacion |
| --- | --- | --- | --- |
| `go test ./...` | 2026-07-14 | Aprobado local | Incluye pruebas iniciales de reloj de negocio y reservas; pendiente ampliar `QA-001` |
| `npm test` | 2026-07-14 | Aprobado local | Tres casos de zona horaria: invierno, verano y cruce de medianoche |
| `npm run test:run` |  | No disponible | Pendiente configurar suite en `QA-002` |
| `npm run build` | 2026-07-14 | Aprobado | Vite completa build de produccion |
