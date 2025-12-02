-- +goose Up
-- modify "template_history" table
ALTER TABLE "template_history" ADD COLUMN "kind" character varying NULL DEFAULT 'QUESTIONNAIRE';
-- modify "templates" table
ALTER TABLE "templates" ADD COLUMN "kind" character varying NULL DEFAULT 'QUESTIONNAIRE';

-- +goose Down
-- reverse: modify "templates" table
ALTER TABLE "templates" DROP COLUMN "kind";
-- reverse: modify "template_history" table
ALTER TABLE "template_history" DROP COLUMN "kind";
