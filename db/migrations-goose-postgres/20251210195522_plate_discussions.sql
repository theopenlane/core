-- +goose Up
-- modify "note_history" table
ALTER TABLE "note_history" ADD COLUMN "note_ref" character varying NULL, ADD COLUMN "discussion_id" character varying NULL, ADD COLUMN "is_edited" boolean NOT NULL DEFAULT false;
-- create "discussion_history" table
CREATE TABLE "discussion_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "external_id" character varying NOT NULL, "is_resolved" boolean NOT NULL DEFAULT false, PRIMARY KEY ("id"));
-- create index "discussionhistory_history_time" to table: "discussion_history"
CREATE INDEX "discussionhistory_history_time" ON "discussion_history" ("history_time");
-- create "discussions" table
CREATE TABLE "discussions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "external_id" character varying NOT NULL, "is_resolved" boolean NOT NULL DEFAULT false, "control_discussions" character varying NULL, "internal_policy_discussions" character varying NULL, "owner_id" character varying NULL, "procedure_discussions" character varying NULL, "risk_discussions" character varying NULL, "subcontrol_discussions" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "discussions_controls_discussions" FOREIGN KEY ("control_discussions") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "discussions_internal_policies_discussions" FOREIGN KEY ("internal_policy_discussions") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "discussions_organizations_discussions" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "discussions_procedures_discussions" FOREIGN KEY ("procedure_discussions") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "discussions_risks_discussions" FOREIGN KEY ("risk_discussions") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "discussions_subcontrols_discussions" FOREIGN KEY ("subcontrol_discussions") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "discussion_owner_id" to table: "discussions"
CREATE INDEX "discussion_owner_id" ON "discussions" ("owner_id") WHERE (deleted_at IS NULL);
-- create index "discussions_external_id_key" to table: "discussions"
CREATE UNIQUE INDEX "discussions_external_id_key" ON "discussions" ("external_id");
-- modify "notes" table
ALTER TABLE "notes" ADD COLUMN "note_ref" character varying NULL, ADD COLUMN "discussion_id" character varying NULL, ADD COLUMN "is_edited" boolean NOT NULL DEFAULT false, ADD COLUMN "discussion_comments" character varying NULL, ADD CONSTRAINT "notes_discussions_comments" FOREIGN KEY ("discussion_comments") REFERENCES "discussions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP CONSTRAINT "notes_discussions_comments", DROP COLUMN "discussion_comments", DROP COLUMN "is_edited", DROP COLUMN "discussion_id", DROP COLUMN "note_ref";
-- reverse: create index "discussions_external_id_key" to table: "discussions"
DROP INDEX "discussions_external_id_key";
-- reverse: create index "discussion_owner_id" to table: "discussions"
DROP INDEX "discussion_owner_id";
-- reverse: create "discussions" table
DROP TABLE "discussions";
-- reverse: create index "discussionhistory_history_time" to table: "discussion_history"
DROP INDEX "discussionhistory_history_time";
-- reverse: create "discussion_history" table
DROP TABLE "discussion_history";
-- reverse: modify "note_history" table
ALTER TABLE "note_history" DROP COLUMN "is_edited", DROP COLUMN "discussion_id", DROP COLUMN "note_ref";
