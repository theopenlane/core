-- +goose Up
-- modify "onboardings" table
ALTER TABLE "onboardings" ADD COLUMN "demo_dataset" character varying NULL DEFAULT 'soc2';

-- +goose Down
-- reverse: modify "onboardings" table
ALTER TABLE "onboardings" DROP COLUMN "demo_dataset";
