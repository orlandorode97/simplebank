DB_URL=postgres://root:secret@localhost:5432/postgres?sslmode=disable
DB_AWS_URL=postgres://root:Jl6JYVWUunMzVnf3IlNN@simplebank.cmqllemktzud.us-east-2.rds.amazonaws.com:5432/postgres
TEST_DB_URL=postgres://root:secret@localhost:5432/simplebank_test?sslmode=disable

setup:
	docker-compose up
stop:
	docker-compose down
migrate-up:
	goose -dir ${CURDIR}/sql/migrations/ -table schema_migrations postgres ${DB_URL} up

migrate-aws-up:
	goose -dir ${CURDIR}/sql/migrations/ -table schema_migrations postgres ${DB_AWS_URL} up

migrate-test-up:
	goose -dir ${CURDIR}/sql/migrations/ -table schema_migrations postgres ${TEST_DB_URL} up

migrate-status:
	goose -dir ${CURDIR}/sql/migrations/ -table schema_migrations postgres ${DB_URL} status

migrate-test-status:
	goose -dir ${CURDIR}/sql/migrations/ -table schema_migrations postgres ${TEST_DB_URL} status

migrate-down:
	goose -dir ${CURDIR}/sql/migrations/ -table schema_migrations postgres ${DB_URL} down

migrate-test_down:
	goose -dir ${CURDIR}/sql/migrations/ -table schema_migrations postgres ${TEST_DB_URL} down

migrate-create-%:
	goose -dir ${CURDIR}/sql/migrations/ -table schema_migrations create $(*F) sql

gen-sql:
	sqlc generate

clean-sql:
	rm -rf ./generated/*

test:
	go test -v -cover ./...

coverage:
	go test -coverprofile cover.out -v ./...
	go tool cover -html=cover.out

build:
	go build -o ${CURDIR}/bin ${CURDIR}/cmd/simplebank

run: build
	./bin/simplebank

gen-mock:
	mockgen -package mockdb -destination store/mockdb/store.go -source store/store.go

docker-db: 
	docker run --name simplebankdb -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:15-alpine
docker-run:
	docker run --name simplebank --network simplebank-network -p 8081:8081 -e DB_SOURCE="postgres://root:secret@simplebankdb:5432/postgres?sslmode=disable" simplebank:latest
docker-build:
	docker build -t simplebank:latest .
docker-network:
	docker network create simplebank-network
docker-connect-network: docker-network
	docker network connect simplebank-network simplebankdb
removes:
	docker-compose down
	docker rmi $$(docker images -q)

.PHONY: migrate-up migrate-status migrate-down migrate-create  gen-sql clean-sql

