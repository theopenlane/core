-- +goose Up
-- create index "file_category_name_created_at" to table: "files"
CREATE INDEX "file_category_name_created_at" ON "files" ("category_name", "created_at") WHERE (deleted_at IS NULL);
-- create index "file_created_at" to table: "files"
CREATE INDEX "file_created_at" ON "files" ("created_at") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "file_created_at" to table: "files"
DROP INDEX "file_created_at";
-- reverse: create index "file_category_name_created_at" to table: "files"
DROP INDEX "file_category_name_created_at";
