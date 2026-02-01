-- +goose Up
-- modify "trust_center_subprocessors" table
ALTER TABLE "trust_center_subprocessors" DROP COLUMN "category", ADD COLUMN "trust_center_subprocessor_kind_name" character varying NULL, ADD COLUMN "trust_center_subprocessor_kind_id" character varying NULL, ADD CONSTRAINT "trust_center_subprocessors_cus_d5ebb915269b07a0bf77b5b0ec180583" FOREIGN KEY ("trust_center_subprocessor_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "trust_center_subprocessors" table
ALTER TABLE "trust_center_subprocessors" DROP CONSTRAINT "trust_center_subprocessors_cus_d5ebb915269b07a0bf77b5b0ec180583", DROP COLUMN "trust_center_subprocessor_kind_id", DROP COLUMN "trust_center_subprocessor_kind_name", ADD COLUMN "category" character varying NOT NULL;
