package job

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	"github.com/PRO-Robotech/sgroups/internal/shared/patterns"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/nl"
	"github.com/H-BF/corlib/pkg/patterns/observer"
	cfg "github.com/H-BF/corlib/pkg/plain-config"
	"github.com/pkg/errors"
)

type netwatcherImpl struct {
	subject observer.Subject
	NetNS   string

	runOnce   sync.Once
	closeOnce sync.Once
	stopped   chan struct{}
	stop      chan struct{}
}

// NewNetWatcher -
func NewNetWatcher(netNS string) *netwatcherImpl {
	return &netwatcherImpl{
		subject: patterns.NewSubject(),
		NetNS:   netNS,
		stop:    make(chan struct{}),
	}
}

var _ Task = (*netwatcherImpl)(nil)

// Subject -
func (nw *netwatcherImpl) Subject() observer.Subject {
	return patterns.NoNotifySubject{Subject: nw.subject}
}

// MakeObserver -
func (nw *netwatcherImpl) MakeObserver(_ context.Context) patterns.Observer {
	return patterns.NullObserver{}
}

// Close -
func (w *netwatcherImpl) Close() error {
	w.closeOnce.Do(func() {
		w.runOnce.Do(func() {})
		close(w.stop)
		_ = w.subject.Close()
		if w.stopped != nil {
			<-w.stopped
		}
	})
	return nil
}

// Run -
func (w *netwatcherImpl) Run(ctx context.Context) error {
	const job = "net-watcher"

	var neverRun bool
	w.runOnce.Do(func() {
		neverRun = true
		w.stopped = make(chan struct{})
	})
	if !neverRun {
		return fmt.Errorf("%s: it has been run or closed yet", job)
	}
	log := logger.FromContext(ctx).Named(job)
	log.Info("start")
	defer func() {
		log.Info("stop")
		close(w.stopped)
	}()

	return w.handleNetwatcher(ctx)
}

func (nw *netwatcherImpl) handleNetwatcher(ctx context.Context) error {
	watcher, err := nw.makeNetconfWatcher(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = watcher.Close() }()

	if err = nw.gatherLinkState(ctx); err != nil {
		return err
	}
	for stm := watcher.Stream(); ; {
		select {
		case <-nw.stop:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case msgs, ok := <-stm:
			if !ok {
				return nil
			}
			var ev NetlinkUpdatesEvent
			for _, m := range msgs {
				switch t := m.(type) {
				case nl.AddrUpdateMsg:
					ev.Updates = append(ev.Updates, t)
				case nl.LinkUpdateMsg:
					ev.Updates = append(ev.Updates, t)
				case nl.ErrMsg:
					nw.subject.Notify(NetlinkErrorEvent{ErrMsg: t})
					return t
				}
			}
			if len(ev.Updates) > 0 {
				nw.subject.Notify(ev)
			}
		}
	}
}

func (w *netwatcherImpl) gatherLinkState(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			w.subject.Notify(NetlinkErrorEvent{ErrMsg: nl.ErrMsg{Err: err}})
		}
	}()

	var lister nl.LinkLister
	if lister, err = nl.NewLinkLister(ctx, nl.WithNetnsName(w.NetNS)); err != nil {
		return err
	}
	defer lister.Close() //nolint
	var links []nl.Link
	if links, err = lister.List(ctx); err != nil {
		return err
	}
	var ret NetlinkUpdatesEvent
	for _, lnk := range links {
		var addrs []nl.Addr
		if addrs, err = lister.Addrs(ctx, lnk); err != nil {
			return err
		}
		for _, a := range addrs {
			ret.Updates = append(ret.Updates,
				nl.AddrUpdateMsg{
					Address:   *a.IPNet,
					LinkIndex: lnk.Attrs().Index,
				})
		}
	}
	w.subject.Notify(ret)
	return nil
}

func (handler *netwatcherImpl) makeNetconfWatcher(ctx context.Context) (ret nl.NetlinkWatcher, err error) {
	defer func() {
		err = errors.WithMessage(err, "create net-watcher")
	}()
	var ling time.Duration
	if ling, err = config.NetlinkWatcherLinger.Value(ctx); err != nil && !errors.As(err, cfg.ErrNotFound) {
		return nil, err
	}
	opts := []nl.WatcherOption{nl.IgnoreLinks}
	if ling > 0 {
		opts = append(opts, nl.WithLinger{Linger: ling})
	}
	if len(handler.NetNS) > 0 {
		opts = append(opts, nl.WithNetnsName(handler.NetNS))
	}
	return nl.NewNetlinkWatcher(opts...)
}
