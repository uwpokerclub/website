-- Modify "rankings" table
ALTER TABLE "rankings" ADD COLUMN IF NOT EXISTS "attendance" integer NOT NULL DEFAULT 0;
