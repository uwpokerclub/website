-- atlas:txmode none

-- Create index "idx_memberships_semester_id" to table: "memberships"
CREATE INDEX CONCURRENTLY "idx_memberships_semester_id" ON "memberships" ("semester_id");
