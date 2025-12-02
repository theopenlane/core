-- +goose Up
-- create "org_modules" table
CREATE TABLE "org_modules" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "module" character varying NOT NULL, "price" jsonb NULL, "stripe_price_id" character varying NULL, "status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "trial_expires_at" timestamptz NULL, "expires_at" timestamptz NULL, "subscription_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "org_modules_org_subscriptions_modules" FOREIGN KEY ("subscription_id") REFERENCES "org_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "org_modules_organizations_org_modules" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "orgmodule_id" to table: "org_modules"
CREATE UNIQUE INDEX "orgmodule_id" ON "org_modules" ("id");
-- create index "orgmodule_owner_id" to table: "org_modules"
CREATE INDEX "orgmodule_owner_id" ON "org_modules" ("owner_id") WHERE (deleted_at IS NULL);
-- create "org_products" table
CREATE TABLE "org_products" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "module" character varying NOT NULL, "stripe_product_id" character varying NULL, "status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "trial_expires_at" timestamptz NULL, "expires_at" timestamptz NULL, "subscription_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "org_products_org_subscriptions_products" FOREIGN KEY ("subscription_id") REFERENCES "org_subscriptions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "org_products_organizations_org_products" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "orgproduct_id" to table: "org_products"
CREATE UNIQUE INDEX "orgproduct_id" ON "org_products" ("id");
-- create index "orgproduct_owner_id" to table: "org_products"
CREATE INDEX "orgproduct_owner_id" ON "org_products" ("owner_id") WHERE (deleted_at IS NULL);
-- create "org_prices" table
CREATE TABLE "org_prices" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "price" jsonb NULL, "stripe_price_id" character varying NULL, "status" character varying NULL, "active" boolean NOT NULL DEFAULT true, "trial_expires_at" timestamptz NULL, "expires_at" timestamptz NULL, "product_id" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "org_prices_org_products_prices" FOREIGN KEY ("product_id") REFERENCES "org_products" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "org_prices_organizations_org_prices" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "orgprice_id" to table: "org_prices"
CREATE UNIQUE INDEX "orgprice_id" ON "org_prices" ("id");
-- create index "orgprice_owner_id" to table: "org_prices"
CREATE INDEX "orgprice_owner_id" ON "org_prices" ("owner_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "orgprice_owner_id" to table: "org_prices"
DROP INDEX "orgprice_owner_id";
-- reverse: create index "orgprice_id" to table: "org_prices"
DROP INDEX "orgprice_id";
-- reverse: create "org_prices" table
DROP TABLE "org_prices";
-- reverse: create index "orgproduct_owner_id" to table: "org_products"
DROP INDEX "orgproduct_owner_id";
-- reverse: create index "orgproduct_id" to table: "org_products"
DROP INDEX "orgproduct_id";
-- reverse: create "org_products" table
DROP TABLE "org_products";
-- reverse: create index "orgmodule_owner_id" to table: "org_modules"
DROP INDEX "orgmodule_owner_id";
-- reverse: create index "orgmodule_id" to table: "org_modules"
DROP INDEX "orgmodule_id";
-- reverse: create "org_modules" table
DROP TABLE "org_modules";
