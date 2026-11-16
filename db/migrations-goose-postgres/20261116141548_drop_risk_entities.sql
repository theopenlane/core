-- +goose Up
-- modify "entities" table
ALTER TABLE "entities" DROP COLUMN "risk_entities";

-- +goose Down
-- reverse: modify "entities" table
ALTER TABLE "entities" ADD COLUMN "risk_entities" character varying NULL;
