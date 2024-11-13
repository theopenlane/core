-- +goose Up
-- modify "entitlement_history" table
ALTER TABLE "entitlement_history" ALTER COLUMN "expires" SET DEFAULT false;
-- modify "entitlements" table
ALTER TABLE "entitlements" ALTER COLUMN "expires" SET DEFAULT false;

-- +goose Down
-- reverse: modify "entitlements" table
ALTER TABLE "entitlements" ALTER COLUMN "expires" SET DEFAULT true;
-- reverse: modify "entitlement_history" table
ALTER TABLE "entitlement_history" ALTER COLUMN "expires" SET DEFAULT true;
