-- +goose Up
-- +goose StatementBegin
ALTER TABLE events ADD COLUMN points_multiplier DECIMAL NOT NULL DEFAULT '1'::DECIMAL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events DROP COLUMN points_multiplier;
-- +goose StatementEnd
