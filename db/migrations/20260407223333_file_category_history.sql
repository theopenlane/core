-- Modify "file_history" table
ALTER TABLE "file_history" DROP COLUMN "file_category_name", DROP COLUMN "file_category_id", ADD COLUMN "category_name" character varying NULL, ADD COLUMN "category_id" character varying NULL;
