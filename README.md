### Simple bank
Simple bank a is completed backend tutorial that involves different technologies such as:
1. Gorila mux.
2. Docker
3. Kubernetes
4. AWS
5. gRPC / grpc gateway
6. sqlc

### Getting started

### Migrations
To get the pending status migration is required to have [goose](https://github.com/pressly/goose) installed.

- To verify pending migration run `make migrate-status`.
- To create a migration file run `make migrate-create-%` where `%` is the name of the new migration file. E.Q. `make-migrate-create-add-alter-users-table`.
- To run all the migration run `make migrate-up`.
- To roll back run `make migrate-down`.

