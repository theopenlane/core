-- Drop index "org_subscriptions_stripe_customer_id_key" from table: "org_subscriptions"
DROP INDEX "org_subscriptions_stripe_customer_id_key";
-- Modify "organization_history" table
ALTER TABLE "organization_history" ADD COLUMN "stripe_customer_id" character varying NULL;
-- Modify "organizations" table
ALTER TABLE "organizations" ADD COLUMN "stripe_customer_id" character varying NULL;
-- Create index "organizations_stripe_customer_id_key" to table: "organizations"
CREATE UNIQUE INDEX "organizations_stripe_customer_id_key" ON "organizations" ("stripe_customer_id");
