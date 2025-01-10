-- Modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" ADD COLUMN "product_price" jsonb NULL;
-- Modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" ADD COLUMN "product_price" jsonb NULL;
