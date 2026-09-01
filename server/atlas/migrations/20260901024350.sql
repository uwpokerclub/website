-- Modify "participants" table
ALTER TABLE "participants" DROP COLUMN "placement", ADD COLUMN "points" integer NOT NULL DEFAULT 0;
