-- +goose Up
-- modify "trust_center_subprocessor_history" table
ALTER TABLE "trust_center_subprocessor_history" DROP COLUMN "tags", ADD COLUMN "subprocessor_id" character varying NOT NULL, ADD COLUMN "trust_center_id" character varying NULL, ADD COLUMN "countries" jsonb NULL, ADD COLUMN "category" character varying NOT NULL;
-- modify "trust_center_subprocessors" table
ALTER TABLE "trust_center_subprocessors" DROP COLUMN "tags", ADD COLUMN "countries" jsonb NULL, ADD COLUMN "category" character varying NOT NULL, ADD COLUMN "subprocessor_id" character varying NOT NULL, ADD COLUMN "trust_center_id" character varying NULL, ADD CONSTRAINT "trust_center_subprocessors_sub_24055b695e9bd0e49b3edea05d355a0b" FOREIGN KEY ("subprocessor_id") REFERENCES "subprocessors" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "trust_center_subprocessors_tru_bb0fd7936579c86ecda7d42ebfe60199" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "trustcentersubprocessor_subprocessor_id_trust_center_id" to table: "trust_center_subprocessors"
CREATE UNIQUE INDEX "trustcentersubprocessor_subprocessor_id_trust_center_id" ON "trust_center_subprocessors" ("subprocessor_id", "trust_center_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "trustcentersubprocessor_subprocessor_id_trust_center_id" to table: "trust_center_subprocessors"
DROP INDEX "trustcentersubprocessor_subprocessor_id_trust_center_id";
-- reverse: modify "trust_center_subprocessors" table
ALTER TABLE "trust_center_subprocessors" DROP CONSTRAINT "trust_center_subprocessors_tru_bb0fd7936579c86ecda7d42ebfe60199", DROP CONSTRAINT "trust_center_subprocessors_sub_24055b695e9bd0e49b3edea05d355a0b", DROP COLUMN "trust_center_id", DROP COLUMN "subprocessor_id", DROP COLUMN "category", DROP COLUMN "countries", ADD COLUMN "tags" jsonb NULL;
-- reverse: modify "trust_center_subprocessor_history" table
ALTER TABLE "trust_center_subprocessor_history" DROP COLUMN "category", DROP COLUMN "countries", DROP COLUMN "trust_center_id", DROP COLUMN "subprocessor_id", ADD COLUMN "tags" jsonb NULL;
