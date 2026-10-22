-- Modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" ADD COLUMN "purpose" character varying NULL DEFAULT 'TRUSTCENTER';
