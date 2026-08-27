-- Modify "participants" table
ALTER TABLE "participants" DROP CONSTRAINT "fk_events_entries", ADD CONSTRAINT "fk_events_entries" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
