-- atlas:txmode none

-- Modify "memberships" table
ALTER TABLE "memberships" ADD COLUMN "created_at" timestamp NULL, ADD COLUMN "source" text NULL;
-- Create index "idx_memberships_semester_created" to table: "memberships"
CREATE INDEX CONCURRENTLY "idx_memberships_semester_created" ON "memberships" ("semester_id", "created_at");
