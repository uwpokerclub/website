-- Modify "logins" table
ALTER TABLE "logins" ADD COLUMN "status" character varying(20) NOT NULL DEFAULT 'active';
