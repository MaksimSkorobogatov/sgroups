package lister

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/H-BF/corlib/pkg/filter"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/suite"
)

type (
	pgRow struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	domainRow struct {
		ID   int
		Name string
	}
)

func pg2domain(r pgRow) (domainRow, error) {
	return domainRow{ID: r.ID, Name: r.Name}, nil
}

type notifyConn struct {
	pgxmock.PgxConnIface

	channel string
	q       []pgconn.Notification
	cancel  context.CancelFunc
	err     error
}

func (c *notifyConn) WaitForNotification(_ context.Context) (*pgconn.Notification, error) {
	if c.err != nil {
		return nil, c.err
	}
	if len(c.q) == 0 {
		if c.cancel != nil {
			c.cancel()
		}
		return nil, nil
	}
	n := c.q[0]
	c.q = c.q[1:]
	if len(c.q) == 0 && c.cancel != nil {
		c.cancel()
	}
	return &n, nil
}

type listerWatcherTestSuite struct {
	suite.Suite
}

func Test_ListerWatcher(t *testing.T) {
	suite.Run(t, new(listerWatcherTestSuite))
}

func (sui *listerWatcherTestSuite) Test_Init() {
	var lw ListerWatcher[domainRow, pgRow]

	err := lw.Init(nil, "select 1", 10, "x")
	sui.Require().NoError(err)
	sui.Require().Equal("select 1", lw.sql)
	sui.Require().Equal([]any{10, "x"}, lw.qryParams)
}

func (sui *listerWatcherTestSuite) Test_List_FiltersRows() {
	ctx := context.Background()

	mock, err := pgxmock.NewConn()
	sui.Require().NoError(err)
	defer func() { _ = mock.Close(ctx) }()

	rows := mock.NewRows([]string{"id", "name"}).
		AddRow(1, "a").
		AddRow(2, "b").
		AddRow(3, "c")

	var lw ListerWatcher[domainRow, pgRow]
	sc := filter.ScopeFromFunc(func(d domainRow) bool { return d.ID%2 == 0 })
	sui.Require().NoError(lw.Init(sc, "sql", 1, 2))

	querier := func(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
		sui.Require().Equal("sql", sql)
		sui.Require().Equal([]any{1, 2}, args)
		return rows.Kind(), nil
	}

	got, err := lw.List(ctx, querier, pg2domain)
	sui.Require().NoError(err)
	sui.Require().Equal([]domainRow{{ID: 2, Name: "b"}}, got)
}

func (sui *listerWatcherTestSuite) Test_Get_ReturnsErrFilteredOut() {
	ctx := context.Background()

	mock, err := pgxmock.NewConn()
	sui.Require().NoError(err)
	defer func() { _ = mock.Close(ctx) }()

	rows := mock.NewRows([]string{"id", "name"}).AddRow(1, "a")

	var lw ListerWatcher[domainRow, pgRow]
	sc := filter.ScopeFromFunc(func(d domainRow) bool { return d.ID == 2 })
	sui.Require().NoError(lw.Init(sc, "sql"))

	querier := func(_ context.Context, sql string, args ...any) (pgx.Rows, error) {
		sui.Require().Equal("sql", sql)
		sui.Require().Len(args, 0)
		return rows.Kind(), nil
	}

	_, err = lw.Get(ctx, querier, pg2domain)
	sui.Require().ErrorIs(err, ErrFilteredOut)
}

func (sui *listerWatcherTestSuite) Test_Get_ReturnsValue() {
	ctx := context.Background()

	mock, err := pgxmock.NewConn()
	sui.Require().NoError(err)
	defer func() { _ = mock.Close(ctx) }()

	rows := mock.NewRows([]string{"id", "name"}).AddRow(1, "a")

	var lw ListerWatcher[domainRow, pgRow]
	sui.Require().NoError(lw.Init(filter.NoScope{}, "sql"))

	querier := func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
		return rows.Kind(), nil
	}

	got, err := lw.Get(ctx, querier, pg2domain)
	sui.Require().NoError(err)
	sui.Require().Equal(domainRow{ID: 1, Name: "a"}, got)
}

func (sui *listerWatcherTestSuite) Test_Get_PropagatesNoRows() {
	ctx := context.Background()

	mock, err := pgxmock.NewConn()
	sui.Require().NoError(err)
	defer func() { _ = mock.Close(ctx) }()

	rows := mock.NewRows([]string{"id", "name"})

	var lw ListerWatcher[domainRow, pgRow]
	sui.Require().NoError(lw.Init(filter.NoScope{}, "sql"))

	querier := func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
		return rows.Kind(), nil
	}

	_, err = lw.Get(ctx, querier, pg2domain)
	sui.Require().ErrorIs(err, pgx.ErrNoRows)
}

