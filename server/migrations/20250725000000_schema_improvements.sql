-- +goose Up
-- +goose StatementBegin

-- Change blinds.time to smallint
ALTER TABLE blinds ALTER COLUMN "time" TYPE smallint;

-- Rename FK constraint on blinds.structure_id
ALTER TABLE blinds DROP CONSTRAINT blinds_structure_id_fkey;
ALTER TABLE blinds ADD CONSTRAINT fk_structures_blinds FOREIGN KEY (structure_id) REFERENCES structures(id);

-- Change semesters.id default to gen_random_uuid()
ALTER TABLE semesters ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- Change semesters.start_date to timestamptz
ALTER TABLE semesters ALTER COLUMN start_date TYPE timestamptz;

-- Change semesters.end_date to timestamptz
ALTER TABLE semesters ALTER COLUMN end_date TYPE timestamptz;

-- Rename FK constraint on events.semester_id
ALTER TABLE events DROP CONSTRAINT events_semester_id_fkey;
ALTER TABLE events ADD CONSTRAINT fk_events_semester FOREIGN KEY (semester_id) REFERENCES semesters(id);

-- Rename FK constraint on events.structure_id
ALTER TABLE events DROP CONSTRAINT events_structure_id_fkey;
ALTER TABLE events ADD CONSTRAINT fk_events_structure FOREIGN KEY (structure_id) REFERENCES structures(id);

-- Change events.start_date to timestamptz
ALTER TABLE events ALTER COLUMN start_date TYPE timestamptz;

-- Change users.created_at to timestamp not null default LOCALTIMESTAMP
ALTER TABLE users ALTER COLUMN created_at TYPE timestamp;
ALTER TABLE users ALTER COLUMN created_at SET DEFAULT LOCALTIMESTAMP;

-- Change memberships.id default to gen_random_uuid()
ALTER TABLE memberships ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- Rename FK on memberships.semester_id
ALTER TABLE memberships DROP CONSTRAINT memberships_semester_id_fkey;
ALTER TABLE memberships ADD CONSTRAINT fk_memberships_semester FOREIGN KEY (semester_id) REFERENCES semesters(id);

-- Rename FK on memberships.user_id
ALTER TABLE memberships DROP CONSTRAINT memberships_user_id_fkey;
ALTER TABLE memberships ADD CONSTRAINT fk_memberships_user FOREIGN KEY (user_id) REFERENCES users(id);

-- Rename FK on participants.membership_id
ALTER TABLE participants DROP CONSTRAINT participants_membership_id_fkey;
ALTER TABLE participants ADD CONSTRAINT fk_participants_membership FOREIGN KEY (membership_id) REFERENCES memberships(id);

-- Rename FK on participants.event_id
ALTER TABLE participants DROP CONSTRAINT participants_event_id_fkey;
ALTER TABLE participants ADD CONSTRAINT fk_events_entries FOREIGN KEY (event_id) REFERENCES events(id);

-- Change participants.signed_out_at to timestamptz
ALTER TABLE participants ALTER COLUMN signed_out_at TYPE timestamptz;

-- Rename FK on rankings.membership_id
ALTER TABLE rankings DROP CONSTRAINT rankings_membership_id_fkey;
ALTER TABLE rankings ADD CONSTRAINT fk_memberships_ranking FOREIGN KEY (membership_id) REFERENCES memberships(id);

-- Change logins.password to not null
ALTER TABLE logins ALTER COLUMN password SET NOT NULL;

-- Change sessions.id default to gen_random_uuid()
ALTER TABLE sessions ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- Change sessions.started_at to timestamptz
ALTER TABLE sessions ALTER COLUMN started_at TYPE timestamptz;

-- Change sessions.expires_at to timestamptz
ALTER TABLE sessions ALTER COLUMN expires_at TYPE timestamptz;

-- Rename FK on sessions.username
ALTER TABLE sessions DROP CONSTRAINT sessions_username_fkey;
ALTER TABLE sessions ADD CONSTRAINT fk_logins_sessions FOREIGN KEY (username) REFERENCES logins(username) ON DELETE CASCADE;

-- Rename FK on transactions.semester_id
ALTER TABLE transactions DROP CONSTRAINT transactions_semester_id_fkey;
ALTER TABLE transactions ADD CONSTRAINT fk_transactions_semester FOREIGN KEY (semester_id) REFERENCES semesters(id);

-- Set blinds.structure_id to NOT NULL
ALTER TABLE blinds ALTER COLUMN structure_id SET NOT NULL;

-- Set events.semester_id to nullable (allow NULL)
ALTER TABLE events ALTER COLUMN semester_id DROP NOT NULL;

-- Set events.structure_id to NOT NULL
ALTER TABLE events ALTER COLUMN structure_id SET NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert events.structure_id to nullable (allow NULL)
ALTER TABLE events ALTER COLUMN structure_id DROP NOT NULL;

