-- Modify "template_history" table
ALTER TABLE "template_history" ADD COLUMN "trust_center_id" character varying NULL;
-- Modify "trust_center_doc_history" table
ALTER TABLE "trust_center_doc_history" ADD COLUMN "owner_id" character varying NULL;
-- Modify "templates" table
ALTER TABLE "templates" ADD CONSTRAINT "templates_check" CHECK ((trust_center_id IS NOT NULL) OR ((kind)::text <> 'TRUSTCENTER_NDA'::text)), ADD COLUMN "trust_center_id" character varying NULL, ADD CONSTRAINT "templates_trust_centers_templates" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "template_trust_center_id" to table: "templates"
CREATE UNIQUE INDEX "template_trust_center_id" ON "templates" ("trust_center_id") WHERE ((deleted_at IS NULL) AND ((kind)::text = 'TRUSTCENTER_NDA'::text));
-- Modify "trust_center_docs" table
ALTER TABLE "trust_center_docs" ADD COLUMN "owner_id" character varying NULL, ADD CONSTRAINT "trust_center_docs_organizations_trust_center_docs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "trustcenterdoc_owner_id" to table: "trust_center_docs"
CREATE INDEX "trustcenterdoc_owner_id" ON "trust_center_docs" ("owner_id") WHERE (deleted_at IS NULL);
