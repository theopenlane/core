-- +goose Up
-- modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" ADD COLUMN "product_price" jsonb NULL;
-- modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" ADD COLUMN "product_price" jsonb NULL;

-- +goose Down
-- reverse: modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "product_price";
-- reverse: modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "product_price";
