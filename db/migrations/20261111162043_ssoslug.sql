-- Create index "organization_slug_name" to table: "organizations"
CREATE UNIQUE INDEX "organization_slug_name" ON "organizations" ("slug_name") WHERE ((deleted_at IS NULL) AND ((slug_name)::text <> ''::text));
