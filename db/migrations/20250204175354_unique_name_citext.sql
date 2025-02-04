-- Modify "entities" table
ALTER TABLE "entities" ALTER COLUMN "name" TYPE citext;
-- Modify "entity_history" table
ALTER TABLE "entity_history" ALTER COLUMN "name" TYPE citext;
-- Modify "entity_type_history" table
ALTER TABLE "entity_type_history" ALTER COLUMN "name" TYPE citext;
-- Modify "entity_types" table
ALTER TABLE "entity_types" ALTER COLUMN "name" TYPE citext;
-- Modify "group_history" table
ALTER TABLE "group_history" ALTER COLUMN "name" TYPE citext;
-- Modify "groups" table
ALTER TABLE "groups" ALTER COLUMN "name" TYPE citext;
-- Modify "organization_history" table
ALTER TABLE "organization_history" ALTER COLUMN "name" TYPE citext;
-- Modify "organizations" table
ALTER TABLE "organizations" ALTER COLUMN "name" TYPE citext;
