-- +goose Up
-- +goose StatementBegin
WITH count AS (
  SELECT COUNT(*) AS cnt FROM events WHERE structure_id IS NULL
)

INSERT INTO structures (id, name)
SELECT 1, 'Dummy Structure'
WHERE (SELECT cnt FROM count) > 0
AND NOT EXISTS (
  SELECT 1 FROM structures WHERE id = 1
);

UPDATE events SET structure_id = 1 WHERE structure_id IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
