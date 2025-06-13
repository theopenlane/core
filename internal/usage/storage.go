package usage

import (
	"context"

	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/file"
)

// OrganizationStorageUsage returns the total persisted file size in bytes for an organization
func OrganizationStorageUsage(ctx context.Context, client *generated.Client, orgID string) (int64, error) {
	var out struct {
		Sum int64 `json:"sum"`
	}

	err := client.File.Query().Where(func(s *sql.Selector) {
		s.Where(sql.EQ("owner_id", orgID))
	}).Aggregate(generated.As(generated.Sum(file.FieldPersistedFileSize), "sum")).Scan(ctx, &out)
	if err != nil {
		return 0, err
	}

	return out.Sum, nil
}
