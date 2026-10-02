-- Modify "memberships" table
ALTER TABLE "memberships" ADD COLUMN "trial_started_at" timestamp NULL, ADD COLUMN "converted_at" timestamp NULL;
