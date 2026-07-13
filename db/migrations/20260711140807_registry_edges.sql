-- Modify "entities" table
ALTER TABLE "entities" DROP COLUMN "scan_entities";
-- Modify "system_details" table
ALTER TABLE "system_details" DROP COLUMN "platform_id", DROP COLUMN "program_id";
-- Create "entity_system_details" table
CREATE TABLE "entity_system_details" ("entity_id" character varying NOT NULL, "system_detail_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "system_detail_id"), CONSTRAINT "entity_system_details_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "entity_system_details_system_detail_id" FOREIGN KEY ("system_detail_id") REFERENCES "system_details" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Modify "scans" table
ALTER TABLE "scans" DROP COLUMN "entity_scans", DROP COLUMN "finding_scans";
-- Create "finding_scans" table
CREATE TABLE "finding_scans" ("finding_id" character varying NOT NULL, "scan_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "scan_id"), CONSTRAINT "finding_scans_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "finding_scans_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "platform_system_details" table
CREATE TABLE "platform_system_details" ("platform_id" character varying NOT NULL, "system_detail_id" character varying NOT NULL, PRIMARY KEY ("platform_id", "system_detail_id"), CONSTRAINT "platform_system_details_platform_id" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "platform_system_details_system_detail_id" FOREIGN KEY ("system_detail_id") REFERENCES "system_details" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "program_system_details" table
CREATE TABLE "program_system_details" ("program_id" character varying NOT NULL, "system_detail_id" character varying NOT NULL, PRIMARY KEY ("program_id", "system_detail_id"), CONSTRAINT "program_system_details_program_id" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "program_system_details_system_detail_id" FOREIGN KEY ("system_detail_id") REFERENCES "system_details" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "scan_entities" table
CREATE TABLE "scan_entities" ("scan_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "entity_id"), CONSTRAINT "scan_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "scan_entities_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "system_detail_assets" table
CREATE TABLE "system_detail_assets" ("system_detail_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("system_detail_id", "asset_id"), CONSTRAINT "system_detail_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "system_detail_assets_system_detail_id" FOREIGN KEY ("system_detail_id") REFERENCES "system_details" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
