-- +goose Up
-- modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "stripe_customer_id", DROP COLUMN "payment_method_added";
-- modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "stripe_customer_id", DROP COLUMN "payment_method_added";
-- modify "organization_history" table
ALTER TABLE "organization_history" ADD COLUMN "stripe_customer_id" character varying NULL;
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "payment_method_added" boolean NOT NULL DEFAULT false;
-- modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "payment_method_added" boolean NOT NULL DEFAULT false;
-- modify "organizations" table
ALTER TABLE "organizations" ADD COLUMN "stripe_customer_id" character varying NULL;
-- create index "organizations_stripe_customer_id_key" to table: "organizations"
CREATE UNIQUE INDEX "organizations_stripe_customer_id_key" ON "organizations" ("stripe_customer_id");

-- +goose Down
-- reverse: create index "organizations_stripe_customer_id_key" to table: "organizations"
DROP INDEX "organizations_stripe_customer_id_key";
-- reverse: modify "organizations" table
ALTER TABLE "organizations" DROP COLUMN "stripe_customer_id";
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "payment_method_added";
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "payment_method_added";
-- reverse: modify "organization_history" table
ALTER TABLE "organization_history" DROP COLUMN "stripe_customer_id";
-- reverse: modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" ADD COLUMN "payment_method_added" boolean NULL, ADD COLUMN "stripe_customer_id" character varying NULL;
-- reverse: modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" ADD COLUMN "payment_method_added" boolean NULL, ADD COLUMN "stripe_customer_id" character varying NULL;
