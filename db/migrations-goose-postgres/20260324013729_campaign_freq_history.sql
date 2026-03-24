-- +goose Up
-- modify "campaign_history" table
ALTER TABLE "campaign_history" ALTER COLUMN "recurrence_frequency" SET DEFAULT 'NONE';

-- +goose Down
-- reverse: modify "campaign_history" table
ALTER TABLE "campaign_history" ALTER COLUMN "recurrence_frequency" DROP DEFAULT;
