FROM mcr.microsoft.com/mssql/server:2022-latest

USER root
COPY docker/init/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

COPY database /scripts

ENTRYPOINT ["/entrypoint.sh"]
