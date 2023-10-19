-- +goose Up
-- +goose StatementBegin
UPDATE events SET structure_id = 1 WHERE structure_id IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
