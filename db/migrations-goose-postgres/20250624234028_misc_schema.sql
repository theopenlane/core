-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "tag_suggestions" jsonb NULL, ADD COLUMN "dismissed_tag_suggestions" jsonb NULL, ADD COLUMN "control_suggestions" jsonb NULL, ADD COLUMN "dismissed_control_suggestions" jsonb NULL, ADD COLUMN "improvement_suggestions" jsonb NULL, ADD COLUMN "dismissed_improvement_suggestions" jsonb NULL;
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "tag_suggestions" jsonb NULL, ADD COLUMN "dismissed_tag_suggestions" jsonb NULL, ADD COLUMN "control_suggestions" jsonb NULL, ADD COLUMN "dismissed_control_suggestions" jsonb NULL, ADD COLUMN "improvement_suggestions" jsonb NULL, ADD COLUMN "dismissed_improvement_suggestions" jsonb NULL;
-- modify "entities" table
ALTER TABLE "entities" ADD COLUMN "risk_entities" character varying NULL, ADD COLUMN "scan_entities" character varying NULL;
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "asset_blocked_groups" character varying NULL, ADD COLUMN "asset_editors" character varying NULL, ADD COLUMN "asset_viewers" character varying NULL, ADD COLUMN "entity_blocked_groups" character varying NULL, ADD COLUMN "entity_editors" character varying NULL, ADD COLUMN "entity_viewers" character varying NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "tag_suggestions" jsonb NULL, ADD COLUMN "dismissed_tag_suggestions" jsonb NULL, ADD COLUMN "control_suggestions" jsonb NULL, ADD COLUMN "dismissed_control_suggestions" jsonb NULL, ADD COLUMN "improvement_suggestions" jsonb NULL, ADD COLUMN "dismissed_improvement_suggestions" jsonb NULL;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "tag_suggestions" jsonb NULL, ADD COLUMN "dismissed_tag_suggestions" jsonb NULL, ADD COLUMN "control_suggestions" jsonb NULL, ADD COLUMN "dismissed_control_suggestions" jsonb NULL, ADD COLUMN "improvement_suggestions" jsonb NULL, ADD COLUMN "dismissed_improvement_suggestions" jsonb NULL;
-- modify "org_modules" table
ALTER TABLE "org_modules" DROP COLUMN "trial_expires_at", DROP COLUMN "expires_at", ADD COLUMN "visibility" character varying NULL, ADD COLUMN "module_lookup_key" character varying NULL, ADD COLUMN "price_id" character varying NULL, ADD COLUMN "org_product_org_modules" character varying NULL;
-- modify "org_prices" table
ALTER TABLE "org_prices" DROP CONSTRAINT "org_prices_org_products_prices", DROP COLUMN "trial_expires_at", DROP COLUMN "expires_at", ADD COLUMN "subscription_id" character varying NULL;
-- modify "org_products" table
ALTER TABLE "org_products" DROP COLUMN "trial_expires_at", DROP COLUMN "expires_at", ADD COLUMN "price_id" character varying NULL, ADD COLUMN "org_module_org_products" character varying NULL;
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "identity_provider" character varying NULL DEFAULT 'NONE', ADD COLUMN "identity_provider_client_id" character varying NULL, ADD COLUMN "identity_provider_client_secret" character varying NULL, ADD COLUMN "identity_provider_metadata_endpoint" character varying NULL, ADD COLUMN "identity_provider_entity_id" character varying NULL, ADD COLUMN "oidc_discovery_endpoint" character varying NULL, ADD COLUMN "identity_provider_login_enforced" boolean NOT NULL DEFAULT false, ADD COLUMN "compliance_webhook_token" character varying NULL;
-- modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "identity_provider" character varying NULL DEFAULT 'NONE', ADD COLUMN "identity_provider_client_id" character varying NULL, ADD COLUMN "identity_provider_client_secret" character varying NULL, ADD COLUMN "identity_provider_metadata_endpoint" character varying NULL, ADD COLUMN "identity_provider_entity_id" character varying NULL, ADD COLUMN "oidc_discovery_endpoint" character varying NULL, ADD COLUMN "identity_provider_login_enforced" boolean NOT NULL DEFAULT false, ADD COLUMN "compliance_webhook_token" character varying NULL;
-- create index "organization_settings_compliance_webhook_token_key" to table: "organization_settings"
CREATE UNIQUE INDEX "organization_settings_compliance_webhook_token_key" ON "organization_settings" ("compliance_webhook_token");
-- modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "tag_suggestions" jsonb NULL, ADD COLUMN "dismissed_tag_suggestions" jsonb NULL, ADD COLUMN "control_suggestions" jsonb NULL, ADD COLUMN "dismissed_control_suggestions" jsonb NULL, ADD COLUMN "improvement_suggestions" jsonb NULL, ADD COLUMN "dismissed_improvement_suggestions" jsonb NULL;
-- modify "procedures" table
ALTER TABLE "procedures" ADD COLUMN "tag_suggestions" jsonb NULL, ADD COLUMN "dismissed_tag_suggestions" jsonb NULL, ADD COLUMN "control_suggestions" jsonb NULL, ADD COLUMN "dismissed_control_suggestions" jsonb NULL, ADD COLUMN "improvement_suggestions" jsonb NULL, ADD COLUMN "dismissed_improvement_suggestions" jsonb NULL;
-- create "assets" table
CREATE TABLE "assets" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "asset_type" character varying NOT NULL DEFAULT 'TECHNOLOGY', "name" character varying NOT NULL, "description" character varying NULL, "identifier" character varying NULL, "website" character varying NULL, "cpe" character varying NULL, "categories" jsonb NULL, "owner_id" character varying NULL, "risk_assets" character varying NULL, PRIMARY KEY ("id"));
-- create index "asset_id" to table: "assets"
CREATE UNIQUE INDEX "asset_id" ON "assets" ("id");
-- create index "asset_name_owner_id" to table: "assets"
CREATE UNIQUE INDEX "asset_name_owner_id" ON "assets" ("name", "owner_id") WHERE (deleted_at IS NULL);
-- create index "asset_owner_id" to table: "assets"
CREATE INDEX "asset_owner_id" ON "assets" ("owner_id") WHERE (deleted_at IS NULL);
-- create "asset_history" table
CREATE TABLE "asset_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "asset_type" character varying NOT NULL DEFAULT 'TECHNOLOGY', "name" character varying NOT NULL, "description" character varying NULL, "identifier" character varying NULL, "website" character varying NULL, "cpe" character varying NULL, "categories" jsonb NULL, PRIMARY KEY ("id"));
-- create index "assethistory_history_time" to table: "asset_history"
CREATE INDEX "assethistory_history_time" ON "asset_history" ("history_time");
-- create "scans" table
CREATE TABLE "scans" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "target" character varying NOT NULL, "scan_type" character varying NOT NULL DEFAULT 'DOMAIN', "metadata" jsonb NULL, "status" character varying NOT NULL DEFAULT 'PROCESSING', "control_scans" character varying NULL, "entity_scans" character varying NULL, "owner_id" character varying NULL, "risk_scans" character varying NULL, PRIMARY KEY ("id"));
-- create index "scan_id" to table: "scans"
CREATE UNIQUE INDEX "scan_id" ON "scans" ("id");
-- create index "scan_owner_id" to table: "scans"
CREATE INDEX "scan_owner_id" ON "scans" ("owner_id") WHERE (deleted_at IS NULL);
-- create "scan_history" table
CREATE TABLE "scan_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "target" character varying NOT NULL, "scan_type" character varying NOT NULL DEFAULT 'DOMAIN', "metadata" jsonb NULL, "status" character varying NOT NULL DEFAULT 'PROCESSING', PRIMARY KEY ("id"));
-- create index "scanhistory_history_time" to table: "scan_history"
CREATE INDEX "scanhistory_history_time" ON "scan_history" ("history_time");
-- create "control_assets" table
CREATE TABLE "control_assets" ("control_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("control_id", "asset_id"));
-- create "entity_assets" table
CREATE TABLE "entity_assets" ("entity_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("entity_id", "asset_id"));
-- create "org_module_org_prices" table
CREATE TABLE "org_module_org_prices" ("org_module_id" character varying NOT NULL, "org_price_id" character varying NOT NULL, PRIMARY KEY ("org_module_id", "org_price_id"));
-- create "org_product_org_prices" table
CREATE TABLE "org_product_org_prices" ("org_product_id" character varying NOT NULL, "org_price_id" character varying NOT NULL, PRIMARY KEY ("org_product_id", "org_price_id"));
-- create "scan_blocked_groups" table
CREATE TABLE "scan_blocked_groups" ("scan_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "group_id"));
-- create "scan_editors" table
CREATE TABLE "scan_editors" ("scan_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "group_id"));
-- create "scan_viewers" table
CREATE TABLE "scan_viewers" ("scan_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "group_id"));
-- create "scan_assets" table
CREATE TABLE "scan_assets" ("scan_id" character varying NOT NULL, "asset_id" character varying NOT NULL, PRIMARY KEY ("scan_id", "asset_id"));
-- modify "entities" table
ALTER TABLE "entities" ADD CONSTRAINT "entities_risks_entities" FOREIGN KEY ("risk_entities") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "entities_scans_entities" FOREIGN KEY ("scan_entities") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "groups" table
ALTER TABLE "groups" ADD CONSTRAINT "groups_assets_blocked_groups" FOREIGN KEY ("asset_blocked_groups") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_assets_editors" FOREIGN KEY ("asset_editors") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_assets_viewers" FOREIGN KEY ("asset_viewers") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_entities_blocked_groups" FOREIGN KEY ("entity_blocked_groups") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_entities_editors" FOREIGN KEY ("entity_editors") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_entities_viewers" FOREIGN KEY ("entity_viewers") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "org_modules" table
ALTER TABLE "org_modules" ADD CONSTRAINT "org_modules_org_products_org_modules" FOREIGN KEY ("org_product_org_modules") REFERENCES "org_products" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "org_prices" table
ALTER TABLE "org_prices" ADD CONSTRAINT "org_prices_org_subscriptions_prices" FOREIGN KEY ("subscription_id") REFERENCES "org_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "org_products" table
ALTER TABLE "org_products" ADD CONSTRAINT "org_products_org_modules_org_products" FOREIGN KEY ("org_module_org_products") REFERENCES "org_modules" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "assets" table
ALTER TABLE "assets" ADD CONSTRAINT "assets_organizations_assets" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "assets_risks_assets" FOREIGN KEY ("risk_assets") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "scans" table
ALTER TABLE "scans" ADD CONSTRAINT "scans_controls_scans" FOREIGN KEY ("control_scans") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_entities_scans" FOREIGN KEY ("entity_scans") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_organizations_scans" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "scans_risks_scans" FOREIGN KEY ("risk_scans") REFERENCES "risks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "control_assets" table
ALTER TABLE "control_assets" ADD CONSTRAINT "control_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "control_assets_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "entity_assets" table
ALTER TABLE "entity_assets" ADD CONSTRAINT "entity_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "entity_assets_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "org_module_org_prices" table
ALTER TABLE "org_module_org_prices" ADD CONSTRAINT "org_module_org_prices_org_module_id" FOREIGN KEY ("org_module_id") REFERENCES "org_modules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "org_module_org_prices_org_price_id" FOREIGN KEY ("org_price_id") REFERENCES "org_prices" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "org_product_org_prices" table
ALTER TABLE "org_product_org_prices" ADD CONSTRAINT "org_product_org_prices_org_price_id" FOREIGN KEY ("org_price_id") REFERENCES "org_prices" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "org_product_org_prices_org_product_id" FOREIGN KEY ("org_product_id") REFERENCES "org_products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "scan_blocked_groups" table
ALTER TABLE "scan_blocked_groups" ADD CONSTRAINT "scan_blocked_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_blocked_groups_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "scan_editors" table
ALTER TABLE "scan_editors" ADD CONSTRAINT "scan_editors_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_editors_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "scan_viewers" table
ALTER TABLE "scan_viewers" ADD CONSTRAINT "scan_viewers_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_viewers_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "scan_assets" table
ALTER TABLE "scan_assets" ADD CONSTRAINT "scan_assets_asset_id" FOREIGN KEY ("asset_id") REFERENCES "assets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "scan_assets_scan_id" FOREIGN KEY ("scan_id") REFERENCES "scans" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "scan_assets" table
ALTER TABLE "scan_assets" DROP CONSTRAINT "scan_assets_scan_id", DROP CONSTRAINT "scan_assets_asset_id";
-- reverse: modify "scan_viewers" table
ALTER TABLE "scan_viewers" DROP CONSTRAINT "scan_viewers_scan_id", DROP CONSTRAINT "scan_viewers_group_id";
-- reverse: modify "scan_editors" table
ALTER TABLE "scan_editors" DROP CONSTRAINT "scan_editors_scan_id", DROP CONSTRAINT "scan_editors_group_id";
-- reverse: modify "scan_blocked_groups" table
ALTER TABLE "scan_blocked_groups" DROP CONSTRAINT "scan_blocked_groups_scan_id", DROP CONSTRAINT "scan_blocked_groups_group_id";
-- reverse: modify "org_product_org_prices" table
ALTER TABLE "org_product_org_prices" DROP CONSTRAINT "org_product_org_prices_org_product_id", DROP CONSTRAINT "org_product_org_prices_org_price_id";
-- reverse: modify "org_module_org_prices" table
ALTER TABLE "org_module_org_prices" DROP CONSTRAINT "org_module_org_prices_org_price_id", DROP CONSTRAINT "org_module_org_prices_org_module_id";
-- reverse: modify "entity_assets" table
ALTER TABLE "entity_assets" DROP CONSTRAINT "entity_assets_entity_id", DROP CONSTRAINT "entity_assets_asset_id";
-- reverse: modify "control_assets" table
ALTER TABLE "control_assets" DROP CONSTRAINT "control_assets_control_id", DROP CONSTRAINT "control_assets_asset_id";
-- reverse: modify "scans" table
ALTER TABLE "scans" DROP CONSTRAINT "scans_risks_scans", DROP CONSTRAINT "scans_organizations_scans", DROP CONSTRAINT "scans_entities_scans", DROP CONSTRAINT "scans_controls_scans";
-- reverse: modify "assets" table
ALTER TABLE "assets" DROP CONSTRAINT "assets_risks_assets", DROP CONSTRAINT "assets_organizations_assets";
-- reverse: modify "org_products" table
ALTER TABLE "org_products" DROP CONSTRAINT "org_products_org_modules_org_products";
-- reverse: modify "org_prices" table
ALTER TABLE "org_prices" DROP CONSTRAINT "org_prices_org_subscriptions_prices";
-- reverse: modify "org_modules" table
ALTER TABLE "org_modules" DROP CONSTRAINT "org_modules_org_products_org_modules";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_entities_viewers", DROP CONSTRAINT "groups_entities_editors", DROP CONSTRAINT "groups_entities_blocked_groups", DROP CONSTRAINT "groups_assets_viewers", DROP CONSTRAINT "groups_assets_editors", DROP CONSTRAINT "groups_assets_blocked_groups";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP CONSTRAINT "entities_scans_entities", DROP CONSTRAINT "entities_risks_entities";
-- reverse: create "scan_assets" table
DROP TABLE "scan_assets";
-- reverse: create "scan_viewers" table
DROP TABLE "scan_viewers";
-- reverse: create "scan_editors" table
DROP TABLE "scan_editors";
-- reverse: create "scan_blocked_groups" table
DROP TABLE "scan_blocked_groups";
-- reverse: create "org_product_org_prices" table
DROP TABLE "org_product_org_prices";
-- reverse: create "org_module_org_prices" table
DROP TABLE "org_module_org_prices";
-- reverse: create "entity_assets" table
DROP TABLE "entity_assets";
-- reverse: create "control_assets" table
DROP TABLE "control_assets";
-- reverse: create index "scanhistory_history_time" to table: "scan_history"
DROP INDEX "scanhistory_history_time";
-- reverse: create "scan_history" table
DROP TABLE "scan_history";
-- reverse: create index "scan_owner_id" to table: "scans"
DROP INDEX "scan_owner_id";
-- reverse: create index "scan_id" to table: "scans"
DROP INDEX "scan_id";
-- reverse: create "scans" table
DROP TABLE "scans";
-- reverse: create index "assethistory_history_time" to table: "asset_history"
DROP INDEX "assethistory_history_time";
-- reverse: create "asset_history" table
DROP TABLE "asset_history";
-- reverse: create index "asset_owner_id" to table: "assets"
DROP INDEX "asset_owner_id";
-- reverse: create index "asset_name_owner_id" to table: "assets"
DROP INDEX "asset_name_owner_id";
-- reverse: create index "asset_id" to table: "assets"
DROP INDEX "asset_id";
-- reverse: create "assets" table
DROP TABLE "assets";
-- reverse: modify "procedures" table
ALTER TABLE "procedures" DROP COLUMN "dismissed_improvement_suggestions", DROP COLUMN "improvement_suggestions", DROP COLUMN "dismissed_control_suggestions", DROP COLUMN "control_suggestions", DROP COLUMN "dismissed_tag_suggestions", DROP COLUMN "tag_suggestions";
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "dismissed_improvement_suggestions", DROP COLUMN "improvement_suggestions", DROP COLUMN "dismissed_control_suggestions", DROP COLUMN "control_suggestions", DROP COLUMN "dismissed_tag_suggestions", DROP COLUMN "tag_suggestions";
-- reverse: create index "organization_settings_compliance_webhook_token_key" to table: "organization_settings"
DROP INDEX "organization_settings_compliance_webhook_token_key";
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "compliance_webhook_token", DROP COLUMN "identity_provider_login_enforced", DROP COLUMN "oidc_discovery_endpoint", DROP COLUMN "identity_provider_entity_id", DROP COLUMN "identity_provider_metadata_endpoint", DROP COLUMN "identity_provider_client_secret", DROP COLUMN "identity_provider_client_id", DROP COLUMN "identity_provider";
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "compliance_webhook_token", DROP COLUMN "identity_provider_login_enforced", DROP COLUMN "oidc_discovery_endpoint", DROP COLUMN "identity_provider_entity_id", DROP COLUMN "identity_provider_metadata_endpoint", DROP COLUMN "identity_provider_client_secret", DROP COLUMN "identity_provider_client_id", DROP COLUMN "identity_provider";
-- reverse: modify "org_products" table
ALTER TABLE "org_products" DROP COLUMN "org_module_org_products", DROP COLUMN "price_id", ADD COLUMN "expires_at" timestamptz NULL, ADD COLUMN "trial_expires_at" timestamptz NULL;
-- reverse: modify "org_prices" table
ALTER TABLE "org_prices" DROP COLUMN "subscription_id", ADD COLUMN "expires_at" timestamptz NULL, ADD COLUMN "trial_expires_at" timestamptz NULL, ADD CONSTRAINT "org_prices_org_products_prices" FOREIGN KEY ("product_id") REFERENCES "org_products" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- reverse: modify "org_modules" table
ALTER TABLE "org_modules" DROP COLUMN "org_product_org_modules", DROP COLUMN "price_id", DROP COLUMN "module_lookup_key", DROP COLUMN "visibility", ADD COLUMN "expires_at" timestamptz NULL, ADD COLUMN "trial_expires_at" timestamptz NULL;
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "dismissed_improvement_suggestions", DROP COLUMN "improvement_suggestions", DROP COLUMN "dismissed_control_suggestions", DROP COLUMN "control_suggestions", DROP COLUMN "dismissed_tag_suggestions", DROP COLUMN "tag_suggestions";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "dismissed_improvement_suggestions", DROP COLUMN "improvement_suggestions", DROP COLUMN "dismissed_control_suggestions", DROP COLUMN "control_suggestions", DROP COLUMN "dismissed_tag_suggestions", DROP COLUMN "tag_suggestions";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP COLUMN "entity_viewers", DROP COLUMN "entity_editors", DROP COLUMN "entity_blocked_groups", DROP COLUMN "asset_viewers", DROP COLUMN "asset_editors", DROP COLUMN "asset_blocked_groups";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP COLUMN "scan_entities", DROP COLUMN "risk_entities";
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP COLUMN "dismissed_improvement_suggestions", DROP COLUMN "improvement_suggestions", DROP COLUMN "dismissed_control_suggestions", DROP COLUMN "control_suggestions", DROP COLUMN "dismissed_tag_suggestions", DROP COLUMN "tag_suggestions";
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "dismissed_improvement_suggestions", DROP COLUMN "improvement_suggestions", DROP COLUMN "dismissed_control_suggestions", DROP COLUMN "control_suggestions", DROP COLUMN "dismissed_tag_suggestions", DROP COLUMN "tag_suggestions";
