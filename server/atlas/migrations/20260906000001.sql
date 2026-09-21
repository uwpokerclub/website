-- Create index "idx_account_activations_one_unused" to prevent sibling activation links.
CREATE UNIQUE INDEX "idx_account_activations_one_unused" ON "account_activations" ("username") WHERE (used_at IS NULL);
