# Poli-REDI - Guías de Instalación, Ejecución y Despliegue en la Nube

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
4. Ejecutar servidor API:
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
2. Configurar las variables `VITE_API_BASE_URL` y `VITE_ENTRA_*` en el pipeline de **GitHub Actions**.
3. Ejecutar push a rama principal para disparar despliegue automático.

---

## 3. Plan de Transición y Corte desde Google Calendar

Para sustituir la operación manual legacy en Google Calendar por Poli-REDI:

1. **Fase 1 (Paralelo Informativo):** Mantener Google Calendar en modo lectura mientras los administradores registran los talleres e hitos en Poli-REDI.
2. **Fase 2 (Corte de Reservas Particulares):** Deshabilitar la recepción de solicitudes manuales o presenciales; canalizar el 100% de reservas de alumnos mediante la SPA Poli-REDI.
3. **Fase 3 (Integración de Pantalla Informativa):** Instalar el dashboard público de Poli-REDI en el acceso al Polideportivo, retirando los calendarios impresos.
