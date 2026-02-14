-- Restructure participants table: change PK from composite (membership_id, event_id) to autoincrement id
-- This allows membership_id to be nullable for ON DELETE SET NULL behavior

-- 1. Drop the existing composite primary key
ALTER TABLE "participants" DROP CONSTRAINT "participants_pkey";

-- 2. Make id the new primary key
ALTER TABLE "participants" ADD PRIMARY KEY ("id");

-- 3. Drop the existing unique constraint on id (no longer needed since it's the PK)
ALTER TABLE "participants" DROP CONSTRAINT "uni_participants_id";

-- 4. Add unique index on (membership_id, event_id) to preserve the uniqueness constraint
CREATE UNIQUE INDEX "idx_membership_event" ON "participants" ("membership_id", "event_id");

-- 5. Make membership_id nullable
ALTER TABLE "participants" ALTER COLUMN "membership_id" DROP NOT NULL;

-- 6. Drop existing FK constraints on membership_id and re-add with ON DELETE SET NULL
ALTER TABLE "participants" DROP CONSTRAINT IF EXISTS "participants_membership_id_fkey";
ALTER TABLE "participants" DROP CONSTRAINT IF EXISTS "fk_participants_membership";
ALTER TABLE "participants" ADD CONSTRAINT "fk_participants_membership"
  FOREIGN KEY ("membership_id") REFERENCES "memberships" ("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- Restructure rankings table: add autoincrement id as PK, make membership_id a unique column
-- This allows GORM to properly manage the FK constraint with ON DELETE CASCADE

-- 7. Drop the existing primary key on membership_id
ALTER TABLE "rankings" DROP CONSTRAINT "rankings_pkey";

-- 8. Add autoincrement id column as the new primary key
ALTER TABLE "rankings" ADD COLUMN "id" bigserial PRIMARY KEY;

-- 9. Add unique index on membership_id (matches GORM uniqueIndex convention)
CREATE UNIQUE INDEX "idx_rankings_membership_id" ON "rankings" ("membership_id");

-- 10. Drop old FK and re-add with ON DELETE CASCADE
ALTER TABLE "rankings" DROP CONSTRAINT "fk_memberships_ranking";
ALTER TABLE "rankings" ADD CONSTRAINT "fk_memberships_ranking"
  FOREIGN KEY ("membership_id") REFERENCES "memberships" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
