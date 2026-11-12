-- Create "risk_entities" table
CREATE TABLE "risk_entities" ("risk_id" character varying NOT NULL, "entity_id" character varying NOT NULL, PRIMARY KEY ("risk_id", "entity_id"), CONSTRAINT "risk_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "risk_entities_risk_id" FOREIGN KEY ("risk_id") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "risk_entities_entity_id_idx" to table: "risk_entities"
CREATE INDEX "risk_entities_entity_id_idx" ON "risk_entities" ("entity_id");
