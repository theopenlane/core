-- +goose Up
-- create "trust_center_faqs" table
CREATE TABLE "trust_center_faqs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "trust_center_faq_kind_name" character varying NULL, "reference_link" character varying NULL, "display_order" bigint NULL DEFAULT 0, "note_id" character varying NOT NULL, "trust_center_id" character varying NULL, "trust_center_faq_kind_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "trust_center_faqs_custom_type_enums_trust_center_faq_kind" FOREIGN KEY ("trust_center_faq_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "trust_center_faqs_notes_trust_center_faqs" FOREIGN KEY ("note_id") REFERENCES "notes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "trust_center_faqs_trust_centers_trust_center_faqs" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "trustcenterfaq_note_id_trust_center_id" to table: "trust_center_faqs"
CREATE UNIQUE INDEX "trustcenterfaq_note_id_trust_center_id" ON "trust_center_faqs" ("note_id", "trust_center_id") WHERE (deleted_at IS NULL);
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "trust_center_faq_blocked_groups" character varying NULL, ADD COLUMN "trust_center_faq_editors" character varying NULL, ADD CONSTRAINT "groups_trust_center_faqs_blocked_groups" FOREIGN KEY ("trust_center_faq_blocked_groups") REFERENCES "trust_center_faqs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_trust_center_faqs_editors" FOREIGN KEY ("trust_center_faq_editors") REFERENCES "trust_center_faqs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_trust_center_faqs_editors", DROP CONSTRAINT "groups_trust_center_faqs_blocked_groups", DROP COLUMN "trust_center_faq_editors", DROP COLUMN "trust_center_faq_blocked_groups";
-- reverse: create index "trustcenterfaq_note_id_trust_center_id" to table: "trust_center_faqs"
DROP INDEX "trustcenterfaq_note_id_trust_center_id";
-- reverse: create "trust_center_faqs" table
DROP TABLE "trust_center_faqs";
