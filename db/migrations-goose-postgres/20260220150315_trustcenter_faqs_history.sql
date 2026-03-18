-- +goose Up
-- create "trust_center_faqs_history" table
CREATE TABLE "trust_center_faqs_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "trust_center_faq_kind_name" character varying NULL, "trust_center_faq_kind_id" character varying NULL, "note_id" character varying NOT NULL, "trust_center_id" character varying NULL, "reference_link" character varying NULL, "display_order" bigint NULL DEFAULT 0, PRIMARY KEY ("id"));
-- create index "trustcenterfaqhistory_history_time" to table: "trust_center_faqs_history"
CREATE INDEX "trustcenterfaqhistory_history_time" ON "trust_center_faqs_history" ("history_time");

-- +goose Down
-- reverse: create index "trustcenterfaqhistory_history_time" to table: "trust_center_faqs_history"
DROP INDEX "trustcenterfaqhistory_history_time";
-- reverse: create "trust_center_faqs_history" table
DROP TABLE "trust_center_faqs_history";
