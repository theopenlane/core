-- +goose Up
-- create "template_recipients" table
CREATE TABLE "template_recipients" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NOT NULL, "ttl" timestamptz NOT NULL, "email" character varying NOT NULL, "secret" bytea NOT NULL, "send_attempts" bigint NOT NULL DEFAULT 1, "status" character varying NOT NULL DEFAULT 'ACTIVE', "document_data_id" character varying NULL, "template_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "template_recipients_document_data_document" FOREIGN KEY ("document_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "template_recipients_templates_template" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "template_recipients_token_key" to table: "template_recipients"
CREATE UNIQUE INDEX "template_recipients_token_key" ON "template_recipients" ("token");
-- create index "templaterecipient_id" to table: "template_recipients"
CREATE UNIQUE INDEX "templaterecipient_id" ON "template_recipients" ("id");
-- create index "templaterecipient_token" to table: "template_recipients"
CREATE UNIQUE INDEX "templaterecipient_token" ON "template_recipients" ("token") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "templaterecipient_token" to table: "template_recipients"
DROP INDEX "templaterecipient_token";
-- reverse: create index "templaterecipient_id" to table: "template_recipients"
DROP INDEX "templaterecipient_id";
-- reverse: create index "template_recipients_token_key" to table: "template_recipients"
DROP INDEX "template_recipients_token_key";
-- reverse: create "template_recipients" table
DROP TABLE "template_recipients";
