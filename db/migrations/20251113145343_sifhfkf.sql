-- Modify "assessment_history" table
ALTER TABLE "assessment_history" DROP COLUMN "assessment_owner_id", ADD COLUMN "response_due_duration" bigint NOT NULL DEFAULT 604800;
-- Modify "assessments" table
ALTER TABLE "assessments" DROP COLUMN "assessment_owner_id", ADD COLUMN "response_due_duration" bigint NOT NULL DEFAULT 604800;
-- Modify "groups" table
ALTER TABLE "groups" DROP COLUMN "assessment_blocked_groups", DROP COLUMN "assessment_editors", DROP COLUMN "assessment_viewers";
