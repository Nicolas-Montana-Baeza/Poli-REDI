\set ON_ERROR_STOP on
\getenv app_password POSTGRES_APP_PASSWORD

SELECT format(
    'CREATE ROLE poliredi_app LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION',
    :'app_password'
)
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'poliredi_app')
\gexec

ALTER ROLE poliredi_app SET timezone TO 'America/Santiago';
