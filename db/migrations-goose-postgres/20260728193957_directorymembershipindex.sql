-- +goose Up
-- drop index "directorymembership_directory_account_id_directory_group_id" from table: "directory_memberships"
DROP INDEX "directorymembership_directory_account_id_directory_group_id";
-- create index "directorymembership_directory_account_id_directory_group_id" to table: "directory_memberships"
CREATE UNIQUE INDEX "directorymembership_directory_account_id_directory_group_id" ON "directory_memberships" ("directory_account_id", "directory_group_id") WHERE (removed_at IS NULL);

-- +goose Down
-- reverse: create index "directorymembership_directory_account_id_directory_group_id" to table: "directory_memberships"
DROP INDEX "directorymembership_directory_account_id_directory_group_id";
-- reverse: drop index "directorymembership_directory_account_id_directory_group_id" from table: "directory_memberships"
CREATE UNIQUE INDEX "directorymembership_directory_account_id_directory_group_id" ON "directory_memberships" ("directory_account_id", "directory_group_id");
