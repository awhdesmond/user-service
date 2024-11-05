#!/bin/sh

DB_NAME=${1-postgres}
flyway \
    -url="jdbc:postgresql://127.0.0.1:5432/${DB_NAME}" \
    -user="postgres" \
    -password="postgres" \
    -locations="filesystem:db/migrations/" \
    -validateMigrationNaming=true \
    migrate
