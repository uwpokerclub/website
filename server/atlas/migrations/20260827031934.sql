-- Modify "memberships" table
ALTER TABLE "memberships" ADD COLUMN "free_trial_available" boolean NOT NULL DEFAULT true;
-- Modify "semesters" table
ALTER TABLE "semesters" ADD COLUMN "free_trial_limit" smallint NOT NULL DEFAULT 0;
