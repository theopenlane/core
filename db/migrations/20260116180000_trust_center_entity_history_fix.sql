-- Drop legacy table created by early history migration
DROP TABLE "trustcenter_entity_history";
-- Create history_time index on correctly named table
CREATE INDEX "trustcenterentityhistory_history_time" ON "trust_center_entity_history" ("history_time");
