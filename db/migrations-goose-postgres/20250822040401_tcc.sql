-- +goose Up
-- modify "trust_center_compliance_history" table
ALTER TABLE "trust_center_compliance_history" ADD COLUMN "standard_id" character varying NOT NULL, ADD COLUMN "trust_center_id" character varying NULL;
-- modify "trust_center_compliances" table
ALTER TABLE "trust_center_compliances" ADD COLUMN "standard_id" character varying NOT NULL, ADD COLUMN "trust_center_id" character varying NULL, ADD CONSTRAINT "trust_center_compliances_standards_trust_center_compliances" FOREIGN KEY ("standard_id") REFERENCES "standards" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "trust_center_compliances_trust_centers_trust_center_compliances" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "trustcentercompliance_standard_id_trust_center_id" to table: "trust_center_compliances"
CREATE UNIQUE INDEX "trustcentercompliance_standard_id_trust_center_id" ON "trust_center_compliances" ("standard_id", "trust_center_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "trustcentercompliance_standard_id_trust_center_id" to table: "trust_center_compliances"
DROP INDEX "trustcentercompliance_standard_id_trust_center_id";
-- reverse: modify "trust_center_compliances" table
ALTER TABLE "trust_center_compliances" DROP CONSTRAINT "trust_center_compliances_trust_centers_trust_center_compliances", DROP CONSTRAINT "trust_center_compliances_standards_trust_center_compliances", DROP COLUMN "trust_center_id", DROP COLUMN "standard_id";
-- reverse: modify "trust_center_compliance_history" table
ALTER TABLE "trust_center_compliance_history" DROP COLUMN "trust_center_id", DROP COLUMN "standard_id";
