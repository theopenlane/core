-- Create "trust_center_docs" table
CREATE TABLE "trust_center_docs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, PRIMARY KEY ("id"));
-- Create "trust_center_doc_history" table
CREATE TABLE "trust_center_doc_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "trustcenterdochistory_history_time" to table: "trust_center_doc_history"
CREATE INDEX "trustcenterdochistory_history_time" ON "trust_center_doc_history" ("history_time");
