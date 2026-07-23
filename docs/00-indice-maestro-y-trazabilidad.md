# Índice Maestro de Documentación y Trazabilidad Tesis-Código

> **Proyecto:** Poli-REDI (Sistema Web para Gestión de Reservas Deportivas Institucionales)  
> **Fecha de unificación:** 2026-07-23  
> **Propósito:** Actuar como el nexo central entre los capítulos de la tesis, la documentación de diseño y la implementación en el repositorio de código.

---

## 1. Mapa Navegable de Documentación

### A. Documentación Académica y de Levantamiento (`Documentos/`)
* [01-alcance-definitivo-prototipo.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Documentos/01-alcance-definitivo-prototipo.md): Delimitación oficial de prototipo (MVP 1 + MVP 2 núcleo, MVP 3 parcial).
* [02-backlog-tesis-master.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Documentos/02-backlog-tesis-master.md): Backlog de 54 tareas académicas organizadas por capítulos con % de avance.
* [03-levantamiento-proceso-actual.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Documentos/03-levantamiento-proceso-actual.md): Diagnóstico del proceso legacy con Google Calendar y las 12 reglas de negocio originales (RN-01 a RN-12).
* [04-matriz-requisitos-integrada.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Documentos/04-matriz-requisitos-integrada.md): Catálogo unificado de Requisitos Funcionales (RF-01 a RF-22) y No Funcionales (RNF-01 a RNF-08).
* **`Entregables_Previos/`**: Entregables pasados (`.docx`, `.pptx`, cronogramas).

### B. Documentación Técnica Consolidada del Sistema (`Poli-REDI/docs/`)
* [01-resumen-y-estado-actual.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Poli-REDI/docs/01-resumen-y-estado-actual.md): Resumen ejecutivo, estado por módulo, brechas y decisiones aprobadas.
* [02-arquitectura-y-sistema.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Poli-REDI/docs/02-arquitectura-y-sistema.md): Arquitectura en capas, esquema Azure SQL Database, API Go Backend y SPA Vue 3 Frontend.
* [03-requisitos-casos-uso.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Poli-REDI/docs/03-requisitos-casos-uso.md): Roadmap por MVP, Historias de Usuario (HU) y Casos de Uso (CU).
* [04-guias-y-despliegue.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Poli-REDI/docs/04-guias-y-despliegue.md): Guías de ejecución local, redespliegue en Microsoft Azure y plan de corte desde Google Calendar.
* **`historico_y_checklists/`**: Subcarpeta con checklists de demostración y notas de auditoría histórica.

---

## 2. Matriz de Trazabilidad: Capítulos de Tesis vs. Documentación y Código Fuente

| Capítulo del Informe de Tesis | Documentos de Respaldo | Módulo de Código Principal | Pruebas / Verificación |
| :--- | :--- | :--- | :--- |
| **Capítulo 1: Introducción y Problema** | [03-levantamiento-proceso-actual.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Documentos/03-levantamiento-proceso-actual.md)<br>[01-resumen-y-estado-actual.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Poli-REDI/docs/01-resumen-y-estado-actual.md) | `README.md` | Análisis de proceso legacy y brechas en Google Calendar. |
| **Capítulo 2: Alcance y Objetivos** | [01-alcance-definitivo-prototipo.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Documentos/01-alcance-definitivo-prototipo.md)<br>[03-requisitos-casos-uso.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Poli-REDI/docs/03-requisitos-casos-uso.md) | `Poli-REDI/` | Definición de incrementos funcionales MVP 1 a MVP 4. |
| **Capítulo 3: Análisis de Requisitos y Reglas** | [04-matriz-requisitos-integrada.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Documentos/04-matriz-requisitos-integrada.md)<br>[03-requisitos-casos-uso.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Poli-REDI/docs/03-requisitos-casos-uso.md) | `backend/internal/services/` | Pruebas unitarias de reglas horarias y frecuencia semanal. |
| **Capítulo 4: Diseño de Arquitectura y Base de Datos** | [02-arquitectura-y-sistema.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Poli-REDI/docs/02-arquitectura-y-sistema.md) | `database/schema.sql`<br>`database/seed.sql` | Compilación de scripts T-SQL en Azure SQL Database. |
| **Capítulo 5: Implementación del Prototipo** | [02-arquitectura-y-sistema.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Poli-REDI/docs/02-arquitectura-y-sistema.md)<br>[04-guias-y-despliegue.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Poli-REDI/docs/04-guias-y-despliegue.md) | `backend/cmd/`<br>`frontend/src/` | Build de producción frontend (`npm run build`) y servidor Go. |
| **Capítulo 6: Pruebas y Validación** | [01-resumen-y-estado-actual.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Poli-REDI/docs/01-resumen-y-estado-actual.md)<br>`historico_y_checklists/12-checklist-demo-mvp1.md` | `backend/..._test.go`<br>`frontend/src/.../*.spec.js` | `go test ./...` y `npm test` en verde. |
| **Capítulo 7: Conclusiones y Trabajo Futuro** | [02-backlog-tesis-master.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Documentos/02-backlog-tesis-master.md)<br>[01-alcance-definitivo-prototipo.md](file:///C:/Users/Nicol%C3%A1s/Desktop/Tesis/Documentos/01-alcance-definitivo-prototipo.md) | `docs/01-resumen-y-estado-actual.md` | Matriz de evaluación de cumplimiento de objetivos. |

---

## 3. Comandos de Verificación del Sistema

Para validar la consistencia del sistema y la coherencia del prototipo funcional:

```powershell
# 1. Validación de Backend Go
cd C:\Users\Nicolás\Desktop\Tesis\Poli-REDI\backend
go test ./...

# 2. Validación de Frontend Vue
cd C:\Users\Nicolás\Desktop\Tesis\Poli-REDI\frontend
npm test
npm run build
```
