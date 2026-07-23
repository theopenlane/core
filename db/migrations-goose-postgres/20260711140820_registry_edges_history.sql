-- +goose Up
-- modify "system_detail_history" table
ALTER TABLE "system_detail_history" DROP COLUMN "program_id", DROP COLUMN "platform_id";

-- +goose Down
-- reverse: modify "system_detail_history" table
ALTER TABLE "system_detail_history" ADD COLUMN "platform_id" character varying NULL, ADD COLUMN "program_id" character varying NULL;
