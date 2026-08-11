# Poli-REDI - Guías de Instalación, Ejecución y Despliegue en la Nube

> **Revisión operativa:** 2026-07-30. La secuencia evolutiva vigente llega hasta
> `008`.

> **Fecha de consolidación:** 2026-07-23  
> **Propósito:** Unificar las instrucciones operativas de ejecución local en desarrollo, despliegue en Microsoft Azure y el plan de transición desde Google Calendar.

---

## 1. Guía de Instalación y Ejecución Local

### Prerrequisitos:
- **Node.js** v18+ y `npm`.
- **Go** 1.20+ (compatible con `backend/go.mod`).
- Instancia activa de **Azure SQL Database** (o ejecutor de T-SQL).

### A. Configuración del Backend (Go):
1. Copiar plantilla de variables: `cp backend/.env.example backend/.env`
2. Configurar variables de base de datos y llaves cifradas:
   ```env
   PORT=3000
   DB_SERVER=poli-redi-server.database.windows.net
   DB_NAME=poli-redi-database
   DEV_AUTH_ENABLED=true
   ```
3. Configurar llaves de cifrado de códigos grupales:
   ```powershell
   ./scripts/configure-join-code-encryption.ps1
   ```
   Modos disponibles:
   ```powershell
   # Validar/reutilizar o crear configuración
   ./scripts/configure-join-code-encryption.ps1

   # Reemplazar una configuración incompleta o inválida
   ./scripts/configure-join-code-encryption.ps1 -Repair

   # Agregar una versión activa conservando las anteriores
   ./scripts/configure-join-code-encryption.ps1 -Rotate
   ```
   `-Repair` y `-Rotate` no se pueden combinar. El script valida que `.env` y sus backups estén ignorados, rechaza enlaces/junctions, escribe atómicamente y crea un backup recuperable sin mostrar claves.
   Las variables obligatorias son:
   ```env
   JOIN_CODE_ENCRYPTION_KEYS=1:<clave-base64-de-32-bytes>
   JOIN_CODE_KEY_VERSION=1
   ```
4. En una base existente, ejecutar con herramienta compatible con `GO` y backup previo:
   1. `001_mvp2_group_participants.sql`
   2. `002_mvp2_target_participants.sql`
   3. `003_open_use_frequency_scope.sql`
   4. `004_group_flow_completion.sql`
   5. `005_rut_integrity_and_admin_exemption.sql`
   6. `006_workshop_occurrences.sql`
   7. `007_repair_bootstrap_group_policy.sql`
   8. `008_personal_overlap_includes_participations.sql`

   Para `007` y `008`: crear backup/export recuperable, ejecutar cada archivo
   completo conservando `GO`, verificar todos los prechecks/postchecks y
   reejecutar para demostrar idempotencia. `007` debe detenerse si no reconoce
   inequívocamente el bootstrap; no editar políticas para forzarla. `008` debe
   probar solape en ambas direcciones y permitir extremos contiguos. Ninguna
   migra datos históricos.

   Ejecutar los prechecks y verificar el `POSTCHECK` de cada migración antes de continuar. Los 12 indicadores de `004` deben valer `1`. Antes de 005/006 crear backup: 005 se detiene ante RUT inválidos/duplicados y 006 ante catálogo activo incompleto o estructura divergente. Si una fase falla, abrir una sesión nueva, confirmar `@@TRANCOUNT = 0` y `XACT_STATE() = 0`, y seguir la recuperación de `database/migrations/README.md`; no ejecutar `drop.sql`, `schema.sql` ni `seed.sql` sobre la base única.
5. Ejecutar servidor API:
   ```bash
   cd backend
   go run ./cmd
   ```
   *Servidor escuchando en `http://localhost:3000`. Chequeo de salud: `GET /api/health`.*

### B. Configuración del Frontend (Vue 3):
1. Instalar dependencias e iniciar servidor de desarrollo:
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
   *Aplicación disponible en `http://localhost:5173`.*

---

## 2. Despliegue en Microsoft Azure (Demo Nube)

La infraestructura de demostración online utiliza:
* **Frontend:** Azure Static Web Apps (`https://purple-ground-0205c9f10.7.azurestaticapps.net`).
* **Backend:** Azure App Service con contenedor Docker (`https://poli-redi.azurewebsites.net`).
* **Base de Datos:** Azure SQL Database.
* **Autenticación:** Microsoft Entra ID (Tenant Universitario).

