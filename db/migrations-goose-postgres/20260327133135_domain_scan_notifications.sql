-- +goose Up
-- modify "sla_definitions" table
ALTER TABLE "sla_definitions" DROP COLUMN "sla_definition_severity_level_name", DROP COLUMN "sla_definition_severity_level_id";

-- +goose Down
-- reverse: modify "sla_definitions" table
ALTER TABLE "sla_definitions" ADD COLUMN "sla_definition_severity_level_id" character varying NULL, ADD COLUMN "sla_definition_severity_level_name" character varying NULL;
