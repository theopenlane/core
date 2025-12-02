-- +goose Up
-- modify "notes" table
ALTER TABLE "notes" ADD COLUMN "internal_policy_comments" character varying NULL, ADD COLUMN "procedure_comments" character varying NULL, ADD COLUMN "risk_comments" character varying NULL, ADD CONSTRAINT "notes_internal_policies_comments" FOREIGN KEY ("internal_policy_comments") REFERENCES "internal_policies" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_procedures_comments" FOREIGN KEY ("procedure_comments") REFERENCES "procedures" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "notes_risks_comments" FOREIGN KEY ("risk_comments") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "notes" table
ALTER TABLE "notes" DROP CONSTRAINT "notes_risks_comments", DROP CONSTRAINT "notes_procedures_comments", DROP CONSTRAINT "notes_internal_policies_comments", DROP COLUMN "risk_comments", DROP COLUMN "procedure_comments", DROP COLUMN "internal_policy_comments";
