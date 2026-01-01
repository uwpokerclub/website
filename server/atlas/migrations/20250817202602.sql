-- Fix foreign key column types from serial to integer
-- Modify "blinds" table
ALTER TABLE "blinds" ALTER COLUMN "structure_id" DROP DEFAULT;
ALTER TABLE "blinds" ALTER COLUMN "structure_id" TYPE integer;
DROP SEQUENCE IF EXISTS "blinds_structure_id_seq";

-- Modify "events" table  
ALTER TABLE "events" ALTER COLUMN "structure_id" DROP DEFAULT;
ALTER TABLE "events" ALTER COLUMN "structure_id" TYPE integer;
DROP SEQUENCE IF EXISTS "events_structure_id_seq";

-- Modify "participants" table
ALTER TABLE "participants" ALTER COLUMN "event_id" DROP DEFAULT;
ALTER TABLE "participants" ALTER COLUMN "event_id" TYPE integer;
DROP SEQUENCE IF EXISTS "participants_event_id_seq";
