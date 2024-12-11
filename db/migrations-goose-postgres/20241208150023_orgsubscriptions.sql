-- +goose Up
-- create "org_subscription_history" table
CREATE TABLE "org_subscription_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "stripe_subscription_id" character varying NULL, "product_tier" character varying NULL, "stripe_product_tier_id" character varying NULL, "stripe_subscription_status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "stripe_customer_id" character varying NULL, "expires_at" timestamptz NULL, "features" jsonb NULL, PRIMARY KEY ("id"));
-- create index "orgsubscriptionhistory_history_time" to table: "org_subscription_history"
CREATE INDEX "orgsubscriptionhistory_history_time" ON "org_subscription_history" ("history_time");
-- create "org_subscriptions" table
CREATE TABLE "org_subscriptions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "mapping_id" character varying NOT NULL, "tags" jsonb NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "stripe_subscription_id" character varying NULL, "product_tier" character varying NULL, "stripe_product_tier_id" character varying NULL, "stripe_subscription_status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "stripe_customer_id" character varying NULL, "expires_at" timestamptz NULL, "features" jsonb NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "org_subscriptions_organizations_orgsubscriptions" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "org_subscriptions_mapping_id_key" to table: "org_subscriptions"
CREATE UNIQUE INDEX "org_subscriptions_mapping_id_key" ON "org_subscriptions" ("mapping_id");
-- create index "org_subscriptions_stripe_customer_id_key" to table: "org_subscriptions"
CREATE UNIQUE INDEX "org_subscriptions_stripe_customer_id_key" ON "org_subscriptions" ("stripe_customer_id");

-- +goose Down
-- reverse: create index "org_subscriptions_stripe_customer_id_key" to table: "org_subscriptions"
DROP INDEX "org_subscriptions_stripe_customer_id_key";
-- reverse: create index "org_subscriptions_mapping_id_key" to table: "org_subscriptions"
DROP INDEX "org_subscriptions_mapping_id_key";
-- reverse: create "org_subscriptions" table
DROP TABLE "org_subscriptions";
-- reverse: create index "orgsubscriptionhistory_history_time" to table: "org_subscription_history"
DROP INDEX "orgsubscriptionhistory_history_time";
-- reverse: create "org_subscription_history" table
DROP TABLE "org_subscription_history";
