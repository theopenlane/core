-- +goose Up
-- create "assessment_domains" table
CREATE TABLE "assessment_domains" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "custom_domain_id" character varying NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "assessment_domains_custom_domains_custom_domain" FOREIGN KEY ("custom_domain_id") REFERENCES "custom_domains" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "assessment_domains_organizations_assessment_domains" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "assessment_domain_custom_domain_id_idx" to table: "assessment_domains"
CREATE INDEX "assessment_domain_custom_domain_id_idx" ON "assessment_domains" ("custom_domain_id");
-- create index "assessment_domain_owner_id_idx" to table: "assessment_domains"
CREATE INDEX "assessment_domain_owner_id_idx" ON "assessment_domains" ("owner_id");
-- create index "assessmentdomain_owner_id" to table: "assessment_domains"
CREATE UNIQUE INDEX "assessmentdomain_owner_id" ON "assessment_domains" ("owner_id");

-- +goose Down
-- reverse: create index "assessmentdomain_owner_id" to table: "assessment_domains"
DROP INDEX "assessmentdomain_owner_id";
-- reverse: create index "assessment_domain_owner_id_idx" to table: "assessment_domains"
DROP INDEX "assessment_domain_owner_id_idx";
-- reverse: create index "assessment_domain_custom_domain_id_idx" to table: "assessment_domains"
DROP INDEX "assessment_domain_custom_domain_id_idx";
-- reverse: create "assessment_domains" table
DROP TABLE "assessment_domains";
