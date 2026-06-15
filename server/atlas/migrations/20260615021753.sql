-- Modify "events" table
ALTER TABLE "events" DROP CONSTRAINT "fk_events_structure", ADD CONSTRAINT "fk_events_structure" FOREIGN KEY ("structure_id") REFERENCES "structures" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
