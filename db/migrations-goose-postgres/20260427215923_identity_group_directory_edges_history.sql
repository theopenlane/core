-- +goose Up
-- modify "directory_group_history" table
ALTER TABLE "directory_group_history" ADD COLUMN "identity_holder_id" character varying NULL;
-- modify "entity_history" table
ALTER TABLE "entity_history" ALTER COLUMN "tier" SET DEFAULT 'LOW';

-- +goose Down
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" ALTER COLUMN "tier" SET DEFAULT 'STANDARD';
-- reverse: modify "directory_group_history" table
ALTER TABLE "directory_group_history" DROP COLUMN "identity_holder_id";
