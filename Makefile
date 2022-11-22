DB_URL=postgres://root:secret@localhost:5432/postgres?sslmode=disable

migrate-up:
	goose -dir ${CURDIR}/sql/migrations/ -table schema_migrations postgres ${DB_URL} up

migrate-status:
	goose -dir ${CURDIR}/sql/migrations/ -table schema_migrations postgres ${DB_URL} status

migrate-down:
	goose -dir ${CURDIR}/sql/migrations/ -table schema_migrations postgres ${DB_URL} down

migrate-create-%:
	goose -dir ${CURDIR}/sql/migrations/ -table schema_migrations create $(*F) sql

gen-sql:
	sqlc generate

clean-sql:
	rm -rf ./generated/*

test:
	go test -v -cover ./...

.PHONY: migrate-up migrate-status migrate-down migrate-create  gen-sql clean-sql

## When building the application we have to pass the simplebank in docker postgres setup