### Pasos de Redespliegue:
1. Asegurar que `CORS_ALLOWED_ORIGINS` en Azure App Service incluya la URL de Static Web Apps.
2. Configurar `JOIN_CODE_ENCRYPTION_KEYS` y `JOIN_CODE_KEY_VERSION` en App Service sin exponer sus valores; conservar versiones anteriores mientras existan secretos cifrados con ellas.
3. Configurar las variables `VITE_API_BASE_URL` y `VITE_ENTRA_*` en el pipeline de **GitHub Actions**.
4. Asegurar que `DEV_AUTH_ENABLED=false` en App Service y en cualquier ambiente de preproducción o demo pública.
4. Ejecutar y verificar migraciones `001` a `008`; Azure SQL real para `004` a
   `008`, idempotencia, DDL y carreras/concurrencia siguen pendientes.
5. Ejecutar push a rama principal para disparar despliegue automático.

### Recuperación de 007/008

Ante un fallo, detener el despliegue. Abrir una sesión nueva y confirmar
`@@TRANCOUNT = 0` y `XACT_STATE() = 0`; conservar la evidencia e inspeccionar sin
escribir. No ejecutar `drop.sql`, `schema.sql` ni `seed.sql`. Si no puede
demostrarse un estado compatible, restaurar el backup y escalar a Arquitectura.

### Verificación frontend del incremento asíncrono

La evidencia local del 2026-07-30 comprende:

* 18 pruebas Node aprobadas.
* `go test ./...` aprobado.
* 119 pruebas Vitest aprobadas.
* Build frontend de producción aprobado.
* `diff-check` aprobado.
* Advertencia no bloqueante: bundle frontend de 531.79 kB.

Antes del cierre de despliegue se debe ejecutar QA visual con anchos de 377,
500, 768 y 1440 px. La revisión debe confirmar geometría de skeletons, ausencia
de saltos en medios 16:9, indicadores discretos durante refresh, spinners
locales en mutaciones, advertencia parcial del Historial, navegación por teclado,
lectores de pantalla y `prefers-reduced-motion`. La optimización o división del
bundle permanece pendiente y no debe confundirse con un fallo del build.

La matriz de QA debe incluir además:

* chips y leyenda `Tipos de bloque` coherentes entre Por recurso y Agenda del
  día para Reserva, Reserva grupal, Uso libre, Taller, Clase, Entrenamiento,
  Campeonato, Evento e Institucional;
* estado visible por separado del tipo y comprensible sin depender solo del
  color;
* `OPEN_USE` representado como heatmap, con caption de intensidad y sin bloques
  individuales por asistencia;
* chips no enfocables, ausencia de focos adicionales y `aria-label` completo
  en cada bloque interactivo;
* consulta como usuario propietario, usuario ajeno y administrador. La reserva
  ajena debe mostrar `Reserva` y, si corresponde, el tipo grupal, sin PII,
  actividad, métricas ni plazo; la programación institucional debe conservar
  su categoría segura mediante `activityType`.

### Verificación de desinscripción de talleres

Antes de liberar el incremento de MVP 2 se debe comprobar:

* `DELETE /api/workshops/:id/enrollment` solo afecta la inscripción
  `CONFIRMED` del usuario autenticado y no exige RUT;
* repetir la cancelación es idempotente y no crea episodios ni auditorías
  duplicadas;
* un taller inactivo responde `409` con `WORKSHOP_ENROLLMENT_CLOSED`;
* la cancelación libera cupo, deja de bloquear solapes y aparece en Historial
  como `Inscripción cancelada`;
* una reinscripción crea un episodio nuevo `CONFIRMED`, sin reactivar la fila
  cancelada;
* la auditoría registra `WORKSHOP_ENROLLMENT_CANCELLED` y
  `WORKSHOP_ENROLLMENT_CREATED`;
* no existe retiro de terceros ni corte horario hasta definir formalmente el
  período del taller.

Evidencia del 2026-08-04: `go test ./... -count=1` aprobado en todos los
paquetes, 18 pruebas Node, 144 pruebas Vitest, build frontend de producción y
`diff-check` aprobados.

---

## 3. Plan de Transición y Corte desde Google Calendar

Para sustituir la operación manual legacy en Google Calendar por Poli-REDI:

1. **Fase 1 (Paralelo Informativo):** Mantener Google Calendar en modo lectura mientras los administradores registran los talleres e hitos en Poli-REDI.
2. **Fase 2 (Corte de Reservas Particulares):** Deshabilitar la recepción de solicitudes manuales o presenciales; canalizar el 100% de reservas de alumnos mediante la SPA Poli-REDI.
3. **Fase 3 (Integración de Pantalla Informativa):** Instalar el dashboard público de Poli-REDI en el acceso al Polideportivo, retirando los calendarios impresos.
