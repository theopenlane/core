-- Modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" ADD COLUMN "owner_id" character varying NULL;
-- Modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" ADD COLUMN "owner_id" character varying NULL, ADD CONSTRAINT "trust_center_docs_organizations_trust_center_docs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "trustcenterdoc_owner_id" to table: "trust_center_docs"
CREATE INDEX "trustcenterdoc_owner_id" ON "trust_center_docs" ("owner_id") WHERE (deleted_at IS NULL);
-- Modify "trust_centers" table
ALTER TABLE "trust_centers" ADD COLUMN "template_trust_centers" character varying NULL, ADD CONSTRAINT "trust_centers_templates_trust_centers" FOREIGN KEY ("template_trust_centers") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
