-- Create "file_download_tokens" table
CREATE TABLE "file_download_tokens" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "token" character varying NULL, "ttl" timestamptz NULL, "user_id" character varying NULL, "organization_id" character varying NULL, "file_id" character varying NULL, "secret" bytea NULL, "owner_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "file_download_tokens_users_file_download_tokens" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create index "file_download_tokens_token_key" to table: "file_download_tokens"
CREATE UNIQUE INDEX "file_download_tokens_token_key" ON "file_download_tokens" ("token");
-- Create index "filedownloadtoken_token" to table: "file_download_tokens"
CREATE UNIQUE INDEX "filedownloadtoken_token" ON "file_download_tokens" ("token") WHERE (deleted_at IS NULL);
