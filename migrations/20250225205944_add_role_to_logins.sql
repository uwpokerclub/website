-- +goose Up
-- +goose StatementBegin
ALTER TABLE logins ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'executive';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE logins DROP COLUMN role;
-- +goose StatementEnd
