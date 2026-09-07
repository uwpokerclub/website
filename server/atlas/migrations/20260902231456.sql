-- Create "event_clocks" table
CREATE TABLE "event_clocks" (
  "event_id" integer NOT NULL,
  "level_index" integer NOT NULL,
  "level_ends_at" timestamptz NOT NULL,
  "paused_at" timestamptz NULL,
  "version" bigint NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("event_id"),
  CONSTRAINT "fk_event_clocks_event" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
