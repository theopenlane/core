-- Modify "trust_center_subprocessor_history" table
ALTER TABLE "trust_center_subprocessor_history" DROP COLUMN "category", ADD COLUMN "trust_center_subprocessor_kind_name" character varying NULL, ADD COLUMN "trust_center_subprocessor_kind_id" character varying NULL;
