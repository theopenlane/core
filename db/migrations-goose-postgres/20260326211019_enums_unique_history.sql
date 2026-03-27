-- +goose Up
-- modify "sla_definition_history" table
ALTER TABLE "sla_definition_history" DROP COLUMN "sla_definition_severity_level_name", DROP COLUMN "sla_definition_severity_level_id";

-- +goose Down
-- reverse: modify "sla_definition_history" table
ALTER TABLE "sla_definition_history" ADD COLUMN "sla_definition_severity_level_id" character varying NULL, ADD COLUMN "sla_definition_severity_level_name" character varying NULL;
