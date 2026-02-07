-- Modify "sessions" table
ALTER TABLE "sessions" ADD COLUMN "role" character varying(20) NOT NULL DEFAULT 'executive';
