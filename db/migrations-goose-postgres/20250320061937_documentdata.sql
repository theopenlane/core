-- +goose Up
-- modify "document_data" table
ALTER TABLE "document_data" DROP CONSTRAINT "document_data_organizations_document_data", ADD CONSTRAINT "document_data_organizations_documents" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "document_data" table
ALTER TABLE "document_data" DROP CONSTRAINT "document_data_organizations_documents", ADD CONSTRAINT "document_data_organizations_document_data" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
