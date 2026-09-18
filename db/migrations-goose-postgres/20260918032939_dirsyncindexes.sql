-- +goose Up
-- create index "directorymembership_owner_id_m_dd26e892afd63c1f93f69002db380441" to table: "directory_memberships"
CREATE INDEX "directorymembership_owner_id_m_dd26e892afd63c1f93f69002db380441" ON "directory_memberships" ("owner_id", "managed_by", "source_definition_id", "source_instance_id") WHERE (removed_at IS NULL);
-- create index "integrationrun_integration_id_operation_name_finished_at" to table: "integration_runs"
CREATE INDEX "integrationrun_integration_id_operation_name_finished_at" ON "integration_runs" ("integration_id", "operation_name", "finished_at") WHERE ((deleted_at IS NULL) AND ((status)::text = 'SUCCESS'::text) AND (finished_at IS NOT NULL));

-- +goose Down
-- reverse: create index "integrationrun_integration_id_operation_name_finished_at" to table: "integration_runs"
DROP INDEX "integrationrun_integration_id_operation_name_finished_at";
-- reverse: create index "directorymembership_owner_id_m_dd26e892afd63c1f93f69002db380441" to table: "directory_memberships"
DROP INDEX "directorymembership_owner_id_m_dd26e892afd63c1f93f69002db380441";
