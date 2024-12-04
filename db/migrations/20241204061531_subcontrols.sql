-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "owner_id" character varying NOT NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "owner_id" character varying NOT NULL, ADD CONSTRAINT "subcontrols_organizations_subcontrols" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
