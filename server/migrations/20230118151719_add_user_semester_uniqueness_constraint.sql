-- +goose Up
-- +goose StatementBegin
ALTER TABLE memberships ADD CONSTRAINT memberships_user_id_semester_id_key UNIQUE (user_id, semester_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE memberships DROP CONSTRAINT memberships_user_id_semester_id_key;
-- +goose StatementEnd
