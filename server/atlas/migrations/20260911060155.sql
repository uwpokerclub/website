-- Record the transition that created a staged login. Existing logins remain
-- unowned so cancellation cannot delete them.
ALTER TABLE "logins" ADD COLUMN "staged_transition_id" uuid NULL, ADD CONSTRAINT "fk_logins_staged_transition" FOREIGN KEY ("staged_transition_id") REFERENCES "officer_transitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