-- Revert events.semester_id to NOT NULL
ALTER TABLE events ALTER COLUMN semester_id SET NOT NULL;

-- Revert blinds.structure_id to nullable (allow NULL)
ALTER TABLE blinds ALTER COLUMN structure_id DROP NOT NULL;

-- Revert FK rename on transactions.semester_id
ALTER TABLE transactions DROP CONSTRAINT fk_transactions_semester;
ALTER TABLE transactions ADD CONSTRAINT transactions_semester_id_fkey FOREIGN KEY (semester_id) REFERENCES semesters(id);

-- Revert FK rename on sessions.username
ALTER TABLE sessions DROP CONSTRAINT fk_logins_sessions;
ALTER TABLE sessions ADD CONSTRAINT sessions_username_fkey FOREIGN KEY (username) REFERENCES logins(username) ON DELETE CASCADE;

-- Revert sessions.expires_at to timestamp without time zone
ALTER TABLE sessions ALTER COLUMN expires_at TYPE timestamp without time zone;

-- Revert sessions.started_at to timestamp without time zone
ALTER TABLE sessions ALTER COLUMN started_at TYPE timestamp without time zone;

-- Revert sessions.id default to uuid_generate_v4()
ALTER TABLE sessions ALTER COLUMN id SET DEFAULT uuid_generate_v4();

-- Revert logins.password to allow null
ALTER TABLE logins ALTER COLUMN password DROP NOT NULL;

-- Revert FK rename on rankings.membership_id
ALTER TABLE rankings DROP CONSTRAINT fk_memberships_ranking;
ALTER TABLE rankings ADD CONSTRAINT rankings_membership_id_fkey FOREIGN KEY (membership_id) REFERENCES memberships(id);

-- Revert participants.signed_out_at to timestamp without time zone
ALTER TABLE participants ALTER COLUMN signed_out_at TYPE timestamp without time zone;

-- Revert FK rename on participants.event_id
ALTER TABLE participants DROP CONSTRAINT fk_events_entries;
ALTER TABLE participants ADD CONSTRAINT participants_event_id_fkey FOREIGN KEY (event_id) REFERENCES events(id);

-- Revert FK rename on participants.membership_id
ALTER TABLE participants DROP CONSTRAINT fk_participants_membership;
ALTER TABLE participants ADD CONSTRAINT participants_membership_id_fkey FOREIGN KEY (membership_id) REFERENCES memberships(id);

-- Revert FK rename on memberships.user_id
ALTER TABLE memberships DROP CONSTRAINT fk_memberships_user;
ALTER TABLE memberships ADD CONSTRAINT memberships_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);

-- Revert FK rename on memberships.semester_id
ALTER TABLE memberships DROP CONSTRAINT fk_memberships_semester;
ALTER TABLE memberships ADD CONSTRAINT memberships_semester_id_fkey FOREIGN KEY (semester_id) REFERENCES semesters(id);

-- Revert memberships.id default to uuid_generate_v4()
ALTER TABLE memberships ALTER COLUMN id SET DEFAULT uuid_generate_v4();

-- Revert users.created_at to date default ('now'::text)::date
ALTER TABLE users ALTER COLUMN created_at TYPE date;
ALTER TABLE users ALTER COLUMN created_at SET DEFAULT ('now'::text)::date;

-- Revert events.start_date to timestamp with time zone
ALTER TABLE events ALTER COLUMN start_date TYPE timestamp with time zone;

-- Revert FK rename on events.structure_id
ALTER TABLE events DROP CONSTRAINT fk_events_structure;
ALTER TABLE events ADD CONSTRAINT events_structure_id_fkey FOREIGN KEY (structure_id) REFERENCES structures(id);

-- Revert FK rename on events.semester_id
ALTER TABLE events DROP CONSTRAINT fk_events_semester;
ALTER TABLE events ADD CONSTRAINT events_semester_id_fkey FOREIGN KEY (semester_id) REFERENCES semesters(id);

-- Revert semesters.end_date to date
ALTER TABLE semesters ALTER COLUMN end_date TYPE date;

-- Revert semesters.start_date to date
ALTER TABLE semesters ALTER COLUMN start_date TYPE date;

-- Revert semesters.id default to uuid_generate_v4()
ALTER TABLE semesters ALTER COLUMN id SET DEFAULT uuid_generate_v4();

-- Revert FK rename on blinds.structure_id
ALTER TABLE blinds DROP CONSTRAINT fk_structures_blinds;
ALTER TABLE blinds ADD CONSTRAINT blinds_structure_id_fkey FOREIGN KEY (structure_id) REFERENCES structures(id);

-- Revert blinds.time to integer
ALTER TABLE blinds ALTER COLUMN "time" TYPE integer;

-- +goose StatementEnd
