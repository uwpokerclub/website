-- Modify "memberships" table
-- Drop whichever constraint name exists (legacy vs GORM naming)
ALTER TABLE "memberships" DROP CONSTRAINT IF EXISTS "memberships_user_id_fkey";
ALTER TABLE "memberships" DROP CONSTRAINT IF EXISTS "fk_memberships_user";
ALTER TABLE "memberships" ADD CONSTRAINT "fk_memberships_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
