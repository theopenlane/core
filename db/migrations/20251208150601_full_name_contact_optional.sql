-- Modify "contact_history" table
ALTER TABLE "contact_history" ALTER COLUMN "full_name" DROP NOT NULL;
-- Modify "contacts" table
ALTER TABLE "contacts" ALTER COLUMN "full_name" DROP NOT NULL;
