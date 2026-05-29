-- Modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "external_file_id" character varying NULL, ADD COLUMN "external_contents" character varying NULL;
-- Modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "external_file_id" character varying NULL, ADD COLUMN "external_contents" character varying NULL;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "external_file_id" character varying NULL, ADD COLUMN "external_contents" character varying NULL;
-- Create "integration_internal_policies" table
CREATE TABLE "integration_internal_policies" ("integration_id" character varying NOT NULL, "internal_policy_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "internal_policy_id"), CONSTRAINT "integration_internal_policies_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "integration_internal_policies_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
