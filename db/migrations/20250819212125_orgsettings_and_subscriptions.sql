-- Modify "org_subscription_history" table
ALTER TABLE "org_subscription_history" DROP COLUMN "stripe_customer_id", DROP COLUMN "payment_method_added";
-- Modify "org_subscriptions" table
ALTER TABLE "org_subscriptions" DROP COLUMN "stripe_customer_id", DROP COLUMN "payment_method_added";
-- Modify "organization_history" table
ALTER TABLE "organization_history" ADD COLUMN "stripe_customer_id" character varying NULL;
-- Modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "payment_method_added" boolean NOT NULL DEFAULT false;
-- Modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "payment_method_added" boolean NOT NULL DEFAULT false;
-- Modify "organizations" table
ALTER TABLE "organizations" ADD COLUMN "stripe_customer_id" character varying NULL;
-- Create index "organizations_stripe_customer_id_key" to table: "organizations"
CREATE UNIQUE INDEX "organizations_stripe_customer_id_key" ON "organizations" ("stripe_customer_id");
