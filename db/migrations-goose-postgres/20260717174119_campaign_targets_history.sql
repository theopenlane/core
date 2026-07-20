-- +goose Up
-- modify "campaign_target_history" table
ALTER TABLE "campaign_target_history" ALTER COLUMN "campaign_id" DROP NOT NULL;

-- +goose Down
-- reverse: modify "campaign_target_history" table
ALTER TABLE "campaign_target_history" ALTER COLUMN "campaign_id" SET NOT NULL;
