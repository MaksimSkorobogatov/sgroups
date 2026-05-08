package main

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/sg-agent/app"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/job"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/nft"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/parallel"
	"github.com/H-BF/corlib/pkg/patterns/observer"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type tasks struct {
	observers []observer.Observer
	tasks     []job.Task
}

func (t *tasks) cleanup() {
	for _, o := range t.observers {
		_ = o.Close()
	}
	for _, task := range t.tasks {
		_ = task.Close()
	}
}

func (t *tasks) init(ctx context.Context) (err error) {
	var (
		netNs                   string
		syncStatusCheckInterval time.Duration
		rulesApplier            nft.RuleApplier
		rulesApplierProvider    nft.RuleApplierProvider
	)

	syncStatusCheckInterval, err = config.SGroupsSyncStatusInterval.Value(ctx)
	if err != nil && !errors.Is(err, config.ErrNotFound) {
		return err
	}
	netNs, err = config.NetNS.Value(ctx)
	if err != nil && !errors.Is(err, config.ErrNotFound) {
		return err
	}

	if rulesApplierProvider, err = makeRuleApplierProvider(ctx); err != nil {
		return err
	}

	if rulesApplier, err = rulesApplierProvider.NewApplier(ctx); err != nil {
		return err
	}
	defer func() { _ = rulesApplier.Close() }()

	var (
		taskRulesApplier job.Task
		taskNetWatcher   job.Task
		taskDnsRefresher job.Task
		taskSyncStatus   job.Task
	)
	if taskRulesApplier, err = makeRuleApplier(ctx, rulesApplier); err != nil {
		return err
	}

	taskNetWatcher = job.NewNetWatcher(netNs)
	taskDnsRefresher = job.NewDnsRefresher()
	taskSyncStatus = job.NewSyncStatusWatcher(app.SGClientProviderInstance, syncStatusCheckInterval)

	t.tasks = misc.Sli(
		taskSyncStatus, taskRulesApplier, taskNetWatcher,
		taskDnsRefresher,
	)
	raObs := taskRulesApplier.MakeObserver(ctx)
	dnsRefObs := taskDnsRefresher.MakeObserver(ctx)
	taskRulesApplier.Subject().ObserversAttach(dnsRefObs)
	taskNetWatcher.Subject().ObserversAttach(raObs)
	taskSyncStatus.Subject().ObserversAttach(raObs)
	taskDnsRefresher.Subject().ObserversAttach(raObs)
	t.observers = []observer.Observer{raObs}

	return nil
}

func (t *tasks) run(ctx context.Context) error {
	c := make(chan struct{})
	tasks := misc.Sli(func() error {
		select {
		case <-c:
		case <-ctx.Done():
			return context.Cause(ctx)
		}
		return nil
	})
	var closeTasks []io.Closer
	for _, task := range t.tasks {
		closeTasks = append(closeTasks, task)
		tasks = append(tasks, func() error {
			return task.Run(ctx)
		})
	}
	var stopAllOnce sync.Once
	errs := make([]error, len(tasks))
	_ = parallel.ExecAbstract(len(tasks), int32(len(tasks)-1), func(i int) error { //nolint:gosec
		defer stopAllOnce.Do(func() {
			close(c)
			for _, c := range closeTasks {
				_ = c.Close()
			}
		})
		e := tasks[i]()
		if !misc.ErrorIsInAny(e, context.Canceled, context.DeadlineExceeded) {
			errs[i] = e
		}
		return nil
	})
	return multierr.Combine(errs...)
}

func runTasks(ctx context.Context) (err error) {
	var (
		waitBeforeRestart time.Duration
		continueOnFailure bool
	)
	if waitBeforeRestart, err = config.ContinueAfterTimeout.Value(ctx); err != nil {
		return err
	}
	if continueOnFailure, err = config.ContinueOnFailure.Value(ctx); err != nil {
		return err
	}

	ctx1 := logger.ToContext(ctx,
		logger.FromContext(ctx).Named("main"),
	)

	logger.Infof(ctx1, "start")
	defer logger.Infof(ctx1, "exit")

	run := func() error {
		var t tasks
		defer t.cleanup()
		if e := t.init(ctx1); e != nil {
			return e
		}
		return t.run(ctx1)
	}
	for err = run(); err == nil; {
		if !continueOnFailure {
			logger.Info(ctx1, "will exit cause 'ContinueOnFailure' policy is off")
			break
		}
		logger.Infof(ctx1, "will retry after %s", waitBeforeRestart)
		if waitBeforeRestart >= time.Second {
			select {
			case <-time.After(waitBeforeRestart):
			case <-ctx.Done():
				return nil
			}
		}
	}

	return err
}
