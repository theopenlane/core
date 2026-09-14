-- +goose Up
-- modify "trust_center_faqs" table
ALTER TABLE "trust_center_faqs" ADD COLUMN "category_name" character varying NULL, ADD COLUMN "category_id" character varying NULL, ADD CONSTRAINT "trust_center_faqs_custom_type_enums_category" FOREIGN KEY ("category_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "trustcenterfaq_category_name_created_at" to table: "trust_center_faqs"
CREATE INDEX "trustcenterfaq_category_name_created_at" ON "trust_center_faqs" ("category_name", "created_at") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "trustcenterfaq_category_name_created_at" to table: "trust_center_faqs"
DROP INDEX "trustcenterfaq_category_name_created_at";
-- reverse: modify "trust_center_faqs" table
ALTER TABLE "trust_center_faqs" DROP CONSTRAINT "trust_center_faqs_custom_type_enums_category", DROP COLUMN "category_id", DROP COLUMN "category_name";
