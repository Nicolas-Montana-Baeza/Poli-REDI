# Calidad, pruebas y checklists de Poli-REDI

**Estado:** CANÓNICO PARA EVIDENCIA

## 1. Principio

Una prueba automatizada local no certifica Azure, Microsoft Entra ID ni experiencia visual real. Cada evidencia debe registrar fecha, ambiente, comando, resultado y observaciones.

## 2. Evidencia automatizada más reciente documentada

| Verificación | Resultado |
|---|---|
| Backend | `go test ./... -count=1` aprobado. |
| Node | 18 pruebas aprobadas. |
| Vitest | 144 pruebas aprobadas. |
| Build frontend | Aprobado. |
| `diff-check` | Aprobado. |
| Talleres — desinscripción | Incremento verificado localmente el 2026-08-04. |

## 3. Pruebas de base de datos pendientes

Para `007` y `008`:

- backup recuperable;
- precheck y postcheck;
- segunda ejecución idempotente;
- prueba de recuperación;
- verificación de no retroactividad;
- reserva contra participación;
- confirmación contra reserva/participación;
- contigüidad permitida;
- concurrencia real en Azure SQL.

La migración propuesta `009` del paquete mejorado requiere su propio plan y aprobación; no está validada por las fuentes originales.

## 4. Checklist MVP 1 — humo técnico

### Preparación

- [ ] Registrar fecha, responsable, ambiente y versiones.
- [ ] Levantar backend y frontend.
- [ ] Confirmar variables, CORS y modo de autenticación.
- [ ] Verificar health local u online.

### Identidad y permisos

- [ ] Login usuario normal.
- [ ] Login administrador.
- [ ] Rechazo de vista o endpoint admin para usuario normal.
- [ ] Logout y redirección.

### Flujo base

- [ ] Recursos y actividades reales.
- [ ] Disponibilidad sin datos privados.
- [ ] Crear reserva válida.
- [ ] Rechazar conflicto.
- [ ] Cancelar reserva propia permitida.
- [ ] Consultar Mis Reservas e Historial.

## 5. Checklist MVP 2 — flujo de usuario

### Reserva grupal

- [ ] Crear solicitud grupal `PENDING`.
- [ ] Recuperar código solo como owner.
- [ ] Abrir `/join/:code`.
- [ ] Confirmar con participantes diferentes.
- [ ] Cambiar a `CONFIRMED` al alcanzar mínimo.
- [ ] Retirar y reconfirmar antes del deadline.
- [ ] Rechazar retiro del propietario.
- [ ] Rotar código y invalidar el anterior.
- [ ] Expirar bajo mínimo de forma idempotente.

### Agenda personal

- [ ] Rechazar reserva contra participación confirmada.
- [ ] Rechazar confirmación contra reserva propia.
- [ ] Rechazar confirmación contra otra participación.
- [ ] Permitir extremos contiguos.

### Talleres

- [ ] Inscribir con RUT y cupo.
- [ ] Rechazar sin RUT cuando aplica.
- [ ] Desinscribir sin RUT.
- [ ] Repetir baja y comprobar idempotencia.
- [ ] Liberar cupo y solape.
- [ ] Conservar episodio cancelado en historial.
- [ ] Reinscribir creando episodio nuevo.
- [ ] Rechazar baja en taller inactivo con `409`.

## 6. QA visual y accesibilidad pendiente

Probar 377, 500, 768 y 1440 px:

- [ ] navegación y shell;
- [ ] agenda y tarjetas;
- [ ] diálogos, foco, Escape y restauración;
- [ ] teclado con Enter y Espacio;
- [ ] `aria-busy` y anuncios;
- [ ] texto además de color;
- [ ] movimiento reducido;
- [ ] privacidad por audiencia;
- [ ] refresh y mutaciones sin reemplazar contenido por skeleton.

## 7. Plantilla de ejecución

| Campo | Valor |
|---|---|
| Fecha | |
| Responsable | |
| Ambiente | Local / Online / Copia Azure SQL |
| Commit o versión | |
| Frontend | |
| Backend | |
| Base de datos | |
| Usuario normal | |
| Usuario participante | |
| Usuario admin | |
| Resultado | Pendiente / Aprobado / Con observaciones / Fallido |

## 8. Criterio de aprobación

- **MVP 1:** flujo base integrado sin errores bloqueantes.
- **MVP 2:** checklist grupal y talleres aprobado, migraciones reales verificadas y flujo online completo.
- **No aprobado:** cualquier P0/P1, fuga de datos, permiso incorrecto, migración inexplicable o pérdida de recuperación.
