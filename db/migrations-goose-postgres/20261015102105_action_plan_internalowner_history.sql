-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "internal_owner" character varying NULL, ADD COLUMN "internal_owner_user_id" character varying NULL, ADD COLUMN "internal_owner_group_id" character varying NULL, ADD COLUMN "internal_owner_identity_holder_id" character varying NULL;

-- +goose Down
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "internal_owner_identity_holder_id", DROP COLUMN "internal_owner_group_id", DROP COLUMN "internal_owner_user_id", DROP COLUMN "internal_owner";
