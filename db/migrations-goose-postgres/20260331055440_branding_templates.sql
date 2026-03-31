-- +goose Up
-- modify "email_templates" table
ALTER TABLE "email_templates" DROP COLUMN "email_branding_id", ALTER COLUMN "template_context" SET NOT NULL;
-- create "email_branding_email_templates" table
CREATE TABLE "email_branding_email_templates" ("email_branding_id" character varying NOT NULL, "email_template_id" character varying NOT NULL, PRIMARY KEY ("email_branding_id", "email_template_id"), CONSTRAINT "email_branding_email_templates_email_branding_id" FOREIGN KEY ("email_branding_id") REFERENCES "email_brandings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "email_branding_email_templates_email_template_id" FOREIGN KEY ("email_template_id") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "email_template_blocked_groups" character varying NULL, ADD COLUMN "email_template_editors" character varying NULL, ADD COLUMN "email_template_viewers" character varying NULL, ADD CONSTRAINT "groups_email_templates_blocked_groups" FOREIGN KEY ("email_template_blocked_groups") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_email_templates_editors" FOREIGN KEY ("email_template_editors") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_email_templates_viewers" FOREIGN KEY ("email_template_viewers") REFERENCES "email_templates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_email_templates_viewers", DROP CONSTRAINT "groups_email_templates_editors", DROP CONSTRAINT "groups_email_templates_blocked_groups", DROP COLUMN "email_template_viewers", DROP COLUMN "email_template_editors", DROP COLUMN "email_template_blocked_groups";
-- reverse: create "email_branding_email_templates" table
DROP TABLE "email_branding_email_templates";
-- reverse: modify "email_templates" table
ALTER TABLE "email_templates" ALTER COLUMN "template_context" DROP NOT NULL, ADD COLUMN "email_branding_id" character varying NULL;
