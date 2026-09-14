-- +goose Up
-- modify "trust_center_faqs_history" table
ALTER TABLE "trust_center_faqs_history" ADD COLUMN "category_name" character varying NULL, ADD COLUMN "category_id" character varying NULL;

-- +goose Down
-- reverse: modify "trust_center_faqs_history" table
ALTER TABLE "trust_center_faqs_history" DROP COLUMN "category_id", DROP COLUMN "category_name";
