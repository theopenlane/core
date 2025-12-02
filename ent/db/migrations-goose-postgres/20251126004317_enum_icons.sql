-- +goose Up
-- modify "custom_type_enums" table
ALTER TABLE "custom_type_enums" ADD COLUMN "icon" character varying NULL;

-- +goose Down
-- reverse: modify "custom_type_enums" table
ALTER TABLE "custom_type_enums" DROP COLUMN "icon";
