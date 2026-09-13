-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- modify "asset_history" table
ALTER TABLE "asset_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- modify "contact_history" table
ALTER TABLE "contact_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- modify "finding_history" table
ALTER TABLE "finding_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;

-- +goose Down
-- reverse: modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "finding_history" table
ALTER TABLE "finding_history" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "contact_history" table
ALTER TABLE "contact_history" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "asset_history" table
ALTER TABLE "asset_history" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
