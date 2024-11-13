-- Modify "entitlement_history" table
ALTER TABLE "entitlement_history" ALTER COLUMN "expires" SET DEFAULT false;
-- Modify "entitlements" table
ALTER TABLE "entitlements" ALTER COLUMN "expires" SET DEFAULT false;
