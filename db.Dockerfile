FROM mysql/mysql-server
EXPOSE 3306
COPY ./sql-scripts/ /docker-entrypoint-initdb.d/