-- +goose Up
-- create "trust_center_entity_history" table
CREATE TABLE "trust_center_entity_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "logo_file_id" character varying NULL, "url" character varying NULL, "trust_center_id" character varying NULL, "name" character varying NOT NULL, "entity_type_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "trustcenterentityhistory_history_time" to table: "trust_center_entity_history"
CREATE INDEX "trustcenterentityhistory_history_time" ON "trust_center_entity_history" ("history_time");

-- +goose Down
-- reverse: create index "trustcenterentityhistory_history_time" to table: "trust_center_entity_history"
DROP INDEX "trustcenterentityhistory_history_time";
-- reverse: create "trust_center_entity_history" table
DROP TABLE "trust_center_entity_history";
