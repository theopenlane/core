-- Drop index "workflowproposal_workflow_object_ref_id_domain_key" from table: "workflow_proposals"
DROP INDEX "workflowproposal_workflow_object_ref_id_domain_key";
-- Create index "workflowproposal_workflow_object_ref_id_domain_key" to table: "workflow_proposals"
CREATE UNIQUE INDEX "workflowproposal_workflow_object_ref_id_domain_key" ON "workflow_proposals" ("workflow_object_ref_id", "domain_key") WHERE ((state)::text = ANY ((ARRAY['DRAFT'::character varying, 'SUBMITTED'::character varying])::text[]));
