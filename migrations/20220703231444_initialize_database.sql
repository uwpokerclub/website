-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS logins (
  username VARCHAR PRIMARY KEY,
  password VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
  id BIGINT PRIMARY KEY,
  first_name VARCHAR,
  last_name VARCHAR,
  email VARCHAR,
  faculty VARCHAR,
  quest_id VARCHAR,
  created_at DATE NOT NULL DEFAULT CURRENT_DATE
);

CREATE TABLE IF NOT EXISTS semesters (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  name VARCHAR,
  meta VARCHAR,
  start_date DATE NOT NULL DEFAULT CURRENT_DATE,
  end_date DATE NOT NULL DEFAULT CURRENT_DATE,
  starting_budget DECIMAL NOT NULL DEFAULT '0'::DECIMAL,
  current_budget DECIMAL NOT NULL DEFAULT '0'::DECIMAL,
  membership_fee SMALLINT NOT NULL DEFAULT '0'::SMALLINT,
  membership_discount_fee SMALLINT NOT NULL DEFAULT '0'::SMALLINT,
  rebuy_fee SMALLINT NOT NULL DEFAULT '0'::SMALLINT
);

CREATE TABLE IF NOT EXISTS memberships (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id BIGINT,
  semester_id uuid,
  paid BOOLEAN NOT NULL DEFAULT FALSE,
  discounted BOOLEAN NOT NULL DEFAULT FALSE,
  CONSTRAINT memberships_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT memberships_semester_id_fkey FOREIGN KEY (semester_id) REFERENCES semesters(id)
);

CREATE TABLE IF NOT EXISTS events (
  id SERIAL PRIMARY KEY,
  name VARCHAR,
  format VARCHAR,
  notes VARCHAR,
  semester_id uuid,
  start_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  state SMALLINT DEFAULT 0,
  CONSTRAINT events_semester_id_fkey FOREIGN KEY (semester_id) REFERENCES semesters(id)
);

CREATE TABLE IF NOT EXISTS participants (
  id SERIAL,
  membership_id uuid,
  event_id SERIAL,
  placement INTEGER,
  signed_out_at TIMESTAMP,
  rebuys SMALLINT NOT NULL DEFAULT '0'::SMALLINT,
  PRIMARY KEY (membership_id, event_id),
  CONSTRAINT participants_event_id_fkey FOREIGN KEY (event_id) REFERENCES events(id),
  CONSTRAINT participants_membership_id_fkey FOREIGN KEY (membership_id) REFERENCES memberships(id)
);

CREATE TABLE IF NOT EXISTS rankings (
  membership_id uuid,
  points integer,
  PRIMARY KEY (membership_id),
  CONSTRAINT rankings_membership_id_fkey FOREIGN KEY (membership_id) REFERENCES memberships(id)
);

CREATE TABLE IF NOT EXISTS transactions (
  id SERIAL PRIMARY KEY,
  semester_id uuid,
  amount DECIMAL NOT NULL DEFAULT '0'::DECIMAL,
  description TEXT,
  CONSTRAINT transactions_semester_id_fkey FOREIGN KEY (semester_id) REFERENCES semesters(id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE transactions;

DROP TABLE rankings;

DROP TABLE participants;

DROP TABLE events;

DROP TABLE memberships;

DROP TABLE semesters;

DROP TABLE users;

DROP TABLE logins;

DROP EXTENSION "uuid-ossp";

-- +goose StatementEnd
