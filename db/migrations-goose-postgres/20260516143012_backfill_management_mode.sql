-- +goose Up
-- Backfill management_mode for rows that pre-existed the column addition.
-- The ADD COLUMN migration set NULL DEFAULT 'OPENLANE_MANAGED', which only
-- applies to new inserts — existing rows were left NULL until this runs.
UPDATE "action_plans" SET "management_mode" = 'OPENLANE_MANAGED' WHERE "management_mode" IS NULL;
UPDATE "internal_policies" SET "management_mode" = 'OPENLANE_MANAGED' WHERE "management_mode" IS NULL;
UPDATE "procedures" SET "management_mode" = 'OPENLANE_MANAGED' WHERE "management_mode" IS NULL;
UPDATE "action_plan_history" SET "management_mode" = 'OPENLANE_MANAGED' WHERE "management_mode" IS NULL;
UPDATE "internal_policy_history" SET "management_mode" = 'OPENLANE_MANAGED' WHERE "management_mode" IS NULL;
UPDATE "procedure_history" SET "management_mode" = 'OPENLANE_MANAGED' WHERE "management_mode" IS NULL;

-- +goose Down
-- Reversing a backfill is intentionally a no-op; reverting would set every
-- row back to NULL regardless of whether it had been explicitly set, which
-- would lose user data on subsequent re-up.
SELECT 1;
