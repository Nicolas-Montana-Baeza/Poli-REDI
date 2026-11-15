# SEC-006 - Identidad estable Microsoft Entra ID

Fecha: 2026-11-15
MVP: 4
Estado: IMPLEMENTACION S1-S4 VALIDADA
Vinculacion historica: PENDIENTE

## Objetivo

Resolver la identidad institucional mediante tenant_id
y entra_oid, evitando que el correo electronico permita
apropiarse de usuarios, reservas o privilegios.

## Problema identificado

El flujo anterior buscaba usuarios por email y despues
actualizaba sus identificadores Entra.

Una identidad externa diferente podia coincidir con
el correo de una cuenta local con permisos o reservas.

## Implementacion

### S1 - Repositorio

- Resolucion por tenant_id y entra_oid.
- Validacion y normalizacion de ambos GUID.
- Alta de usuarios nuevos sin privilegios.
- Rechazo de correos vinculados a otras cuentas.
- Conservacion del identificador users.id.
- Proteccion frente a inserciones concurrentes.
- Sin vinculacion automatica por email.

### S2 - Autenticacion HTTP

- Eliminacion del flujo de sincronizacion por correo.
- Validacion de identidad local antes de autorizar.
- Conservacion de usuarios bloqueados.
- Rechazo de identidades ausentes o invalidas.
- Contratos HTTP compatibles con SEC-005.

### S3 - PostgreSQL 16

Escenarios integrados:

- Alta sin privilegios: PASS.
- Correo cambiado: PASS.
- Correo reasignado: rechazo correcto.
- Tenant diferente con correo ocupado: rechazo correcto.
- Cuenta historica administradora: protegida.
- Usuario bloqueado: estado conservado.
- Concurrencia de 16 solicitudes: PASS.
- Identidad original preservada: PASS.

Se utilizaron contenedores y volumenes efimeros.

### S4 - Tenant institucional

- ENTRA_TENANT_ID exige GUID valido.
- Se rechazan configuraciones como common.
- El tid del JWT debe coincidir con el tenant configurado.
- Un tenant ajeno se rechaza antes de consultar usuarios.
- Se mantiene validacion JWT de firma, emisor y audiencia.
- Pruebas HTTP y suite Go: PASS.

## Esquema PostgreSQL

Se reutilizan las columnas existentes:

- users.tenant_id
- users.entra_oid

El indice uq_users_entra_identity protege la
unicidad de la identidad externa.

No se requieren migraciones de esquema para S1-S4.

Las reservas, bloqueos, roles y auditorias conservan
sus referencias al identificador users.id.

## Vinculacion de cuentas historicas

Estado: PROCEDIMIENTO DEFINIDO, EJECUCION PENDIENTE.

No existe vinculacion automatica por coincidencia de
correo, nombre o RUT.

Procedimiento institucional requerido:

1. Identificar la cuenta historica de destino.
2. Confirmar su identidad con evidencia institucional
   independiente del correo declarado en el JWT.
3. Obtener tenant_id y oid desde una fuente Entra
   autenticada y autorizada.
4. Verificar que la identidad externa no este vinculada
   a otra cuenta local.
5. Verificar que la cuenta de destino no tenga una
   identidad externa distinta.
6. Obtener autorizacion administrativa documentada,
   especialmente para cuentas con privilegios.
7. Realizar la vinculacion mediante una operacion
   transaccional, condicionada a que ambos campos
   de identidad continuen sin vincular.
8. Verificar que users.id, RUT, roles, reservas y
   estado de bloqueo permanezcan inalterados.
9. Registrar evidencia de la operacion y responsable.
10. Probar el acceso y mantener un procedimiento
    de recuperacion ante vinculación incorrecta.

No realizar actualizaciones masivas por correo.

Este documento no autoriza ejecutar cambios sobre
datos institucionales ni sustituye las aprobaciones
requeridas para vincular cuentas reales.

## Limitaciones y riesgos residuales

- Las cuentas historicas sin tenant/OID vinculados
  pueden recibir rechazo de acceso hasta completar
  el proceso supervisado.
- El correo de una identidad ya vinculada no se
  actualiza automaticamente al cambiar el claim.
- El modo DEV_AUTH_ENABLED sigue utilizando correo
  para pruebas locales; debe estar deshabilitado
  en cualquier despliegue institucional.
- No se ha verificado autenticacion contra Microsoft
  Entra ID institucional real.
- No se ha implementado una herramienta operativa
  para la vinculacion historica supervisada.
- El cierre de S1-S4 no representa la aprobacion de
  un despliegue productivo.

## Evidencia

- Pruebas unitarias S1: PASS.
- Pruebas HTTP S2: PASS.
- Integracion PostgreSQL 16 S3: PASS.
- Validacion de tenant S4: PASS.
- go test -count=1 ./...: PASS.
- go vet ./...: PASS.
- Regresion PostgreSQL 16 posterior a S4: PASS.

Resultado integrado:

CIERRE AUTOMATIZADO MVP3 POSTGRESQL 16: PASS

## Proximos pasos

- Definir y aprobar la operacion administrativa
  de vinculacion historica.
- Probar dicha operacion en PostgreSQL efimero.
- Validar autenticacion institucional real antes
  de habilitar el flujo para usuarios existentes.
- Evaluar condiciones de despliegue y rollback.

SEC-006 permanece abierto para esos pasos.