func (sui *listerWatcherTestSuite) Test_Watch_ListenNotifyUnlistenAndFilter() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mock, err := pgxmock.NewConn()
	sui.Require().NoError(err)
	defer func() { _ = mock.Close(ctx) }()

	mock.ExpectExec(`^listen ("chan"|chan)$`).WillReturnResult(pgxmock.NewResult("LISTEN", 1))
	mock.ExpectExec(`^unlisten ("chan"|chan)$`).WillReturnResult(pgxmock.NewResult("UNLISTEN", 1))

	okPayload, err := json.Marshal(pgRow{ID: 2, Name: "b"})
	sui.Require().NoError(err)
	filteredPayload, err := json.Marshal(pgRow{ID: 1, Name: "a"})
	sui.Require().NoError(err)

	conn := &notifyConn{
		PgxConnIface: mock,
		channel:      "chan",
		cancel:       cancel,
		q: []pgconn.Notification{
			{Channel: "other", Payload: string(okPayload)},      // ignored by channel
			{Channel: "chan", Payload: ""},                      // ignored by empty payload
			{Channel: "chan", Payload: string(filteredPayload)}, // filtered out
			{Channel: "chan", Payload: string(okPayload)},       // passes
		},
	}

	var lw ListerWatcher[domainRow, pgRow]
	sc := filter.ScopeFromFunc(func(d domainRow) bool { return d.ID%2 == 0 })
	sui.Require().NoError(lw.Init(sc, "unused"))

	var got []domainRow
	err = lw.Watch(ctx, conn, "chan", pg2domain, func(e domainRow) error {
		got = append(got, e)
		return nil
	})
	sui.Require().NoError(err)
	sui.Require().Equal([]domainRow{{ID: 2, Name: "b"}}, got)
	sui.Require().NoError(mock.ExpectationsWereMet())
}

func (sui *listerWatcherTestSuite) Test_Watch_FailsOnBadJSON() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mock, err := pgxmock.NewConn()
	sui.Require().NoError(err)
	defer func() { _ = mock.Close(ctx) }()

	mock.ExpectExec(`^listen ("chan"|chan)$`).WillReturnResult(pgxmock.NewResult("LISTEN", 1))
	mock.ExpectExec(`^unlisten ("chan"|chan)$`).WillReturnResult(pgxmock.NewResult("UNLISTEN", 1))

	conn := &notifyConn{
		PgxConnIface: mock,
		channel:      "chan",
		cancel:       cancel,
		q:            []pgconn.Notification{{Channel: "chan", Payload: "not-json"}},
	}

	var lw ListerWatcher[domainRow, pgRow]
	sui.Require().NoError(lw.Init(filter.NoScope{}, "unused"))

	err = lw.Watch(ctx, conn, "chan", pg2domain, func(domainRow) error {
		return nil
	})
	sui.Require().Error(err)
	sui.Require().Contains(err.Error(), "invalid character")
	sui.Require().NoError(mock.ExpectationsWereMet())
}

func (sui *listerWatcherTestSuite) Test_Watch_FailsOnConvError() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mock, err := pgxmock.NewConn()
	sui.Require().NoError(err)
	defer func() { _ = mock.Close(ctx) }()

	mock.ExpectExec(`^listen ("chan"|chan)$`).WillReturnResult(pgxmock.NewResult("LISTEN", 1))
	mock.ExpectExec(`^unlisten ("chan"|chan)$`).WillReturnResult(pgxmock.NewResult("UNLISTEN", 1))

	payload, err := json.Marshal(pgRow{ID: 1, Name: "a"})
	sui.Require().NoError(err)

	conn := &notifyConn{
		PgxConnIface: mock,
		channel:      "chan",
		cancel:       cancel,
		q:            []pgconn.Notification{{Channel: "chan", Payload: string(payload)}},
	}

	var lw ListerWatcher[domainRow, pgRow]
	sui.Require().NoError(lw.Init(filter.NoScope{}, "unused"))

	convErr := errors.New("conv failed")
	conv := func(pgRow) (domainRow, error) { return domainRow{}, convErr }

	err = lw.Watch(ctx, conn, "chan", conv, func(domainRow) error { return nil })
	sui.Require().ErrorIs(err, convErr)
	sui.Require().NoError(mock.ExpectationsWereMet())
}

func (sui *listerWatcherTestSuite) Test_Watch_PropagatesWaitError() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mock, err := pgxmock.NewConn()
	sui.Require().NoError(err)
	defer func() { _ = mock.Close(ctx) }()

	mock.ExpectExec(`^listen ("chan"|chan)$`).WillReturnResult(pgxmock.NewResult("LISTEN", 1))
	mock.ExpectExec(`^unlisten ("chan"|chan)$`).WillReturnResult(pgxmock.NewResult("UNLISTEN", 1))

	werr := errors.New("wait failed")
	conn := &notifyConn{PgxConnIface: mock, channel: "chan", cancel: cancel, err: werr}

	var lw ListerWatcher[domainRow, pgRow]
	sui.Require().NoError(lw.Init(filter.NoScope{}, "unused"))

	err = lw.Watch(ctx, conn, "chan", pg2domain, func(domainRow) error { return nil })
	sui.Require().ErrorIs(err, werr)
	sui.Require().NoError(mock.ExpectationsWereMet())
}
