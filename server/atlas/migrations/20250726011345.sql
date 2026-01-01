-- Create "structures" table
CREATE TABLE "structures" (
  "id" serial NOT NULL,
  "name" text NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "blinds" table
CREATE TABLE "blinds" (
  "id" serial NOT NULL,
  "small" integer NOT NULL,
  "big" integer NOT NULL,
  "ante" integer NOT NULL,
  "time" smallint NOT NULL,
  "index" smallint NOT NULL,
  "structure_id" serial NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_structures_blinds" FOREIGN KEY ("structure_id") REFERENCES "structures" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "semesters" table
CREATE TABLE "semesters" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" text NULL,
  "meta" text NULL,
  "start_date" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "end_date" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "starting_budget" numeric NOT NULL DEFAULT 0,
  "current_budget" numeric NOT NULL DEFAULT 0,
  "membership_fee" smallint NOT NULL DEFAULT 0,
  "membership_discount_fee" smallint NOT NULL DEFAULT 0,
  "rebuy_fee" smallint NOT NULL DEFAULT 0,
  PRIMARY KEY ("id")
);
-- Create "events" table
CREATE TABLE "events" (
  "id" serial NOT NULL,
  "name" text NULL,
  "format" text NULL,
  "notes" text NULL,
  "semester_id" uuid NULL,
  "start_date" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "state" smallint NULL DEFAULT 0,
  "structure_id" serial NOT NULL,
  "rebuys" smallint NOT NULL DEFAULT 0,
  "points_multiplier" numeric NOT NULL DEFAULT 1,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_events_semester" FOREIGN KEY ("semester_id") REFERENCES "semesters" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_events_structure" FOREIGN KEY ("structure_id") REFERENCES "structures" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "users" table
CREATE TABLE "users" (
  "id" bigint NOT NULL,
  "first_name" text NULL,
  "last_name" text NULL,
  "email" text NULL,
  "faculty" text NULL,
  "quest_id" text NULL,
  "created_at" timestamp NOT NULL DEFAULT LOCALTIMESTAMP,
  PRIMARY KEY ("id")
);
-- Create "memberships" table
CREATE TABLE "memberships" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" bigint NULL,
  "semester_id" uuid NULL,
  "paid" boolean NOT NULL DEFAULT false,
  "discounted" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_memberships_semester" FOREIGN KEY ("semester_id") REFERENCES "semesters" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_memberships_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "user_semester_unique" to table: "memberships"
CREATE UNIQUE INDEX "user_semester_unique" ON "memberships" ("user_id", "semester_id");
-- Create "participants" table
CREATE TABLE "participants" (
  "id" serial NOT NULL,
  "membership_id" uuid NOT NULL,
  "event_id" serial NOT NULL,
  "placement" integer NULL,
  "signed_out_at" timestamptz NULL,
  PRIMARY KEY ("membership_id", "event_id"),
  CONSTRAINT "fk_events_entries" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_participants_membership" FOREIGN KEY ("membership_id") REFERENCES "memberships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "rankings" table
CREATE TABLE "rankings" (
  "membership_id" uuid NOT NULL,
  "points" integer NULL,
  PRIMARY KEY ("membership_id"),
  CONSTRAINT "fk_memberships_ranking" FOREIGN KEY ("membership_id") REFERENCES "memberships" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "logins" table
CREATE TABLE "logins" (
  "username" text NOT NULL,
  "password" text NOT NULL,
  "role" character varying(20) NOT NULL DEFAULT 'executive',
  PRIMARY KEY ("username")
);
-- Create "sessions" table
CREATE TABLE "sessions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "started_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "expires_at" timestamptz NOT NULL,
  "username" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_logins_sessions" FOREIGN KEY ("username") REFERENCES "logins" ("username") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create "transactions" table
CREATE TABLE "transactions" (
  "id" serial NOT NULL,
  "semester_id" uuid NULL,
  "amount" numeric NOT NULL DEFAULT 0,
  "description" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_transactions_semester" FOREIGN KEY ("semester_id") REFERENCES "semesters" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
