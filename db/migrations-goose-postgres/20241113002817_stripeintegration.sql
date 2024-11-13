-- +goose Up
-- modify "entitlement_history" table
ALTER TABLE "entitlement_history" ADD COLUMN "cancelled_date" timestamptz NULL, ADD COLUMN "bill_starting" timestamptz NOT NULL, ADD COLUMN "active" boolean NOT NULL DEFAULT true;
-- modify "entitlement_plan_feature_history" table
ALTER TABLE "entitlement_plan_feature_history" ADD COLUMN "stripe_product_id" character varying NULL, ADD COLUMN "stripe_feature_id" character varying NULL;
-- modify "entitlement_plan_features" table
ALTER TABLE "entitlement_plan_features" ADD COLUMN "stripe_product_id" character varying NULL, ADD COLUMN "stripe_feature_id" character varying NULL;
-- modify "entitlement_plan_history" table
ALTER TABLE "entitlement_plan_history" ADD COLUMN "stripe_product_id" character varying NULL, ADD COLUMN "stripe_price_id" character varying NULL;
-- modify "entitlement_plans" table
ALTER TABLE "entitlement_plans" ADD COLUMN "stripe_product_id" character varying NULL, ADD COLUMN "stripe_price_id" character varying NULL;
-- modify "entitlements" table
ALTER TABLE "entitlements" ADD COLUMN "cancelled_date" timestamptz NULL, ADD COLUMN "bill_starting" timestamptz NOT NULL, ADD COLUMN "active" boolean NOT NULL DEFAULT true;
-- modify "feature_history" table
ALTER TABLE "feature_history" ADD COLUMN "stripe_feature_id" character varying NULL;
-- modify "features" table
ALTER TABLE "features" ADD COLUMN "stripe_feature_id" character varying NULL;
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "stripe_id" character varying NULL;
-- modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "stripe_id" character varying NULL;

-- +goose Down
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "stripe_id";
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "stripe_id";
-- reverse: modify "features" table
ALTER TABLE "features" DROP COLUMN "stripe_feature_id";
-- reverse: modify "feature_history" table
ALTER TABLE "feature_history" DROP COLUMN "stripe_feature_id";
-- reverse: modify "entitlements" table
ALTER TABLE "entitlements" DROP COLUMN "active", DROP COLUMN "bill_starting", DROP COLUMN "cancelled_date";
-- reverse: modify "entitlement_plans" table
ALTER TABLE "entitlement_plans" DROP COLUMN "stripe_price_id", DROP COLUMN "stripe_product_id";
-- reverse: modify "entitlement_plan_history" table
ALTER TABLE "entitlement_plan_history" DROP COLUMN "stripe_price_id", DROP COLUMN "stripe_product_id";
-- reverse: modify "entitlement_plan_features" table
ALTER TABLE "entitlement_plan_features" DROP COLUMN "stripe_feature_id", DROP COLUMN "stripe_product_id";
-- reverse: modify "entitlement_plan_feature_history" table
ALTER TABLE "entitlement_plan_feature_history" DROP COLUMN "stripe_feature_id", DROP COLUMN "stripe_product_id";
-- reverse: modify "entitlement_history" table
ALTER TABLE "entitlement_history" DROP COLUMN "active", DROP COLUMN "bill_starting", DROP COLUMN "cancelled_date";
