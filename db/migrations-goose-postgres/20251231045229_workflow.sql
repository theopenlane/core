-- +goose Up
-- drop index "workflowproposal_workflow_object_ref_id_domain_key" from table: "workflow_proposals"
DROP INDEX "workflowproposal_workflow_object_ref_id_domain_key";
-- create index "workflowproposal_workflow_object_ref_id_domain_key" to table: "workflow_proposals"
CREATE UNIQUE INDEX "workflowproposal_workflow_object_ref_id_domain_key" ON "workflow_proposals" ("workflow_object_ref_id", "domain_key") WHERE ((state)::text = ANY ((ARRAY['DRAFT'::character varying, 'SUBMITTED'::character varying])::text[]));

-- +goose Down
-- reverse: create index "workflowproposal_workflow_object_ref_id_domain_key" to table: "workflow_proposals"
DROP INDEX "workflowproposal_workflow_object_ref_id_domain_key";
-- reverse: drop index "workflowproposal_workflow_object_ref_id_domain_key" from table: "workflow_proposals"
CREATE UNIQUE INDEX "workflowproposal_workflow_object_ref_id_domain_key" ON "workflow_proposals" ("workflow_object_ref_id", "domain_key") WHERE ((state)::text = ANY (ARRAY[('DRAFT'::character varying)::text, ('SUBMITTED'::character varying)::text]));
