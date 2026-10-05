-- Backfill memberships.created_at from first participation.
--
-- #429 added created_at with no backfill, on the reasoning that the dates could not
-- be recovered. They can: members are registered for the semester at their first
-- event, so the earliest event a membership entered is when that membership was
-- created. Against the production snapshot this dates 95-99% of each real term's
-- memberships, and no derived date falls before its own semester's start.
--
-- Two deliberate omissions:
--
--   Rows that already carry a date are left alone. Memberships created since #429
--   shipped on 2026-09-06 hold real insert timestamps, and the IS NULL guard is what
--   keeps this migration from overwriting them with a later first-event date.
--
--   source stays NULL. Discord registration is 0.9% of dated signups and cannot be
--   distinguished historically, so filling 'admin' would be a guess that renders as
--   a fabricated all-admin history in the signup timeline's source split. A NULL
--   source reads as "unknown", which is what it is.
--
-- Event starts are timestamptz instants; convert them to UTC before storing in the
-- timestamp-without-time-zone created_at column so the result is independent of the
-- PostgreSQL session TimeZone.
--
-- Note that this makes created_at mean two things: the row's creation time for
-- memberships created from 2026-09-06 onward, and the member's first event before
-- that. For anyone registered at an event those coincide. They diverge for a member
-- registered on one night who did not play until a later one, where the backfilled
-- date is the later of the two.
UPDATE "memberships" m
SET "created_at" = sub.first_event
FROM (
  SELECT p."membership_id", MIN(e."start_date" AT TIME ZONE 'UTC') AS first_event
  FROM "participants" p
  JOIN "memberships" participant_membership ON participant_membership."id" = p."membership_id"
  JOIN "events" e ON e."id" = p."event_id"
    AND e."semester_id" = participant_membership."semester_id"
  GROUP BY p."membership_id"
) sub
WHERE m."id" = sub."membership_id"
  AND m."created_at" IS NULL;
