package job

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/sg-agent/app"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/dns"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/nft"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/resources"
	client "github.com/PRO-Robotech/sgroups/internal/sg-agent/transport/sgroups/grpc"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"
	"github.com/PRO-Robotech/sgroups/internal/shared/patterns"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/host"
	"github.com/H-BF/corlib/pkg/patterns/observer"
	"github.com/H-BF/corlib/pkg/queue"
	"github.com/c-robinson/iplib"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

// NewNftApplier -
func NewNftApplier(applier nft.RuleApplier, clientProvider app.SGClientProvider, opts ...Option) *rulesApplier {
	ret := &rulesApplier{
		subject:        patterns.NewSubject(),
		nftRuleApplier: applier,
		clientProvider: clientProvider,
		que:            queue.NewFIFO[DomainAddressesEvent](),
		stop:           make(chan struct{}),
	}
	ret.trigger2apply.sgn = make(chan struct{}, 1)
	for _, o := range opts {
		switch t := o.(type) {
		case WithNetNS:
			ret.netNS = string(t)
		case WithDefPolicyAccept:
			ret.defPolicyAccept = bool(t)
		}
	}

	return ret
}

type rulesApplier struct {
	subject         observer.Subject
	netNS           string
	defPolicyAccept bool
	clientProvider  app.SGClientProvider
	nftRuleApplier  nft.RuleApplier
	que             queue.FIFO[DomainAddressesEvent]
	stopped         chan struct{}
	stop            chan struct{}
	onceRun         sync.Once
	onceClose       sync.Once
	appliedCount    int
	trigger2apply   struct {
		sync.Mutex
		sgn            chan struct{}
		dbSyncChanged  int
		netConfChanged int
		syncStatus     *domain.SyncStatus
		netConf        *host.NetConf
	}
}

// MakeObserver -
func (jb *rulesApplier) MakeObserver(_ context.Context) patterns.Observer {
	return observer.NewObserver(
		jb.incomingEvents,
		false,
		NetlinkUpdatesEvent{},
		SyncStatusValueEvent{},
		DomainAddressesEvent{},
	)
}

// Subject -
func (jb *rulesApplier) Subject() observer.Subject {
	return patterns.NoNotifySubject{Subject: jb.subject}
}

// Close -
func (jb *rulesApplier) Close() error {
	jb.onceClose.Do(func() {
		_ = jb.que.Close()
		jb.onceRun.Do(func() {})
		close(jb.stop)
		_ = jb.subject.Close()
		if jb.stopped != nil {
			<-jb.stopped
		}
	})
	return nil
}

// Run -
func (jb *rulesApplier) Run(ctx context.Context) (err error) {
	const job = "rules-applier"

	var neverRun bool
	jb.onceRun.Do(func() {
		neverRun = true
		jb.stopped = make(chan struct{})
	})
	if !neverRun {
		return fmt.Errorf("%s: it has been run or closed yet", job)
	}

	log := logger.FromContext(ctx).Named(job)
	log.Info("start")
	defer func() {
		defer log.Info("stop")
		close(jb.stopped)
	}()
	client, err := jb.clientProvider.NewSGClient(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close()
	}()
	for que := jb.que.Reader(); err == nil; {
		select {
		case <-ctx.Done():
			log.Info("will exit cause it has canceled")
			return ctx.Err()
		case <-jb.trigger2apply.sgn:
			tr := &jb.trigger2apply
			tr.Lock()
			if tr.netConfChanged == 1 {
				log.Info("net conf has changed")
				tr.netConfChanged++
			}
			if tr.dbSyncChanged == 1 {
				log.Info("ruleset repo has changed")
				tr.dbSyncChanged++
			}
			if tr.netConf == nil || tr.syncStatus == nil {
				tr.Unlock()
			} else {
				tr.dbSyncChanged, tr.netConfChanged = 0, 0
				nc := *tr.netConf
				tr.Unlock()
				err = jb.doApply(ctx, *client, nc)
			}
		case o, ok := <-que:
			if !ok {
				log.Info("will exit cause it has closed")
				return nil
			}
			err = jb.handleDomainAddressesEvent(ctx, o)
		}
	}
	if err != nil {
		log.Error(err)
	}

	return err
}

// incomingEvents -
func (jb *rulesApplier) incomingEvents(ev observer.EventType) { //async recv
	switch o := ev.(type) {
	case NetlinkUpdatesEvent:
		jb.handleNetlinkEvent(o)
	case SyncStatusValueEvent:
		jb.handleSyncStatus(o)
	case DomainAddressesEvent:
		_ = jb.que.Put(o)
	}
}

func (jb *rulesApplier) handleSyncStatus(ev SyncStatusValueEvent) { //sync recv
	tr := &jb.trigger2apply
	tr.Lock()
	defer tr.Unlock()
	if apply := tr.syncStatus == nil; !apply {
		apply = !ev.UpdatedAt.Equal(tr.syncStatus.UpdatedAt)
		if !apply {
			return
		}
	}
	tr.syncStatus = &ev.SyncStatus
	tr.dbSyncChanged = 1
	misc.Send2ChanNoBlock(tr.sgn, struct{}{})
}

func (jb *rulesApplier) handleNetlinkEvent(ev NetlinkUpdatesEvent) { //sync recv
	tr := &jb.trigger2apply
	tr.Lock()
	defer tr.Unlock()
	var cnf host.NetConf
	if tr.netConf != nil {
		cnf = tr.netConf.Clone()
	}
	updated := cnf.UpdFromWatcher(ev.Updates...)
	apply := tr.netConf == nil || updated
	tr.netConf = &cnf
	if apply {
		tr.netConfChanged = 1
		misc.Send2ChanNoBlock(tr.sgn, struct{}{})
	}
}

