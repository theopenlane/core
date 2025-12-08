-- +goose Up
-- modify "contact_history" table
ALTER TABLE "contact_history" ALTER COLUMN "full_name" DROP NOT NULL;
-- modify "contacts" table
ALTER TABLE "contacts" ALTER COLUMN "full_name" DROP NOT NULL;

-- +goose Down
-- reverse: modify "contacts" table
ALTER TABLE "contacts" ALTER COLUMN "full_name" SET NOT NULL;
-- reverse: modify "contact_history" table
ALTER TABLE "contact_history" ALTER COLUMN "full_name" SET NOT NULL;
