-- +goose Up
-- modify "onboardings" table
ALTER TABLE "onboardings" ADD COLUMN "demo_requested" boolean NULL DEFAULT false;

-- +goose Down
-- reverse: modify "onboardings" table
ALTER TABLE "onboardings" DROP COLUMN "demo_requested";
