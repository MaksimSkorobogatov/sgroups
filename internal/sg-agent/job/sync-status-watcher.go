package job

import (
	"context"
	"sync"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/sg-agent/app"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/patterns"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/atomic"
	"github.com/H-BF/corlib/pkg/parallel"
	"github.com/H-BF/corlib/pkg/patterns/observer"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/emptypb"
)

type (
	syncStatusWatcher struct {
		subject        observer.Subject
		runOnce        sync.Once
		closeOnce      sync.Once
		stopped        chan struct{}
		stop           chan struct{}
		clientProvider app.SGClientProvider
		checkInterval  time.Duration
	}
	streamer interface {
		Recv() (*sgv1.SyncStatusResp, error)
	}
)

// NewSyncStatusWatcher -
func NewSyncStatusWatcher(clientProvider app.SGClientProvider, checkInterval time.Duration) *syncStatusWatcher {
	return &syncStatusWatcher{
		subject:        observer.NewSubject(),
		clientProvider: clientProvider,
		checkInterval:  checkInterval,
		stop:           make(chan struct{}),
	}
}

var _ Task = (*syncStatusWatcher)(nil)

// Subject -
func (ss *syncStatusWatcher) Subject() observer.Subject {
	return patterns.NoNotifySubject{Subject: ss.subject}
}

// MakeObserver -
func (ss *syncStatusWatcher) MakeObserver(_ context.Context) patterns.Observer {
	return patterns.NullObserver{}
}

// Close -
func (ss *syncStatusWatcher) Close() error {
	ss.closeOnce.Do(func() {
		ss.runOnce.Do(func() {})
		close(ss.stop)
		_ = ss.subject.Close()
		if ss.stopped != nil {
			<-ss.stopped
		}
	})
	return nil
}

// Run -
func (ss *syncStatusWatcher) Run(ctx context.Context) error {
	const job = "db-status-watcher"
	if ss.checkInterval < time.Second {
		panic("'SyncStatus/checkInterval' is less than 1s")
	}
	log := logger.FromContext(ctx).Named(job)
	ctx = logger.ToContext(ctx, log)

	ss.runOnce.Do(func() {
		ss.stopped = make(chan struct{})
	})
	if ss.stopped == nil {
		return errors.Errorf("%s: it has been run or closed yet", job)
	}
	log.Infow("start")
	defer func() {
		log.Info("stop")
		close(ss.stopped)
	}()

	return ss.watchEvents(ctx)
}

func (ss *syncStatusWatcher) watchEvents(ctx context.Context) error {
	var (
		cancel     func()
		client     sgv1.SGroupsStatusAPIClient
		log        = logger.FromContext(ctx)
		syncStatus atomic.Value[domain.SyncStatus]
	)
	ctx, cancel = context.WithCancel(
		context.WithoutCancel(ctx),
	)
	defer cancel()

	log.Debug("connecting to SGroup server events ...")
	clients, err := ss.clientProvider.NewSGClient(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = clients.Close() }()

	if client, err = clients.Status(); err != nil {
		return err
	}
	tc := time.NewTicker(ss.checkInterval)
	defer tc.Stop()

	log.Debug("start reading SGroups server events")
	funcs := [...]func() error{
		func() error {
			for {
				select {
				case <-tc.C:
					syncStatus.Clear(func(t domain.SyncStatus) {
						ss.subject.Notify(SyncStatusValueEvent{SyncStatus: t})
					})
				case <-ctx.Done():
					return nil
				case <-ss.stop:
					return nil
				}
			}
		},
		func() error {
			var (
				stm  streamer
				resp *sgv1.SyncStatusResp
			)
			for stm, err = client.Watch(ctx, new(emptypb.Empty)); err == nil; {
				if resp, err = stm.Recv(); err == nil {
					syncStatus.Store(domain.SyncStatus{
						UpdatedAt: resp.GetUpdatedAt().AsTime(),
					}, nil)
				}
			}
			return err
		},
	}

	errs := make([]error, len(funcs))
	_ = parallel.ExecAbstract(len(funcs), 1, func(i int) error {
		defer cancel()
		errs[i] = funcs[i]()
		return nil
	})
	return multierr.Combine(errs...)
}
