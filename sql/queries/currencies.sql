-- name: CreateCurrency :one
INSERT INTO currencies (
  name
) VALUES ( $1 )
RETURNING *;

