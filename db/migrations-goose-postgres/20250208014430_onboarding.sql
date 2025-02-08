-- +goose Up
-- create "onboardings" table
CREATE TABLE "onboardings" ("id" character varying NOT NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "company_name" character varying NOT NULL, "domains" jsonb NULL, "company_details" jsonb NULL, "user_details" jsonb NULL, "compliance" jsonb NULL, "organization_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "onboardings_organizations_organization" FOREIGN KEY ("organization_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);

-- +goose Down
-- reverse: create "onboardings" table
DROP TABLE "onboardings";
