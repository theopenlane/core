-- +goose Up
-- modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" ADD COLUMN "dns_verification_id" character varying NULL;
-- modify "mappable_domain_history" table
ALTER TABLE "mappable_domain_history" ADD COLUMN "zone_id" character varying NOT NULL;
-- modify "mappable_domains" table
ALTER TABLE "mappable_domains" ADD COLUMN "zone_id" character varying NOT NULL;
-- create "dns_verification_history" table
CREATE TABLE "dns_verification_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "cloudflare_hostname_id" character varying NOT NULL, "dns_txt_record" character varying NOT NULL, "dns_txt_value" character varying NOT NULL, "dns_verification_status" character varying NOT NULL DEFAULT 'PENDING', "dns_verification_status_reason" character varying NULL, "acme_challenge_path" character varying NULL, "expected_acme_challenge_value" character varying NULL, "acme_challenge_status" character varying NOT NULL DEFAULT 'PENDING', "acme_challenge_status_reason" character varying NULL, PRIMARY KEY ("id"));
-- create index "dnsverificationhistory_history_time" to table: "dns_verification_history"
CREATE INDEX "dnsverificationhistory_history_time" ON "dns_verification_history" ("history_time");
-- create "dns_verifications" table
CREATE TABLE "dns_verifications" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "cloudflare_hostname_id" character varying NOT NULL, "dns_txt_record" character varying NOT NULL, "dns_txt_value" character varying NOT NULL, "dns_verification_status" character varying NOT NULL DEFAULT 'PENDING', "dns_verification_status_reason" character varying NULL, "acme_challenge_path" character varying NULL, "expected_acme_challenge_value" character varying NULL, "acme_challenge_status" character varying NOT NULL DEFAULT 'PENDING', "acme_challenge_status_reason" character varying NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "dns_verifications_organizations_dns_verifications" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "dnsverification_cloudflare_hostname_id" to table: "dns_verifications"
CREATE UNIQUE INDEX "dnsverification_cloudflare_hostname_id" ON "dns_verifications" ("cloudflare_hostname_id") WHERE (deleted_at IS NULL);
-- create index "dnsverification_id" to table: "dns_verifications"
CREATE UNIQUE INDEX "dnsverification_id" ON "dns_verifications" ("id");
-- modify "custom_domains" table
ALTER TABLE "custom_domains" ADD COLUMN "dns_verification_id" character varying NULL, ADD COLUMN "dns_verification_custom_domains" character varying NULL, ADD CONSTRAINT "custom_domains_dns_verifications_custom_domains" FOREIGN KEY ("dns_verification_custom_domains") REFERENCES "dns_verifications" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "custom_domains_dns_verifications_dns_verification" FOREIGN KEY ("dns_verification_id") REFERENCES "dns_verifications" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "custom_domains" table
ALTER TABLE "custom_domains" DROP CONSTRAINT "custom_domains_dns_verifications_dns_verification", DROP CONSTRAINT "custom_domains_dns_verifications_custom_domains", DROP COLUMN "dns_verification_custom_domains", DROP COLUMN "dns_verification_id";
-- reverse: create index "dnsverification_id" to table: "dns_verifications"
DROP INDEX "dnsverification_id";
-- reverse: create index "dnsverification_cloudflare_hostname_id" to table: "dns_verifications"
DROP INDEX "dnsverification_cloudflare_hostname_id";
-- reverse: create "dns_verifications" table
DROP TABLE "dns_verifications";
-- reverse: create index "dnsverificationhistory_history_time" to table: "dns_verification_history"
DROP INDEX "dnsverificationhistory_history_time";
-- reverse: create "dns_verification_history" table
DROP TABLE "dns_verification_history";
-- reverse: modify "mappable_domains" table
ALTER TABLE "mappable_domains" DROP COLUMN "zone_id";
-- reverse: modify "mappable_domain_history" table
ALTER TABLE "mappable_domain_history" DROP COLUMN "zone_id";
-- reverse: modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" DROP COLUMN "dns_verification_id";
