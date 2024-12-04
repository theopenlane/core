-- +goose Up
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "owner_id" character varying NOT NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "owner_id" character varying NOT NULL, ADD CONSTRAINT "subcontrols_organizations_subcontrols" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

-- +goose Down
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_organizations_subcontrols", DROP COLUMN "owner_id";
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "owner_id";
