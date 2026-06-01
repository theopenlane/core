-- +goose Up
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "external_file_id" character varying NULL, ADD COLUMN "external_contents" character varying NULL;
-- modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "external_file_id" character varying NULL, ADD COLUMN "external_contents" character varying NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "external_file_id" character varying NULL, ADD COLUMN "external_contents" character varying NULL;
-- create "integration_internal_policies" table
CREATE TABLE "integration_internal_policies" ("integration_id" character varying NOT NULL, "internal_policy_id" character varying NOT NULL, PRIMARY KEY ("integration_id", "internal_policy_id"), CONSTRAINT "integration_internal_policies_integration_id" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "integration_internal_policies_internal_policy_id" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

-- +goose Down
-- reverse: create "integration_internal_policies" table
DROP TABLE "integration_internal_policies";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "external_contents", DROP COLUMN "external_file_id";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "external_contents", DROP COLUMN "external_file_id";
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "external_contents", DROP COLUMN "external_file_id";
