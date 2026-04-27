-- Modify "directory_membership_history" table
ALTER TABLE "directory_membership_history" ADD COLUMN "identity_holder_id" character varying NULL;
-- Modify "entity_history" table
ALTER TABLE "entity_history" ALTER COLUMN "tier" SET DEFAULT 'LOW';
