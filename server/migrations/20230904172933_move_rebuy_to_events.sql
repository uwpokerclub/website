-- +goose Up
-- +goose StatementBegin
ALTER TABLE participants DROP COLUMN rebuys;
ALTER TABLE events ADD COLUMN rebuys SMALLINT NOT NULL DEFAULT '0'::SMALLINT
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE events DROP COLUMN rebuys;
ALTER TABLE participants ADD COLUMN rebuys SMALLINT NOT NULL DEFAULT '0'::SMALLINT
-- +goose StatementEnd
