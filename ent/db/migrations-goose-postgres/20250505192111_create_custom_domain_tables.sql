-- +goose Up
-- create "custom_domain_history" table
CREATE TABLE "custom_domain_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "cname_record" character varying NOT NULL, "mappable_domain_id" character varying NOT NULL, "txt_record_subdomain" character varying NOT NULL DEFAULT '_olverify', "txt_record_value" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'PENDING', PRIMARY KEY ("id"));
-- create index "customdomainhistory_history_time" to table: "custom_domain_history"
CREATE INDEX "customdomainhistory_history_time" ON "custom_domain_history" ("history_time");
-- create "mappable_domains" table
CREATE TABLE "mappable_domains" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "mappabledomain_id" to table: "mappable_domains"
CREATE UNIQUE INDEX "mappabledomain_id" ON "mappable_domains" ("id");
-- create index "mappabledomain_name" to table: "mappable_domains"
CREATE UNIQUE INDEX "mappabledomain_name" ON "mappable_domains" ("name") WHERE (deleted_at IS NULL);
-- create "mappable_domain_history" table
CREATE TABLE "mappable_domain_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "name" character varying NOT NULL, PRIMARY KEY ("id"));
-- create index "mappabledomainhistory_history_time" to table: "mappable_domain_history"
CREATE INDEX "mappabledomainhistory_history_time" ON "mappable_domain_history" ("history_time");
-- create "custom_domains" table
CREATE TABLE "custom_domains" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "cname_record" character varying NOT NULL, "txt_record_subdomain" character varying NOT NULL DEFAULT '_olverify', "txt_record_value" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'PENDING', "mappable_domain_id" character varying NOT NULL, "mappable_domain_custom_domains" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "custom_domains_mappable_domains_custom_domains" FOREIGN KEY ("mappable_domain_custom_domains") REFERENCES "mappable_domains" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "custom_domains_mappable_domains_mappable_domain" FOREIGN KEY ("mappable_domain_id") REFERENCES "mappable_domains" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "custom_domains_organizations_custom_domains" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "customdomain_cname_record" to table: "custom_domains"
CREATE UNIQUE INDEX "customdomain_cname_record" ON "custom_domains" ("cname_record") WHERE (deleted_at IS NULL);
-- create index "customdomain_id" to table: "custom_domains"
CREATE UNIQUE INDEX "customdomain_id" ON "custom_domains" ("id");
-- create index "customdomain_owner_id" to table: "custom_domains"
CREATE INDEX "customdomain_owner_id" ON "custom_domains" ("owner_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "customdomain_owner_id" to table: "custom_domains"
DROP INDEX "customdomain_owner_id";
-- reverse: create index "customdomain_id" to table: "custom_domains"
DROP INDEX "customdomain_id";
-- reverse: create index "customdomain_cname_record" to table: "custom_domains"
DROP INDEX "customdomain_cname_record";
-- reverse: create "custom_domains" table
DROP TABLE "custom_domains";
-- reverse: create index "mappabledomainhistory_history_time" to table: "mappable_domain_history"
DROP INDEX "mappabledomainhistory_history_time";
-- reverse: create "mappable_domain_history" table
DROP TABLE "mappable_domain_history";
-- reverse: create index "mappabledomain_name" to table: "mappable_domains"
DROP INDEX "mappabledomain_name";
-- reverse: create index "mappabledomain_id" to table: "mappable_domains"
DROP INDEX "mappabledomain_id";
-- reverse: create "mappable_domains" table
DROP TABLE "mappable_domains";
-- reverse: create index "customdomainhistory_history_time" to table: "custom_domain_history"
DROP INDEX "customdomainhistory_history_time";
-- reverse: create "custom_domain_history" table
DROP TABLE "custom_domain_history";
