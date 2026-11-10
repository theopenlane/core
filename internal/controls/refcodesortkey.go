package controls

import (
	"context"
	stdsql "database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/subcontrol"
	"github.com/theopenlane/core/v2/pkg/naturalsort"
)

// refCodeSortKeyBatchSize is the number of rows read, and written in one transaction, per page of the backfill
const refCodeSortKeyBatchSize = 500

// refCodeRow is the slice of a control or subcontrol the sort key backfill reads
type refCodeRow struct {
	ID      string `sql:"id"`
	RefCode string `sql:"ref_code"`
	SortKey string `sql:"ref_code_sort_key"`
}

// BackfillRefCodeSortKeys derives ref_code_sort_key for controls and subcontrols written before the column existed;
// it only rewrites stale keys, so running it again is a no-op
func BackfillRefCodeSortKeys(ctx context.Context, dbClient *generated.Client) error {
	ctx = entx.SkipSoftDelete(ctx)

	if err := backfillRefCodeSortKeyTable(ctx, dbClient.Driver(), control.Table,
		func(ctx context.Context, afterID string) (rows []refCodeRow, err error) {
			err = dbClient.Control.Query().
				Where(control.IDGT(afterID)).
				Order(control.ByID()).
				Limit(refCodeSortKeyBatchSize).
				Select(control.FieldID, control.FieldRefCode, control.FieldRefCodeSortKey).
				Scan(ctx, &rows)

			return rows, err
		},
	); err != nil {
		return err
	}

	return backfillRefCodeSortKeyTable(ctx, dbClient.Driver(), subcontrol.Table,
		func(ctx context.Context, afterID string) (rows []refCodeRow, err error) {
			err = dbClient.Subcontrol.Query().
				Where(subcontrol.IDGT(afterID)).
				Order(subcontrol.ByID()).
				Limit(refCodeSortKeyBatchSize).
				Select(subcontrol.FieldID, subcontrol.FieldRefCode, subcontrol.FieldRefCodeSortKey).
				Scan(ctx, &rows)

			return rows, err
		},
	)
}

// backfillRefCodeSortKeyTable walks one table by id and rewrites every stale key, one transaction per page
func backfillRefCodeSortKeyTable(ctx context.Context, drv dialect.Driver, table string,
	page func(ctx context.Context, afterID string) ([]refCodeRow, error),
) error {
	var reviewed, updated, skipped int

	lastID := ""

	for {
		rows, err := page(ctx, lastID)
		if err != nil {
			return fmt.Errorf("%w: reading %s rows: %w", ErrRefCodeSortKeyBackfill, table, err)
		}

		reviewed += len(rows)

		pageUpdated, pageSkipped, err := writeRefCodeSortKeys(ctx, drv, table, rows)
		if err != nil {
			return err
		}

		updated += pageUpdated
		skipped += pageSkipped

		if len(rows) < refCodeSortKeyBatchSize {
			break
		}

		lastID = rows[len(rows)-1].ID
	}

	logx.FromContext(ctx).Info().Str("table", table).Int("reviewed", reviewed).Int("updated", updated).Int("skipped", skipped).
		Msg("backfill: ref code sort keys completed")

	return nil
}

// writeRefCodeSortKeys writes the stale keys of one page with plain SQL rather than ent mutations, so the derived column
// does not bump updated_at/updated_by or write history; each write is guarded on the ref code the key was derived from
func writeRefCodeSortKeys(ctx context.Context, drv dialect.Driver, table string, rows []refCodeRow) (updated, skipped int, err error) {
	tx, err := drv.Tx(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: starting %s transaction: %w", ErrRefCodeSortKeyBackfill, table, err)
	}

	for _, row := range rows {
		key := naturalsort.Key(row.RefCode)
		if key == row.SortKey {
			continue
		}

		// controls and subcontrols share these column names: the sort key comes from ControlMixin
		query, args := sql.Dialect(dialect.Postgres).Update(table).
			Set(control.FieldRefCodeSortKey, key).
			Where(sql.And(sql.EQ(control.FieldID, row.ID), sql.EQ(control.FieldRefCode, row.RefCode))).
			Query()

		var res stdsql.Result
		if err := tx.Exec(ctx, query, args, &res); err != nil {
			_ = tx.Rollback()

			return 0, 0, fmt.Errorf("%w: updating %s %s: %w", ErrRefCodeSortKeyBackfill, table, row.ID, err)
		}

		n, err := res.RowsAffected()
		if err != nil {
			_ = tx.Rollback()

			return 0, 0, fmt.Errorf("%w: updating %s %s: %w", ErrRefCodeSortKeyBackfill, table, row.ID, err)
		}

		if n == 0 {
			skipped++

			continue
		}

		updated++
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("%w: committing %s page: %w", ErrRefCodeSortKeyBackfill, table, err)
	}

	return updated, skipped, nil
}
