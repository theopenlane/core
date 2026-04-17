-- Modify "files" table
ALTER TABLE "files" ADD COLUMN "category_name" character varying NULL, ADD COLUMN "name" character varying NULL, ADD COLUMN "category_id" character varying NULL, ADD CONSTRAINT "files_custom_type_enums_category" FOREIGN KEY ("category_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
