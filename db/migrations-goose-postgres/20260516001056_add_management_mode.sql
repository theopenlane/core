-- +goose Up
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED';
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED';
-- modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "management_mode" character varying NULL DEFAULT 'OPENLANE_MANAGED';

-- +goose Down
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "management_mode";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "management_mode";
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "management_mode";
