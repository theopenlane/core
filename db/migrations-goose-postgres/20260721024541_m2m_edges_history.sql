-- +goose Up
-- modify "finding_control_history" table
ALTER TABLE "finding_control_history" ADD COLUMN "owner_id" character varying NULL;

-- +goose Down
-- reverse: modify "finding_control_history" table
ALTER TABLE "finding_control_history" DROP COLUMN "owner_id";
