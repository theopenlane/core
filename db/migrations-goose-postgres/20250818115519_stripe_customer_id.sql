-- +goose Up
-- drop index "org_subscriptions_stripe_customer_id_key" from table: "org_subscriptions"
DROP INDEX "org_subscriptions_stripe_customer_id_key";
-- modify "organization_history" table
ALTER TABLE "organization_history" ADD COLUMN "stripe_customer_id" character varying NULL;
-- modify "organizations" table
ALTER TABLE "organizations" ADD COLUMN "stripe_customer_id" character varying NULL;
-- create index "organizations_stripe_customer_id_key" to table: "organizations"
CREATE UNIQUE INDEX "organizations_stripe_customer_id_key" ON "organizations" ("stripe_customer_id");

-- +goose Down
-- reverse: create index "organizations_stripe_customer_id_key" to table: "organizations"
DROP INDEX "organizations_stripe_customer_id_key";
-- reverse: modify "organizations" table
ALTER TABLE "organizations" DROP COLUMN "stripe_customer_id";
-- reverse: modify "organization_history" table
ALTER TABLE "organization_history" DROP COLUMN "stripe_customer_id";
-- reverse: drop index "org_subscriptions_stripe_customer_id_key" from table: "org_subscriptions"
CREATE UNIQUE INDEX "org_subscriptions_stripe_customer_id_key" ON "org_subscriptions" ("stripe_customer_id");
