#!/bin/bash
set -euo pipefail

: "${MSSQL_SA_PASSWORD:?MSSQL_SA_PASSWORD is required}"
DB_NAME="${DB_NAME:-poli-redi-database}"
DB_USER="${DB_USER:-poli-redi-admin}"
SQLCMD="/opt/mssql-tools18/bin/sqlcmd"

if [[ ! "$DB_NAME" =~ ^[A-Za-z0-9_-]+$ || ! "$DB_USER" =~ ^[A-Za-z0-9_-]+$ ]]; then
  echo "DB_NAME and DB_USER may only contain letters, numbers, underscores, and hyphens." >&2
  exit 1
fi
if [[ "$MSSQL_SA_PASSWORD" == *"'"* ]]; then
  echo "MSSQL_SA_PASSWORD may not contain a single quote for local initialization." >&2
  exit 1
fi

/opt/mssql/bin/sqlservr &
SQLSERVR_PID=$!
trap 'kill -TERM "$SQLSERVR_PID"' EXIT

until "$SQLCMD" -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -Q 'SELECT 1' >/dev/null 2>&1; do
  echo "Waiting for SQL Server..."
  sleep 5
done

DB_EXISTS="$("$SQLCMD" -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -h -1 -W \
  -Q "SET NOCOUNT ON; SELECT CASE WHEN DB_ID(N'$DB_NAME') IS NULL THEN 0 ELSE 1 END")"

if [[ "$DB_EXISTS" == "0" ]]; then
  echo "Creating database [$DB_NAME] and applying schema and seed..."
  "$SQLCMD" -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -b -Q "CREATE DATABASE [$DB_NAME];"
  "$SQLCMD" -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -b -I -d "$DB_NAME" -i /scripts/schema.sql
  "$SQLCMD" -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -b -I -d "$DB_NAME" -i /scripts/seed.sql
else
  echo "Database [$DB_NAME] already exists; preserving its schema and data."
fi

until "$SQLCMD" -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -b -h -1 -W \
  -Q "SET NOCOUNT ON; IF DATABASEPROPERTYEX(N'$DB_NAME', 'Status') = 'ONLINE' SELECT 1 ELSE THROW 50000, 'Database is not online', 1;" \
  | grep -q '^1$'; do
  echo "Waiting for database [$DB_NAME] to become online..."
  sleep 2
done

echo "Ensuring local application login [$DB_USER] exists..."
"$SQLCMD" -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -b \
  -v APP_USER="$DB_USER" APP_PASSWORD="$MSSQL_SA_PASSWORD" \
  -Q "IF SUSER_ID(N'\$(APP_USER)') IS NULL
      BEGIN
        CREATE LOGIN [\$(APP_USER)] WITH PASSWORD=N'\$(APP_PASSWORD)';
      END;"
"$SQLCMD" -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -b -d "$DB_NAME" \
  -v APP_USER="$DB_USER" \
  -Q "IF USER_ID(N'\$(APP_USER)') IS NULL
      BEGIN
        CREATE USER [\$(APP_USER)] FOR LOGIN [\$(APP_USER)];
      END;
      ALTER ROLE db_datareader ADD MEMBER [\$(APP_USER)];
      ALTER ROLE db_datawriter ADD MEMBER [\$(APP_USER)];"

wait "$SQLSERVR_PID"
