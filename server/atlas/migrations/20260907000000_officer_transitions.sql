-- Officer handovers are staged so the outgoing team retains access until the
-- incoming president proves possession of their activation link.
CREATE TABLE "officer_transitions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "initiated_by" character varying NOT NULL,
  "president_username" character varying NOT NULL,
  "vice_president_username" character varying NOT NULL,
  "secretary_username" character varying NOT NULL,
  "treasurer_username" character varying NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'pending',
  "created_at" timestamp NOT NULL DEFAULT LOCALTIMESTAMP,
  "resolved_at" timestamp NULL,
  PRIMARY KEY ("id")
);
ALTER TABLE "account_activations" ADD COLUMN "transition_id" uuid NULL REFERENCES "officer_transitions" ("id") ON DELETE CASCADE;
-- Atlas cannot infer this partial index from the model.  Keep the matching
-- exclusion in atlas.hcl or a later generated migration will drop it.
CREATE UNIQUE INDEX "one_pending_transition" ON "officer_transitions" ((true)) WHERE "status" = 'pending';
