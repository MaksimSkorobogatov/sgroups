package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/puddle/v2"
)

const (
	notifyCommitChannel = "commit"
)

func (impl *pgDbRepository) listenCommits(ctx context.Context) {
	const timeoutBeforeRetry = 10 * time.Second

	for pool, ok := impl.pool.Load(); ok; pool, ok = impl.pool.Load() {
		err := pool.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
			conn := c.Conn()
			if _, err := conn.Exec(ctx, "listen "+notifyCommitChannel); err != nil {
				return err
			}
			defer func() {
				_, _ = conn.Exec(context.WithoutCancel(ctx), "unlisten "+notifyCommitChannel)
			}()
			for {
				nt, e := conn.WaitForNotification(ctx)
				if e != nil {
					return e
				}
				if nt.Channel == notifyCommitChannel {
					impl.subject.Notify(DBUpdated{})
				}
			}
		})
		if errors.Is(err, puddle.ErrClosedPool) {
			return
		}
		t := time.NewTimer(timeoutBeforeRetry)
		select {
		case <-ctx.Done():
			_ = t.Stop()
			return
		case <-t.C:
			impl.subject.Notify(DBUpdated{})
		}
	}
}
