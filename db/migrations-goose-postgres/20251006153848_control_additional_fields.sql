-- +goose Up
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "aliases" jsonb NULL, ADD COLUMN "responsible_party_id" character varying NULL;
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "aliases" jsonb NULL, ADD COLUMN "responsible_party_id" character varying NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "aliases" jsonb NULL, ADD COLUMN "responsible_party_id" character varying NULL, ADD CONSTRAINT "controls_entities_responsible_party" FOREIGN KEY ("responsible_party_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "aliases" jsonb NULL, ADD COLUMN "responsible_party_id" character varying NULL, ADD CONSTRAINT "subcontrols_entities_responsible_party" FOREIGN KEY ("responsible_party_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "notes" table
ALTER TABLE "notes" ADD COLUMN "control_comments" character varying NULL, ADD COLUMN "subcontrol_comments" character varying NULL, ADD CONSTRAINT "notes_controls_comments" FOREIGN KEY ("control_comments") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_subcontrols_comments" FOREIGN KEY ("subcontrol_comments") REFERENCES "subcontrols" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP CONSTRAINT "notes_subcontrols_comments", DROP CONSTRAINT "notes_controls_comments", DROP COLUMN "subcontrol_comments", DROP COLUMN "control_comments";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP CONSTRAINT "subcontrols_entities_responsible_party", DROP COLUMN "responsible_party_id", DROP COLUMN "aliases";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP CONSTRAINT "controls_entities_responsible_party", DROP COLUMN "responsible_party_id", DROP COLUMN "aliases";
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "responsible_party_id", DROP COLUMN "aliases";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "responsible_party_id", DROP COLUMN "aliases";
