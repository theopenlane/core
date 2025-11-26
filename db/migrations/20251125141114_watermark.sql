-- Modify "trust_center_watermark_config_history" table
ALTER TABLE "trust_center_watermark_config_history" ADD COLUMN "is_enabled" boolean NULL DEFAULT true;
-- Modify "trust_center_watermark_configs" table
ALTER TABLE "trust_center_watermark_configs" ADD COLUMN "is_enabled" boolean NULL DEFAULT true;
-- Modify "groups" table
ALTER TABLE "groups" ADD COLUMN "organization_trust_center_doc_creators" character varying NULL, ADD COLUMN "organization_trust_center_subprocessor_creators" character varying NULL, ADD CONSTRAINT "groups_organizations_trust_center_doc_creators" FOREIGN KEY ("organization_trust_center_doc_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_trust_center_subprocessor_creators" FOREIGN KEY ("organization_trust_center_subprocessor_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
