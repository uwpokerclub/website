-- +goose Up
-- +goose StatementBegin
CREATE TABLE structures (
  id SERIAL PRIMARY KEY,
  name VARCHAR NOT NULL
);

CREATE TABLE blinds (
  id SERIAL PRIMARY KEY,
  small INTEGER NOT NULL,
  big INTEGER NOT NULL,
  ante INTEGER NOT NULL,
  time INTEGER NOT NULL,
  index SMALLINT NOT NULL,
  structure_id INTEGER,
  CONSTRAINT blinds_structure_id_fkey FOREIGN KEY (structure_id) REFERENCES structures(id)
);

ALTER TABLE events ADD COLUMN structure_id INTEGER;
ALTER TABLE events ADD CONSTRAINT events_structure_id_fkey FOREIGN KEY (structure_id) REFERENCES structures(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events DROP COLUMN structure_id;
DROP TABLE blinds;
DROP TABLE structures;
-- +goose StatementEnd
