-- Modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "title" character varying NULL;
-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "title" character varying NULL;
-- Modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "product_price", DROP COLUMN "features", DROP COLUMN "feature_lookup_keys";
-- Modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "product_price", DROP COLUMN "features", DROP COLUMN "feature_lookup_keys";
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "title" character varying NULL;
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "title" character varying NULL;
