-- Backfill existing sessions with the correct role from logins table
UPDATE "sessions" s SET role = l.role FROM "logins" l WHERE s.username = l.username;
