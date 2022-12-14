-- +goose Up
-- +goose StatementBegin
INSERT INTO currencies (name) VALUES ('USD'), ('EUR'), ('MXM'), ('CAD'), ('JPY');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
