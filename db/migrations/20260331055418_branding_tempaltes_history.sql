-- Modify "email_template_history" table
ALTER TABLE "email_template_history" DROP COLUMN "email_branding_id", ALTER COLUMN "template_context" SET NOT NULL;
