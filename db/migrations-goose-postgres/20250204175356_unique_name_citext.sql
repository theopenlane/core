-- +goose Up
-- modify "entities" table
ALTER TABLE "entities" ALTER COLUMN "name" TYPE citext;
-- modify "entity_history" table
ALTER TABLE "entity_history" ALTER COLUMN "name" TYPE citext;
-- modify "entity_type_history" table
ALTER TABLE "entity_type_history" ALTER COLUMN "name" TYPE citext;
-- modify "entity_types" table
ALTER TABLE "entity_types" ALTER COLUMN "name" TYPE citext;
-- modify "group_history" table
ALTER TABLE "group_history" ALTER COLUMN "name" TYPE citext;
-- modify "groups" table
ALTER TABLE "groups" ALTER COLUMN "name" TYPE citext;
-- modify "organization_history" table
ALTER TABLE "organization_history" ALTER COLUMN "name" TYPE citext;
-- modify "organizations" table
ALTER TABLE "organizations" ALTER COLUMN "name" TYPE citext;

-- +goose Down
-- reverse: modify "organizations" table
ALTER TABLE "organizations" ALTER COLUMN "name" TYPE character varying;
-- reverse: modify "organization_history" table
ALTER TABLE "organization_history" ALTER COLUMN "name" TYPE character varying;
-- reverse: modify "groups" table
ALTER TABLE "groups" ALTER COLUMN "name" TYPE character varying;
-- reverse: modify "group_history" table
ALTER TABLE "group_history" ALTER COLUMN "name" TYPE character varying;
-- reverse: modify "entity_types" table
ALTER TABLE "entity_types" ALTER COLUMN "name" TYPE character varying;
-- reverse: modify "entity_type_history" table
ALTER TABLE "entity_type_history" ALTER COLUMN "name" TYPE character varying;
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" ALTER COLUMN "name" TYPE character varying;
-- reverse: modify "entities" table
ALTER TABLE "entities" ALTER COLUMN "name" TYPE character varying;
