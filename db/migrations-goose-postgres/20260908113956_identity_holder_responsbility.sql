-- +goose Up
-- modify "identity_holders" table
ALTER TABLE "identity_holders" ADD COLUMN "internal_owner_identity_holder_id" character varying NULL, ADD CONSTRAINT "identity_holders_identity_hold_a071c96868edb317166fe843ed012736" FOREIGN KEY ("internal_owner_identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "identity_holders_internal_owner_identity_holder_id_key" to table: "identity_holders"
CREATE UNIQUE INDEX "identity_holders_internal_owner_identity_holder_id_key" ON "identity_holders" ("internal_owner_identity_holder_id");
-- modify "assets" table
ALTER TABLE "assets" ADD COLUMN "internal_owner_identity_holder_id" character varying NULL, ADD CONSTRAINT "assets_identity_holders_internal_owner_identity_holder" FOREIGN KEY ("internal_owner_identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "campaigns" table
ALTER TABLE "campaigns" ADD COLUMN "internal_owner_identity_holder_id" character varying NULL, ADD CONSTRAINT "campaigns_identity_holders_internal_owner_identity_holder" FOREIGN KEY ("internal_owner_identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "entities" table
ALTER TABLE "entities" ADD COLUMN "internal_owner_identity_holder_id" character varying NULL, ADD COLUMN "reviewed_by_identity_holder_id" character varying NULL, ADD CONSTRAINT "entities_identity_holders_internal_owner_identity_holder" FOREIGN KEY ("internal_owner_identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_identity_holders_reviewed_by_identity_holder" FOREIGN KEY ("reviewed_by_identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "findings" table
ALTER TABLE "findings" ADD COLUMN "reviewed_by_identity_holder_id" character varying NULL, ADD COLUMN "assigned_to_identity_holder_id" character varying NULL, ADD CONSTRAINT "findings_identity_holders_assigned_to_identity_holder" FOREIGN KEY ("assigned_to_identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "findings_identity_holders_reviewed_by_identity_holder" FOREIGN KEY ("reviewed_by_identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "platforms" table
ALTER TABLE "platforms" ADD COLUMN "internal_owner_identity_holder_id" character varying NULL, ADD COLUMN "business_owner_identity_holder_id" character varying NULL, ADD COLUMN "technical_owner_identity_holder_id" character varying NULL, ADD COLUMN "security_owner_identity_holder_id" character varying NULL, ADD CONSTRAINT "platforms_identity_holders_business_owner_identity_holder" FOREIGN KEY ("business_owner_identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_identity_holders_internal_owner_identity_holder" FOREIGN KEY ("internal_owner_identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_identity_holders_security_owner_identity_holder" FOREIGN KEY ("security_owner_identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "platforms_identity_holders_technical_owner_identity_holder" FOREIGN KEY ("technical_owner_identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "scans" table
ALTER TABLE "scans" ADD COLUMN "reviewed_by_identity_holder_id" character varying NULL, ADD COLUMN "assigned_to_identity_holder_id" character varying NULL, ADD CONSTRAINT "scans_identity_holders_assigned_to_identity_holder" FOREIGN KEY ("assigned_to_identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_identity_holders_reviewed_by_identity_holder" FOREIGN KEY ("reviewed_by_identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" ADD COLUMN "reviewed_by_identity_holder_id" character varying NULL, ADD COLUMN "assigned_to_identity_holder_id" character varying NULL, ADD CONSTRAINT "vulnerabilities_identity_holders_assigned_to_identity_holder" FOREIGN KEY ("assigned_to_identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "vulnerabilities_identity_holders_reviewed_by_identity_holder" FOREIGN KEY ("reviewed_by_identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" DROP CONSTRAINT "vulnerabilities_identity_holders_reviewed_by_identity_holder", DROP CONSTRAINT "vulnerabilities_identity_holders_assigned_to_identity_holder", DROP COLUMN "assigned_to_identity_holder_id", DROP COLUMN "reviewed_by_identity_holder_id";
-- reverse: modify "scans" table
ALTER TABLE "scans" DROP CONSTRAINT "scans_identity_holders_reviewed_by_identity_holder", DROP CONSTRAINT "scans_identity_holders_assigned_to_identity_holder", DROP COLUMN "assigned_to_identity_holder_id", DROP COLUMN "reviewed_by_identity_holder_id";
-- reverse: modify "platforms" table
ALTER TABLE "platforms" DROP CONSTRAINT "platforms_identity_holders_technical_owner_identity_holder", DROP CONSTRAINT "platforms_identity_holders_security_owner_identity_holder", DROP CONSTRAINT "platforms_identity_holders_internal_owner_identity_holder", DROP CONSTRAINT "platforms_identity_holders_business_owner_identity_holder", DROP COLUMN "security_owner_identity_holder_id", DROP COLUMN "technical_owner_identity_holder_id", DROP COLUMN "business_owner_identity_holder_id", DROP COLUMN "internal_owner_identity_holder_id";
-- reverse: modify "findings" table
ALTER TABLE "findings" DROP CONSTRAINT "findings_identity_holders_reviewed_by_identity_holder", DROP CONSTRAINT "findings_identity_holders_assigned_to_identity_holder", DROP COLUMN "assigned_to_identity_holder_id", DROP COLUMN "reviewed_by_identity_holder_id";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP CONSTRAINT "entities_identity_holders_reviewed_by_identity_holder", DROP CONSTRAINT "entities_identity_holders_internal_owner_identity_holder", DROP COLUMN "reviewed_by_identity_holder_id", DROP COLUMN "internal_owner_identity_holder_id";
-- reverse: modify "campaigns" table
ALTER TABLE "campaigns" DROP CONSTRAINT "campaigns_identity_holders_internal_owner_identity_holder", DROP COLUMN "internal_owner_identity_holder_id";
-- reverse: modify "assets" table
ALTER TABLE "assets" DROP CONSTRAINT "assets_identity_holders_internal_owner_identity_holder", DROP COLUMN "internal_owner_identity_holder_id";
-- reverse: create index "identity_holders_internal_owner_identity_holder_id_key" to table: "identity_holders"
DROP INDEX "identity_holders_internal_owner_identity_holder_id_key";
-- reverse: modify "identity_holders" table
ALTER TABLE "identity_holders" DROP CONSTRAINT "identity_holders_identity_hold_a071c96868edb317166fe843ed012736", DROP COLUMN "internal_owner_identity_holder_id";
