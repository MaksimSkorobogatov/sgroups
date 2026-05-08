package lister

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/H-BF/corlib/pkg/filter"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrFilteredOut -
var ErrFilteredOut = errors.New("row filtered out")

type (
	// RowsQuerier -
	RowsQuerier = func(context.Context, string, ...any) (pgx.Rows, error)

	pgConn interface {
		Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
		WaitForNotification(context.Context) (*pgconn.Notification, error)
	}
)

// ListerWatcher works with pg connection and helps to list and watch objects from db table
type ListerWatcher[domainT any, pgT any] struct { //nolint:revive
	sql       string
	qryParams []any
	filter    filter.SimpleFilter[domainT]
}

// List gets and filters out table records
func (lw ListerWatcher[domainT, pgT]) List(ctx context.Context, querier RowsQuerier, conv func(pgT) (domainT, error)) ([]domainT, error) {
	rows, err := querier(ctx, lw.sql, lw.qryParams...)
	if err != nil {
		return nil, err
	}
	var ret []domainT
	err = pgxIterateRowsAndClose(rows, func(obj pgT) error {
		domainObj, e := conv(obj)
		if e != nil {
			return e
		}
		if lw.filter(domainObj) {
			ret = append(ret, domainObj)
		}
		return nil
	})
	return ret, err
}

// Get gets one table record
func (lw ListerWatcher[domainT, pgT]) Get(ctx context.Context, querier RowsQuerier, conv func(pgT) (domainT, error)) (ret domainT, err error) {
	rows, err := querier(ctx, lw.sql, lw.qryParams...)
	if err != nil {
		return ret, err
	}
	defer rows.Close()
	var pgDomain pgT

	if pgDomain, err = pgx.CollectExactlyOneRow(rows, rowTo[pgT]()); err != nil {
		return ret, err
	}

	ret, err = conv(pgDomain)
	if err != nil {
		return ret, err
	}

	if lw.filter(ret) {
		return ret, nil
	}

	return ret, ErrFilteredOut
}

// Init constructs ListerWatcher with given scope
func (lw *ListerWatcher[domainT, pgT]) Init(scope filter.Scope, sql string, params ...any) error {
	lw.sql = sql
	lw.qryParams = params
	if scope == nil {
		scope = filter.NoScope{}
	}
	return lw.filter.InitFromScope(scope)
}

// Watch monitors events via LISTEN/NOTIFY payload only.
func (lw ListerWatcher[domainT, pgT]) Watch(
	ctx context.Context,
	conn pgConn,
	channel string,
	conv func(pgT) (domainT, error),
	onEvent func(domainT) error,
) (err error) {
	quotedChannel := pgx.Identifier{channel}.Sanitize()
	if _, err = conn.Exec(ctx, "listen "+quotedChannel); err != nil {
		return err
	}
	defer func() {
		_, _ = conn.Exec(context.WithoutCancel(ctx), "unlisten "+quotedChannel)
	}()

loop:
	for err == nil {
		select {
		case <-ctx.Done():
			err = nil
			break loop
		default:
		}

		n, e := conn.WaitForNotification(ctx)
		if e != nil {
			err = e
			continue
		}
		if n == nil || n.Payload == "" {
			continue
		}
		if n.Channel != channel {
			continue
		}

		var pgObj pgT

		if err = json.Unmarshal([]byte(n.Payload), &pgObj); err != nil {
			continue
		}
		var domainObj domainT
		if domainObj, err = conv(pgObj); err != nil {
			continue
		}

		if lw.filter(domainObj) {
			err = onEvent(domainObj)
		}
	}

	return err
}

func pgxIterateRowsAndClose[pgT any](rows pgx.Rows, consumer func(pgT) error) error {
	scan := pgx.RowToStructByName[pgT]
	defer rows.Close()
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return err
		}
		if err = consumer(v); err != nil {
			return err
		}
	}
	return rows.Err()
}

// rowTo returns a pgx row mapper suitable for pgT.
func rowTo[pgT any]() pgx.RowToFunc[pgT] {
	t := reflect.TypeFor[pgT]()

	if t.Kind() == reflect.Struct {
		return pgx.RowToStructByName[pgT]
	}
	return pgx.RowTo[pgT]
}
