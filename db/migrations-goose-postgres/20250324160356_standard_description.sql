-- +goose Up
-- modify "standard_history" table
ALTER TABLE "standard_history" ALTER COLUMN "description" TYPE text;
-- modify "standards" table
ALTER TABLE "standards" ALTER COLUMN "description" TYPE text;

-- +goose Down
-- reverse: modify "standards" table
ALTER TABLE "standards" ALTER COLUMN "description" TYPE character varying;
-- reverse: modify "standard_history" table
ALTER TABLE "standard_history" ALTER COLUMN "description" TYPE character varying;
