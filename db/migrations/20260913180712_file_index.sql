-- Create index "file_category_name_created_at" to table: "files"
CREATE INDEX "file_category_name_created_at" ON "files" ("category_name", "created_at") WHERE (deleted_at IS NULL);
-- Create index "file_created_at" to table: "files"
CREATE INDEX "file_created_at" ON "files" ("created_at") WHERE (deleted_at IS NULL);
