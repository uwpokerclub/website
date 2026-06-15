-- Modify "blinds" table
ALTER TABLE "blinds" DROP CONSTRAINT "fk_structures_blinds", ADD CONSTRAINT "fk_structures_blinds" FOREIGN KEY ("structure_id") REFERENCES "structures" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
