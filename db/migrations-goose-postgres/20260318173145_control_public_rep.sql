-- +goose Up
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "public_representation" text NULL, ADD COLUMN "source_name" character varying NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "public_representation" text NULL, ADD COLUMN "source_name" character varying NULL;

-- +goose Down
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "source_name", DROP COLUMN "public_representation";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "source_name", DROP COLUMN "public_representation";
