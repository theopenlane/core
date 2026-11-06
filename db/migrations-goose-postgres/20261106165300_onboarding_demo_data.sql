-- +goose Up
-- modify "onboardings" table
ALTER TABLE "onboardings" ADD COLUMN "seed_demo_data" boolean NULL;

-- +goose Down
-- reverse: modify "onboardings" table
ALTER TABLE "onboardings" DROP COLUMN "seed_demo_data";
