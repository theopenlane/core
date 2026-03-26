-- +goose Up
-- drop index "customtypeenum_name_owner_id" from table: "custom_type_enums"
DROP INDEX "customtypeenum_name_owner_id";
-- create index "customtypeenum_name_object_type_field_owner_id" to table: "custom_type_enums"
CREATE UNIQUE INDEX "customtypeenum_name_object_type_field_owner_id" ON "custom_type_enums" ("name", "object_type", "field", "owner_id") WHERE (deleted_at IS NULL);
-- modify "sla_definitions" table
ALTER TABLE "sla_definitions" DROP COLUMN "sla_definition_severity_level_name", DROP COLUMN "sla_definition_severity_level_id";

-- +goose Down
-- reverse: modify "sla_definitions" table
ALTER TABLE "sla_definitions" ADD COLUMN "sla_definition_severity_level_id" character varying NULL, ADD COLUMN "sla_definition_severity_level_name" character varying NULL;
-- reverse: create index "customtypeenum_name_object_type_field_owner_id" to table: "custom_type_enums"
DROP INDEX "customtypeenum_name_object_type_field_owner_id";
-- reverse: drop index "customtypeenum_name_owner_id" from table: "custom_type_enums"
CREATE UNIQUE INDEX "customtypeenum_name_owner_id" ON "custom_type_enums" ("name", "owner_id") WHERE (deleted_at IS NULL);
