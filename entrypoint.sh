#!/bin/sh
set -e

goose -dir ./migrations postgres "host=$DB_HOST port=$DB_PORT user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME sslmode=disable" up

exec /usr/local/bin/fin-aggregator-service
