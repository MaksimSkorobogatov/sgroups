package syncer

import (
	"bytes"
	"context"
	"fmt"
	"slices"

	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/jackc/pgx/v5"
)

// SyncOp -
type SyncOp uint8

const (
	// NoSyncOp -
	NoSyncOp SyncOp = 0
	// Insert -
	Insert SyncOp = 1
	// Update -
	Update SyncOp = 2
	// Upsert -
	Upsert SyncOp = Insert | Update
	// Delete -
	Delete SyncOp = 4
)

// String impl Stringer iface
func (op SyncOp) String() string {
	ret, ok := syncOp2S[op]
	if !ok {
		return fmt.Sprintf("sync-op(%v)", int(op))
	}
	return ret
}

var syncOp2Sql = map[SyncOp]string{
	NoSyncOp: "",
	Upsert:   "ups",
	Delete:   "del",
}

var syncOp2S = map[SyncOp]string{
	NoSyncOp: "NoOp",
	Upsert:   "Upsert",
	Delete:   "Delete",
}

// Syncer -
type Syncer[T any, pgT any] interface {
	Sync(ctx context.Context, tx pgx.Tx, op SyncOp, data []T) ([]T, error)
}

type syncerObj[T any, pgT any] struct {
	sqlBuilder  sqlBuilder
	syncArgs    func(pgT) []any
	delArgs     func(pgT) []any
	toPgConv    func(T) (pgT, error)
	fromPgConv  func(pgT) (T, error)
	disabledOps []SyncOp
}

// Sync -
func (o syncerObj[T, pgT]) Sync(ctx context.Context, tx pgx.Tx, op SyncOp, data []T) ([]T, error) {
	if slices.Contains(o.disabledOps, op) {
		return nil, fmt.Errorf("operation '%s' is not enabled here", op)
	}

	sqlB := bytes.NewBuffer(nil)
	o.sqlBuilder.writeOp(sqlB, op)
	sql := sqlB.String()
	ret := make([]T, 0, len(data))
	for _, item := range data {
		var (
			pgObj     pgT
			domainObj T
			err       error
		)

		if pgObj, err = o.toPgConv(item); err != nil {
			return nil, err
		}
		args := misc.Tern(
			op == Delete && o.delArgs != nil, o.delArgs, o.syncArgs,
		)(pgObj)

		if op == Delete {
			if _, err = tx.Exec(ctx, sql, args...); err != nil {
				return nil, err
			}
			continue
		}
		pgObj, err = pgxCollectOneRowAndClose[pgT](ctx, tx, sql, args...)
		if err != nil {
			return ret, err
		}
		if domainObj, err = o.fromPgConv(pgObj); err != nil {
			return nil, err
		}
		ret = append(ret, domainObj)
	}
	return ret, nil
}

func pgxCollectOneRowAndClose[pgT any](ctx context.Context, tx pgx.Tx, sql string, args ...any) (ret pgT, err error) {
	var (
		zero pgT
		rows pgx.Rows
	)

	if rows, err = tx.Query(ctx, sql, args...); err != nil {
		return zero, err
	}
	defer rows.Close()

	ret, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[pgT])
	if err != nil {
		if rowsErr := rows.Err(); rowsErr != nil {
			return zero, rowsErr
		}
		return zero, err
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return zero, rowsErr
	}
	return ret, nil
}
