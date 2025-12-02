-- +goose Up
-- modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "title" character varying NULL;
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "title" character varying NULL;
-- modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "product_price", DROP COLUMN "features", DROP COLUMN "feature_lookup_keys";
-- modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "product_price", DROP COLUMN "features", DROP COLUMN "feature_lookup_keys";
-- modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "title" character varying NULL;
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "title" character varying NULL;

-- +goose Down
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "title";
-- reverse: modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" DROP COLUMN "title";
-- reverse: modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" ADD COLUMN "feature_lookup_keys" jsonb NULL, ADD COLUMN "features" jsonb NULL, ADD COLUMN "product_price" jsonb NULL;
-- reverse: modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" ADD COLUMN "feature_lookup_keys" jsonb NULL, ADD COLUMN "features" jsonb NULL, ADD COLUMN "product_price" jsonb NULL;
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "title";
-- reverse: modify "control_history" table
ALTER TABLE "control_history" DROP COLUMN "title";
