-- Drop index "assessment_name_owner_id" from table: "assessments"
DROP INDEX "assessment_name_owner_id";
-- Create index "assessment_name_owner_id" to table: "assessments"
CREATE INDEX "assessment_name_owner_id" ON "assessments" ("name", "owner_id") WHERE (deleted_at IS NULL);
