-- +goose Up
-- create "assessment_policies" table
CREATE TABLE "assessment_policies" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "policy_revision" character varying NULL, "assessment_id" character varying NOT NULL, "internal_policy_id" character varying NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "assessment_policies_assessments_assessment" FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "assessment_policies_internal_policies_internal_policy" FOREIGN KEY ("internal_policy_id") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "assessment_policies_organizations_assessment_policies" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "assessment_policy_internal_policy_id_idx" to table: "assessment_policies"
CREATE INDEX "assessment_policy_internal_policy_id_idx" ON "assessment_policies" ("internal_policy_id");
-- create index "assessment_policy_owner_id_idx" to table: "assessment_policies"
CREATE INDEX "assessment_policy_owner_id_idx" ON "assessment_policies" ("owner_id");
-- create index "assessmentpolicy_assessment_id_internal_policy_id" to table: "assessment_policies"
CREATE UNIQUE INDEX "assessmentpolicy_assessment_id_internal_policy_id" ON "assessment_policies" ("assessment_id", "internal_policy_id");
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "organization_assessment_policy_creators" character varying NULL, ADD CONSTRAINT "groups_organizations_assessment_policy_creators" FOREIGN KEY ("organization_assessment_policy_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_organizations_assessment_policy_creators", DROP COLUMN "organization_assessment_policy_creators";
-- reverse: create index "assessmentpolicy_assessment_id_internal_policy_id" to table: "assessment_policies"
DROP INDEX "assessmentpolicy_assessment_id_internal_policy_id";
-- reverse: create index "assessment_policy_owner_id_idx" to table: "assessment_policies"
DROP INDEX "assessment_policy_owner_id_idx";
-- reverse: create index "assessment_policy_internal_policy_id_idx" to table: "assessment_policies"
DROP INDEX "assessment_policy_internal_policy_id_idx";
-- reverse: create "assessment_policies" table
DROP TABLE "assessment_policies";
