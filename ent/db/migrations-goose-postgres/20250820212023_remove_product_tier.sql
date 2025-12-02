-- +goose Up
-- modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "product_tier", DROP COLUMN "stripe_product_tier_id";
-- modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "product_tier", DROP COLUMN "stripe_product_tier_id";

-- +goose Down
-- reverse: modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" ADD COLUMN "stripe_product_tier_id" character varying NULL, ADD COLUMN "product_tier" character varying NULL;
-- reverse: modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" ADD COLUMN "stripe_product_tier_id" character varying NULL, ADD COLUMN "product_tier" character varying NULL;
