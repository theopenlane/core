-- Modify "campaign_targets" table
ALTER TABLE "campaign_targets" DROP CONSTRAINT "campaign_targets_campaigns_campaign_targets", ALTER COLUMN "campaign_id" DROP NOT NULL, ADD CONSTRAINT "campaign_targets_campaigns_campaign_targets" FOREIGN KEY ("campaign_id") REFERENCES "campaigns" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