func (jb *rulesApplier) handleDomainAddressesEvent(ctx context.Context, o DomainAddressesEvent) error {
	appliedRules := nft.LastAppliedRules(jb.netNS)
	if appliedRules == nil {
		return nil
	}
	ev := Ask2ResolveDomainAddressesEvent{
		IpVersion: o.IpVersion,
		FQDN:      o.FQDN,
	}
	if o.DnsAnswer.Err == nil {
		ev.ValidBefore = o.DnsAnswer.At.Add(o.DnsAnswer.TTL)
		p := nft.UpdateFqdnNetsets{
			IPVersion: o.IpVersion,
			FQDN:      o.FQDN,
			Addresses: o.DnsAnswer.IPs,
		}
		if err := jb.nftRuleApplier.ApplyPatch(ctx, *appliedRules, p); err != nil {
			if errors.Is(err, nft.ErrPatchNotApplicable) {
				return nil
			}
			return err
		}
	}
	jb.subject.Notify(ev)
	return nil
}

func (jb *rulesApplier) doApply(ctx context.Context, client client.Clients, nc host.NetConf) error {
	const maxLoadDuration = time.Minute

	log := logger.FromContext(ctx)
	localDataLoader := resources.LocalDataLoader{
		MaxLoadDuration: maxLoadDuration,
		DefPolicyAccept: jb.defPolicyAccept,
	}
	localData, err := localDataLoader.Load(ctx, client, nc)
	if err != nil {
		return err
	}
	log.Debug("local data are loaded")
	doApply := true
	if data := nft.LastAppliedRules(jb.netNS); data != nil {
		doApply = !data.LocalData.IsEq(localData)
		if !doApply {
			log.Debug("local data did not change since last load; new rules will not generate")
		}
	}

	fqdnStrategy, err := config.FqdnStrategy.Value(ctx)
	if err != nil {
		return err
	}

	fqdnFromRules := func(rlType domain.ResourceType) ([]domain.FQDN, error) {
		return lo.MapErr(localData.Rules.At(rlType).Values(), func(rule domain.Rule, _ int) (domain.FQDN, error) {
			fqdn, e := domain.FqdnFromSpec(rule.Spec.Remote)
			return fqdn.Value, e
		})
	}

	if doApply {
		var fqdns []domain.FQDN
		localData.ResolvedFQDN = new(resources.ResolvedFQDN)
		if fqdnStrategy.Eq(config.FqdnRulesStartegyDNS) && localData.Rules.At(domain.Ag2FqdnRule).Len() > 0 {
			if fqdns, err = fqdnFromRules(domain.Ag2FqdnRule); err != nil {
				return err
			}
			log.Debug("resolve FQDN(s) for AG")
			localData.ResolvedFQDN.Resolve(ctx, fqdns, app.GetDnsResolver())
		}
		if fqdnStrategy.Eq(config.FqdnRulesStartegyDNS) && localData.Rules.At(domain.Svc2FqdnRule).Len() > 0 {
			if fqdns, err = fqdnFromRules(domain.Svc2FqdnRule); err != nil {
				return err
			}
			log.Debug("resolve FQDN(s) for SVC")
			localData.ResolvedFQDN.Resolve(ctx, fqdns, app.GetDnsResolver())
		}
		var appliedRules nft.AppliedRules
		if appliedRules, err = jb.nftRuleApplier.ApplyConfig(ctx, localData); err != nil {
			return err
		}
		nft.LastAppliedRulesUpd(jb.netNS, &appliedRules)
		ev := AppliedConfEvent{
			NetConf:      nc,
			AppliedRules: appliedRules,
		}
		jb.subject.Notify(ev)
		jb.appliedCount++
	}
	if fqdnStrategy.Eq(config.FqdnRulesStartegyDNS) {
		applied := nft.LastAppliedRules(jb.netNS)
		jb.enqueFQDNs(applied)
	}
	if doApply {
		err = jb.syncHostInfo(ctx, client, nc)
	}
	return err
}

func (jb *rulesApplier) enqueFQDNs(applied *nft.AppliedRules) {
	if applied == nil {
		return
	}
	reqs := make([]observer.EventType, 0,
		applied.LocalData.ResolvedFQDN.A.Len()+
			applied.LocalData.ResolvedFQDN.AAAA.Len())

	sources := misc.Sli(applied.LocalData.ResolvedFQDN.A,
		applied.LocalData.ResolvedFQDN.AAAA)

	for i, ipV := range misc.Sli(iplib.IP4Version, iplib.IP6Version) {
		sources[i].Iterate(func(domain domain.FQDN, addr dns.DomainAddresses) bool {
			ev := Ask2ResolveDomainAddressesEvent{
				IpVersion:   ipV,
				FQDN:        domain,
				ValidBefore: addr.At.Add(addr.TTL),
			}
			reqs = append(reqs, ev)
			return true
		})
	}
	jb.subject.Notify(reqs...)
}

func (jb *rulesApplier) syncHostInfo(ctx context.Context, client client.Clients, nc host.NetConf) (err error) {
	const api = "rules-applier/syncHostInfo"
	defer func() {
		err = errors.WithMessage(err, api)
	}()

	log := logger.FromContext(ctx).Named(api)
	if data := nft.LastAppliedRules(jb.netNS); data != nil {
		host := data.LocalData.LocalHost
		if !host.Syncable() {
			return nil
		}
		log.Debug("sending host info...")
		err = host.Sync(ctx, client, nc)
		if errors.Is(err, resources.ErrHostNotFound) {
			log.Warnf("host %s is not registered in sgroups registry", host.ResourceID())
			err = nil
		}
	}

	return err
}
