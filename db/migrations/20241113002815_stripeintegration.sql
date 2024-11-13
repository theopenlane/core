-- Modify "entitlement_history" table
ALTER TABLE "entitlement_history" ADD COLUMN "cancelled_date" timestamptz NULL, ADD COLUMN "bill_starting" timestamptz NOT NULL, ADD COLUMN "active" boolean NOT NULL DEFAULT true;
-- Modify "entitlement_plan_feature_history" table
ALTER TABLE "entitlement_plan_feature_history" ADD COLUMN "stripe_product_id" character varying NULL, ADD COLUMN "stripe_feature_id" character varying NULL;
-- Modify "entitlement_plan_features" table
ALTER TABLE "entitlement_plan_features" ADD COLUMN "stripe_product_id" character varying NULL, ADD COLUMN "stripe_feature_id" character varying NULL;
-- Modify "entitlement_plan_history" table
ALTER TABLE "entitlement_plan_history" ADD COLUMN "stripe_product_id" character varying NULL, ADD COLUMN "stripe_price_id" character varying NULL;
-- Modify "entitlement_plans" table
ALTER TABLE "entitlement_plans" ADD COLUMN "stripe_product_id" character varying NULL, ADD COLUMN "stripe_price_id" character varying NULL;
-- Modify "entitlements" table
ALTER TABLE "entitlements" ADD COLUMN "cancelled_date" timestamptz NULL, ADD COLUMN "bill_starting" timestamptz NOT NULL, ADD COLUMN "active" boolean NOT NULL DEFAULT true;
-- Modify "feature_history" table
ALTER TABLE "feature_history" ADD COLUMN "stripe_feature_id" character varying NULL;
-- Modify "features" table
ALTER TABLE "features" ADD COLUMN "stripe_feature_id" character varying NULL;
-- Modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "stripe_id" character varying NULL;
-- Modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "stripe_id" character varying NULL;
