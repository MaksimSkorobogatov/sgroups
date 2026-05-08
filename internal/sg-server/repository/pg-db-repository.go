package repository

import (
	"context"
	"net/url"
	"sync/atomic"
	"time"

	pg "github.com/PRO-Robotech/sgroups/internal/sg-server/repository/pg/domain"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"
	"github.com/PRO-Robotech/sgroups/internal/shared/patterns"

	"github.com/H-BF/corlib/logger"
	atm "github.com/H-BF/corlib/pkg/atomic"
	"github.com/H-BF/corlib/pkg/backoff"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

// NewRepositoryFromPG creates repository from Postgres
func NewRepositoryFromPG(ctx context.Context, dbURL url.URL) (r Repository, err error) {
	var conf *pgxpool.Config
	defer func() {
		if err != nil {
			err = errors.WithMessage(err, "NewRepositoryFromPG")
		}
	}()
	if conf, err = pgxpool.ParseConfig(dbURL.String()); err != nil {
		return nil, err
	}
	conf.HealthCheckPeriod = 30 * time.Second
	conf.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		return pg.RegisterSGroupsTypesOntoPGX(ctx, c)
	}
	var pool *pgxpool.Pool
	if pool, err = pgxpool.NewWithConfig(ctx, conf); err != nil {
		return nil, err
	}
	ret := &pgDbRepository{
		subject: patterns.NewSubject(),
	}
	ret.pool.Store(pool, nil)
	ctx, ret.cancelPubSub = context.WithCancel(ctx)
	go ret.listenCommits(ctx)
	return ret, nil
}

var _ Repository = (*pgDbRepository)(nil)

type (
	pgDbRepository struct {
		subject      patterns.Subject
		pool         atm.Value[*pgxpool.Pool]
		cancelPubSub func()
		syncCount    int64
	}

	pgConn interface {
		Query(context.Context, string, ...any) (pgx.Rows, error)
		Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
		WaitForNotification(context.Context) (*pgconn.Notification, error)
	}
)

// Subject impl Repository interface
func (impl *pgDbRepository) Subject() patterns.Subject {
	return impl.subject
}

// Reader impl Repository interface
func (impl *pgDbRepository) Reader(_ context.Context) (r Reader, err error) {
	defer func() {
		err = errors.WithMessage(err, "PG/Reader")
	}()
	err = ErrNoRepository

	type fu = func(context.Context) (*pgxpool.Conn, error)
	var connAcquirer atm.Value[fu]
	_ = impl.pool.Fetch(func(pool *pgxpool.Pool) {
		connAcquirer.Store(func(ctx1 context.Context) (*pgxpool.Conn, error) {
			return pool.Acquire(ctx1)
		}, nil)
		ret := new(pgDbReader)
		ret.doIt = func(ctx1 context.Context, f func(pgConn) error) error {
			cc, ok := connAcquirer.Load()
			if !ok {
				return ErrReaderClosed
			}
			c, e := cc(ctx1)
			if e != nil {
				return correctPGError(e)
			}
			defer c.Release()
			return correctPGError(f(c.Conn()))
		}
		ret.close = func() {
			connAcquirer.Clear(nil)
		}
		r = ret
		err = nil
	})
	return r, err
}

// Do impl Repository interface
func (impl *pgDbRepository) Do(ctx context.Context, fn func(WriterFace) error) (err error) {
	const api = "PG/Do"
	defer func() {
		err = errors.WithMessage(err, api)
	}()

	do := func() (e error) {
		var wr Writer
		if wr, e = impl.Writer(ctx); e != nil {
			return e
		}
		e = fn(wr)
		if e != nil {
			wr.Abort()
		} else {
			e = wr.Commit()
		}
		return e
	}
	log := logger.FromContext(ctx).Named(api)
	var attempt int
	for bk := makeWriteRetryBackoff(ctx); ; attempt++ {
		if err = do(); err == nil {
			return nil
		}
		if !isSerializationFailure(err) {
			return correctPGError(err)
		}
		pause := bk.NextBackOff()
		if pause == backoff.Stop {
			return errors.WithMessagef(err, "too many serialization retries: (%d)", attempt+1)
		}
		log.Warnf("PG conflict: %v; attempt %d; will retry after %v", err, attempt+1, pause)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pause):
		}
	}
}

// Writer implements Repository
func (impl *pgDbRepository) Writer(ctx context.Context) (w Writer, err error) {
	defer func() {
		err = errors.WithMessage(err, "PG/Writer")
	}()
	err = ErrNoRepository
	_ = impl.pool.Fetch(func(pool *pgxpool.Pool) {
		var txHolder atm.Value[pgx.Tx]
		var tx pgx.Tx
		txOpts := pgx.TxOptions{
			IsoLevel:   pgx.RepeatableRead,
			AccessMode: pgx.ReadWrite,
		}
		if tx, err = pool.BeginTx(ctx, txOpts); err != nil {
			err = correctPGError(err)
			return
		}
		txHolder.Store(tx, nil)
		err = nil
		w = &pgDbWriter{
			tx: func() (pgx.Tx, error) {
				x, ok := txHolder.Load()
				if !ok {
					return nil, ErrWriterClosed
				}
				return x, nil
			},
			abort: func() {
				txHolder.Clear(func(t pgx.Tx) {
					_ = t.Rollback(ctx)
				})
			},
			commit: func() error {
				e := ErrWriterClosed
				txHolder.Clear(func(t pgx.Tx) {
					n := atomic.AddInt64(&impl.syncCount, 1)
					e = (pg.SyncStatus{SyncCount: n}).Store(ctx, tx.Conn())
					if e != nil {
						_ = t.Rollback(ctx)
						return
					}
					if _, e = t.Exec(ctx, "notify "+notifyCommitChannel); e == nil {
						e = t.Commit(ctx)
					}
					if e != nil {
						_ = t.Rollback(ctx)
					}
				})
				return correctPGError(e)
			},
		}
	})
	return w, err
}

// Close implements Repository
func (impl *pgDbRepository) Close() error {
	impl.cancelPubSub()
	_ = impl.subject.Close()
	impl.pool.Clear(func(p *pgxpool.Pool) {
		p.Close()
	})
	return nil
}

func makeWriteRetryBackoff(ctx context.Context) backoff.Backoff {
	const (
		initInt    = 100 * time.Millisecond
		mul        = 1.5
		rand       = 0.5
		maxInt     = 5 * time.Second
		maxElapsed = 15 * time.Second
	)
	bk := backoff.ExponentialBackoffBuilder().
		WithInitialInterval(initInt).
		WithMultiplier(mul).
		WithRandomizationFactor(rand).
		WithMaxInterval(maxInt).
		WithMaxElapsedThreshold(maxElapsed)
	if dl, ok := ctx.Deadline(); ok {
		if d := time.Until(dl); d > 0 && d < maxElapsed {
			bk = bk.WithMaxElapsedThreshold(d)
		}
	}
	b := bk.Build()
	b.Reset()
	return backoff.WithContext(b, ctx)
}

func isSerializationFailure(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return misc.IsIn(pgErr.Code, "40001", "40P01")
}
