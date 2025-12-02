-- Modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "product_tier", DROP COLUMN "stripe_product_tier_id";
-- Modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "product_tier", DROP COLUMN "stripe_product_tier_id";
