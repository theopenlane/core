-- Modify "group_history" table
ALTER TABLE "group_history" ADD COLUMN "is_managed" boolean NULL DEFAULT false;
-- Modify "groups" table
ALTER TABLE "groups" ADD COLUMN "is_managed" boolean NULL DEFAULT false;
