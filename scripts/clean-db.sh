#!/bin/sh

DB_NAME=${1-postgres}
flyway \
    -url="jdbc:postgresql://127.0.0.1:5432/${DB_NAME}" \
    -user="postgres" \
    -password="postgres" \
    -cleanDisabled="false" \
    clean
