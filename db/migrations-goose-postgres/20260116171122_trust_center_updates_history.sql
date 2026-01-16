-- +goose Up
-- create "trust_center_entity_history" table
CREATE TABLE "trust_center_entity_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "logo_file_id" character varying NULL, "url" character varying NULL, "trust_center_id" character varying NULL, "name" character varying NOT NULL, "entity_type_id" character varying NULL, PRIMARY KEY ("id"));

-- +goose Down
-- reverse: create "trust_center_entity_history" table
DROP TABLE "trust_center_entity_history";
