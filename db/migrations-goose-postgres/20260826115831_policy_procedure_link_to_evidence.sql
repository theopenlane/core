-- +goose Up
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "evidence_internal_policies" character varying NULL, ADD CONSTRAINT "internal_policies_evidences_internal_policies" FOREIGN KEY ("evidence_internal_policies") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "evidence_procedures" character varying NULL, ADD CONSTRAINT "procedures_evidences_procedures" FOREIGN KEY ("evidence_procedures") REFERENCES "evidences" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP CONSTRAINT "procedures_evidences_procedures", DROP COLUMN "evidence_procedures";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP CONSTRAINT "internal_policies_evidences_internal_policies", DROP COLUMN "evidence_internal_policies";
