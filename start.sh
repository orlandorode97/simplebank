#!/bin/sh
# Exit in case there is an error.
set -e
echo "run db migration"
source app.env
./goose_linux_x86_64 -dir migrations/ -table schema_migrations postgres "$DB_SOURCE" up
echo "start the app"
## Take all parameters to the script and run it
exec "$@"
