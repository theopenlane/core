-- +goose Up
-- modify "template_history" table
ALTER TABLE "template_history" ADD COLUMN "system_owned" boolean NULL DEFAULT false;
-- modify "templates" table
ALTER TABLE "templates" ADD COLUMN "system_owned" boolean NULL DEFAULT false;

-- +goose Down
-- reverse: modify "templates" table
ALTER TABLE "templates" DROP COLUMN "system_owned";
-- reverse: modify "template_history" table
ALTER TABLE "template_history" DROP COLUMN "system_owned";
