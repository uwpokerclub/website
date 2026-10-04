-- Create "account_activations" table
CREATE TABLE "account_activations" (
  "token_hash" bytea NOT NULL,
  "username" character varying NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "used_at" timestamptz NULL,
  PRIMARY KEY ("token_hash"),
  CONSTRAINT "fk_account_activations_username" FOREIGN KEY ("username") REFERENCES "logins" ("username") ON UPDATE NO ACTION ON DELETE CASCADE
);
