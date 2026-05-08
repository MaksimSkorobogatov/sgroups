package job

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/sg-agent/app"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/patterns"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/dict"
	"github.com/H-BF/corlib/pkg/jsonview"
	"github.com/H-BF/corlib/pkg/patterns/observer"
	"github.com/H-BF/corlib/pkg/queue"
	"github.com/c-robinson/iplib"
	"github.com/pkg/errors"
)

type (
	ask2ResolveDomainAddresses struct {
		IpVersion   int
		FQDN        domain.FQDN
		ValidBefore time.Time

		observer.EventType
	}

	// dnsRefresherImpl -
	dnsRefresherImpl struct {
		subject       observer.Subject
		sema          chan struct{}
		stopped       chan struct{}
		onceClose     sync.Once
		onceRun       sync.Once
		que           queue.FIFO[DomainAddressesEvent]
		activeQueries struct {
			sync.Mutex
			closed bool
			dict.RBDict[ask2ResolveDomainAddresses, *time.Timer]
		}
	}
)

// NewDnsRefresher -
func NewDnsRefresher() *dnsRefresherImpl {
	const semaphoreCap = 4

	ret := dnsRefresherImpl{
		subject: patterns.NewSubject(),
		sema:    make(chan struct{}, semaphoreCap),
		que:     queue.NewFIFO[DomainAddressesEvent](),
	}
	for i := 0; i < cap(ret.sema); i++ {
		ret.sema <- struct{}{}
	}
	return &ret
}

var _ Task = (*dnsRefresherImpl)(nil)

// Subject -
func (rf *dnsRefresherImpl) Subject() observer.Subject {
	return patterns.NoNotifySubject{Subject: rf.subject}
}

// MakeObserver -
func (rf *dnsRefresherImpl) MakeObserver(ctx context.Context) patterns.Observer {
	return observer.NewObserver(func(ev observer.EventType) {
		switch o := ev.(type) {
		case ask2ResolveDomainAddresses:
			rf.onAsk2ResolveDomainAddresses(ctx, o)
		}
	}, false, ask2ResolveDomainAddresses{})
}

// Close -
func (rf *dnsRefresherImpl) Close() error {
	rf.onceClose.Do(func() {
		_ = rf.que.Close()
		rf.onceRun.Do(func() {})
		rf.activeQueries.Lock()
		rf.activeQueries.closed = true
		rf.activeQueries.Iterate(func(_ ask2ResolveDomainAddresses, v *time.Timer) bool {
			_ = v.Stop()
			return true
		})
		rf.activeQueries.Unlock()
		if rf.stopped != nil {
			<-rf.stopped
		}
	})
	return nil
}

// Run -
func (rf *dnsRefresherImpl) Run(ctx context.Context) (err error) {
	const job = "dns-refresher"

	var doRun bool
	rf.onceRun.Do(func() {
		doRun = true
		rf.stopped = make(chan struct{})
	})
	if !doRun {
		return errors.Errorf("%s: it has been run or closed yet", job)
	}

	log := logger.FromContext(ctx).Named(job)
	log.Info("start")
	defer func() {
		close(rf.stopped)
		log.Info("stop")
	}()
	for events := rf.que.Reader(); ; {
		select {
		case <-ctx.Done():
			log.Info("will exit cause it has canceled")
			return ctx.Err()
		case ev, ok := <-events:
			if !ok {
				log.Infof("will exit cause it has closed")
				return nil
			}
			log1 := log.WithField("domain", ev.FQDN).WithField("IPv", ev.IpVersion)
			if e := ev.DnsAnswer.Err; e != nil {
				log1.Errorw("resolved", "error", jsonview.Stringer(e))
			} else {
				log1.Debugw("resolved",
					"TTL", jsonview.Stringer(ev.DnsAnswer.TTL.Round(time.Second)),
					"IP(s)", ev.DnsAnswer.IPs)
			}
			rf.subject.Notify(ev)
		}
	}
}

func (rf *dnsRefresherImpl) resolve(ctx context.Context, ask ask2ResolveDomainAddresses) DomainAddressesEvent {
	ret := DomainAddressesEvent{
		IpVersion: ask.IpVersion,
		FQDN:      ask.FQDN,
	}
	resolver := app.GetDnsResolver()
	domain := ask.FQDN.String()
	switch ask.IpVersion {
	case iplib.IP4Version:
		ret.DnsAnswer = resolver.A(ctx, domain)
	case iplib.IP6Version:
		ret.DnsAnswer = resolver.AAAA(ctx, domain)
	default:
		panic(
			fmt.Errorf("FqdnRefresher: passed unsupported IP version: %v'", ask.IpVersion),
		)
	}
	return ret
}

func (rf *dnsRefresherImpl) onAsk2ResolveDomainAddresses(ctx context.Context, ev ask2ResolveDomainAddresses) {
	log := logger.FromContext(ctx)
	rf.activeQueries.Lock()
	defer rf.activeQueries.Unlock()
	if rf.activeQueries.At(ev) != nil || rf.activeQueries.closed {
		return
	}

	now := time.Now()
	ttl := ev.ValidBefore.Sub(now)
	if ttl < time.Minute {
		ttl = time.Minute
	}
	log.Debugw("ask-to-resolve",
		"IPv", ev.IpVersion,
		"domain", jsonview.Stringer(ev.FQDN),
		"after", jsonview.Stringer(ttl.Round(time.Second)),
	)
	newTimer := time.AfterFunc(ttl, func() {
		select {
		case <-ctx.Done():
		case <-rf.sema:
			defer func() {
				rf.sema <- struct{}{}
			}()
			ret := rf.resolve(ctx, ev)
			rf.que.Put(ret)
		}
		rf.activeQueries.Lock()
		defer rf.activeQueries.Unlock()
		rf.activeQueries.Del(ev)
	})
	rf.activeQueries.Put(ev, newTimer)
}

// Cmp -
func (a ask2ResolveDomainAddresses) Cmp(other ask2ResolveDomainAddresses) int {
	if a.IpVersion > other.IpVersion {
		return 1
	}
	if a.IpVersion < other.IpVersion {
		return -1
	}
	return a.FQDN.Cmp(other.FQDN)
}
