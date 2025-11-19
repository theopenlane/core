-- +goose Up
-- modify "groups" table
ALTER TABLE "groups" DROP COLUMN "organization_assessment_creators";

-- +goose Down
-- reverse: modify "groups" table
ALTER TABLE "groups" ADD COLUMN "organization_assessment_creators" character varying NULL;
