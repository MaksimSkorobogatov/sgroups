//go:build linux

package nft

import (
	"container/list"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/dns"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/resources"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/backoff"
	di "github.com/H-BF/corlib/pkg/dict"
	netrc "github.com/H-BF/corlib/pkg/net/resources"
	"github.com/c-robinson/iplib"
	nftlib "github.com/google/nftables"
	nfte "github.com/google/nftables/expr"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

type (
	jobf = func(tx *Tx) error

	jobItem struct {
		name string
		jobf
	}

	jobGroup struct {
		di.RBDict[int16, []jobItem]
		bt *batch
	}

	batch struct {
		rulesApplierOpts
		log        logger.TypeOfLogger
		txProvider func() (*Tx, error)
		data       resources.LocalData
		table      *nftlib.Table
		addrsets   di.HDict[string, *nftlib.Set]
		chains     di.HDict[string, *nftlib.Chain]
		jobs       *list.List
	}
)

func newBatch(
	ctx context.Context,
	o rulesApplierOpts,
	txProvider func() (*Tx, error),
	data resources.LocalData,
) *batch {
	return &batch{
		rulesApplierOpts: o,
		log:              logger.FromContext(ctx),
		txProvider:       txProvider,
		data:             data,
	}
}

func (bt *batch) execute(ctx context.Context) (err error) {
	var (
		it jobItem
		tx *Tx
	)
	defer func() {
		if tx != nil {
			_ = tx.Close()
		}
	}()
	if err = bt.prepare(); err != nil {
		return err
	}
	if bt.dryRun {
		bt.log.Info("dry-run: skip nftables flush")
		return nil
	}

	bkf := makeBatchBackoff(ctx)
loop:
	for el := bt.jobs.Front(); el != nil; el = bt.jobs.Front() {
		it = bt.jobs.Remove(el).(jobItem)
		bkf.Reset()
		for {
			if tx == nil {
				tx, err = bt.txProvider()
				if err != nil {
					return err
				}
			}
			if err = it.jobf(tx); err != nil {
				break loop
			}
			if err = tx.Flush(); err == nil {
				break
			}
			_ = tx.Close()
			tx = nil
			d := bkf.NextBackOff()
			if d == backoff.Stop {
				break loop
			}
			bt.log.Debugf("'%s' will retry after %s", it.name, d)
			timer := time.NewTimer(d)
			select {
			case <-ctx.Done():
				timer.Stop()
				err = ctx.Err()
				break loop
			case <-timer.C:
			}
		}
	}
	if err != nil && len(it.name) > 0 {
		err = errors.WithMessage(err, it.name)
	}
	return err
}

func (bt *batch) prepare() (err error) {
	bt.addrsets.Clear()
	bt.chains.Clear()
	bt.jobs = nil
	bt.table = nil

	bt.initTable()
	bt.addBaseRulesNetSets()
	bt.addLocalHostIPSets()
	bt.addAgNetSets()
	bt.addFQDNNetSets()
	bt.initRootChains()
	for _, dir := range misc.Sli(dirIN, dirOUT) {
		if err = bt.initBaseRules(dir); err != nil {
			return err
		}
		if err = bt.makeInOutChains(dir); err != nil {
			return err
		}
	}
	bt.fwInOutAddDefaultRules()
	bt.switch2NewConfig()
	return nil
}

func (bt *batch) initTable() {
	bt.table = &nftlib.Table{
		Name:   nameUtils{}.genMainTableName(),
		Family: nftlib.TableFamilyINet,
		Flags:  unix.NFT_TABLE_F_DORMANT,
	}
	bt.addJob("init-table", func(tx *Tx) error {
		tlist, e := tx.ListTablesOfFamily(nftlib.TableFamilyINet)
		if e != nil {
			return e
		}
		newTableName := nameUtils{}.genMainTableName()
		for _, o := range tlist {
			if o.Name == newTableName {
				bt.log.Debugf("delete table '%s'", newTableName)
				tx.DelTable(o)
				if e := tx.Flush(); e != nil {
					return e
				}
				break
			}
		}
		bt.log.Debugf("add table '%s'", newTableName)
		bt.table = tx.AddTable(&nftlib.Table{
			Name:   newTableName,
			Family: nftlib.TableFamilyINet,
			Flags:  unix.NFT_TABLE_F_DORMANT,
		})
		return nil
	})
}

func (bt *batch) addBaseRulesNetSets() {
	for n := range bt.baseRules {
		rule := bt.baseRules[n]
		var nws []net.IPNet
		for i := range rule.Nets {
			nws = append(nws, *rule.Nets[i].IPNet)
		}
		for i, nw := range misc.Sli(misc.SeparateNetworks(nws)) {
			isIP4 := i == 0
			elems := setsUtils{}.nets2SetElements(nw, misc.Tern(isIP4, iplib.IP4Version, iplib.IP6Version))
			if len(elems) == 0 {
				continue
			}
			name := nameUtils{}.nameOfBaseRuleNetSet(misc.Tern(isIP4, iplib.IP4Version, iplib.IP6Version), n)
			job := fmt.Sprintf("add '%s'", name)
			bt.addJob(job, func(tx *Tx) error {
				netSet := &nftlib.Set{
					ID:       nextSetID(),
					KeyType:  misc.Tern(isIP4, nftlib.TypeIPAddr, nftlib.TypeIP6Addr),
					Name:     name,
					Constant: true,
					Table:    bt.table,
					Interval: true,
				}
				err := tx.AddSet(netSet, elems)
				if err == nil {
					bt.addrsets.Insert(name, netSet)
					bt.log.Debug(job)
				}
				return err
			})
		}
	}
}

func (bt *batch) addLocalHostIPSets() {
	ipV4, ipV6 := bt.data.LocalHost.Spec.IPs.IPv4.Values(), bt.data.LocalHost.Spec.IPs.IPv6.Values()
	for i, addrs := range misc.Sli(ipV4, ipV6) {
		isV6 := i > 0
		ipV := misc.Tern(isV6, iplib.IP6Version, iplib.IP4Version)
		if elements := (setsUtils{}).addrs2SetElements(addrs, ipV); len(elements) > 0 {
			bt.addJob("add-host-ip-set", func(tx *Tx) error {
				nameOfSet := nameUtils{}.nameOfHostNetSet(ipV, bt.data.LocalHost.Metadata.ID.NamespacedName())
				netSet := &nftlib.Set{
					ID:       nextSetID(),
					Constant: true,
					Table:    bt.table,
					KeyType:  misc.Tern(isV6, nftlib.TypeIP6Addr, nftlib.TypeIPAddr),
					Interval: true,
					Name:     nameOfSet,
				}
				if err := tx.AddSet(netSet, elements); err != nil {
					return err
				}
				bt.addrsets.Put(nameOfSet, netSet)
				bt.log.Debugf("add NET set '%s'/'%s' with items:[%s]",
					bt.table.Name, nameOfSet, misc.Slice2Stringer(addrs...))
				return nil
			})
		}
	}
}

func (bt *batch) addAgNetSets() {
	type setItem struct {
		isV6     bool
		elements []nftlib.SetElement
		items    []any
	}
	var setDict di.HDict[string, *setItem]
	addSet := func(setName string, isV6 bool, elements []nftlib.SetElement, items any) {
		if len(elements) == 0 {
			return
		}
		v := setDict.At(setName)
		if v == nil {
			v = &setItem{isV6: isV6}
			setDict.Put(setName, v)
		}
		v.elements = append(v.elements, elements...)
		v.items = append(v.items, items)
	}

	processSet := func(ag domain.ResourceIdentifier, ipV int, items any, toElems func() []nftlib.SetElement) {
		elements := toElems()
		agSetName := nameUtils{}.nameOfNetSet(ipV, ag.String())
		addSet(agSetName, ipV == iplib.IP6Version, elements, items)
		for _, svc := range bt.data.Services.GetSvcByAgID(ag) {
			svcSetName := nameUtils{}.nameOfSvcNetSet(ipV, svc.Metadata.ID.NamespacedName())
			addSet(svcSetName, ipV == iplib.IP6Version, elements, items)
		}
	}

	for _, nw := range bt.data.Networks.Iterate {
		ipV := misc.Tern(nw.Spec.CIDR.IsV6(), iplib.IP6Version, iplib.IP4Version)
		nets := misc.Sli(nw.Spec.CIDR.IPNet)
		for _, ag := range nw.GetAddressGroupRefs() {
			processSet(ag, ipV, nets, func() []nftlib.SetElement {
				return slice2SetElements(nets, ipV)
			})
		}
	}

	for _, host := range bt.data.Hosts.Iterate {
		for i, addrs := range misc.Sli(host.Spec.IPs.IPv4.Values(), host.Spec.IPs.IPv6.Values()) {
			ipV := misc.Tern(i > 0, iplib.IP6Version, iplib.IP4Version)
			for _, ag := range host.GetAddressGroupRefs() {
				processSet(ag, ipV, addrs, func() []nftlib.SetElement {
					return slice2SetElements(addrs, ipV)
				})
			}
		}
	}

	for setName, set := range setDict.Iterate {
		elements := setsUtils{}.mergeSetElements(set.elements)
		if len(elements) == 0 {
			continue
		}
		bt.addJob("add-ag-net-set", func(tx *Tx) error {
			netSet := &nftlib.Set{
				ID:       nextSetID(),
				Constant: false,
				Table:    bt.table,
				KeyType:  misc.Tern(set.isV6, nftlib.TypeIP6Addr, nftlib.TypeIPAddr),
				Interval: true,
				Name:     setName,
			}
			if err := tx.AddSet(netSet, elements); err != nil {
				return err
			}
			bt.addrsets.Put(setName, netSet)
			bt.log.Debugf("add NET set '%s'/'%s' with items:[%s]",
				bt.table.Name, setName, misc.Slice2Stringer(set.items...))
			return nil
		})
	}
}

func (bt *batch) addFQDNNetSets() {
	if !bt.fqdnStrategy.Eq(config.FqdnRulesStartegyDNS) {
		return
	}
	f := func(IPv int, domain domain.FQDN, a dns.DomainAddresses) {
		bt.addJob("add-fqdn-net-sets", func(tx *Tx) error {
			nameOfSet := nameUtils{}.nameOfFqdnNetSet(IPv, domain)
			nets := make([]net.IPNet, len(a.IPs))
			isV6 := IPv == iplib.IP6Version
			bits := misc.Tern(isV6, net.IPv6len, net.IPv4len) * 8
			mask := net.CIDRMask(bits, bits)
			for i, ip := range a.IPs {
				nets[i] = net.IPNet{IP: ip, Mask: mask}
			}
			elements := (setsUtils{}).nets2SetElements(nets, IPv)
			netSet := &nftlib.Set{
				ID:       nextSetID(),
				Table:    bt.table,
				KeyType:  misc.Tern(isV6, nftlib.TypeIP6Addr, nftlib.TypeIPAddr),
				Interval: true,
				Name:     nameOfSet,
			}
			if err := tx.AddSet(netSet, elements); err != nil {
				return err
			}
			bt.addrsets.Put(nameOfSet, netSet)
			bt.log.Debugf("add network-set '%s'/'%s' items:[%s]",
				bt.table.Name, nameOfSet, misc.Slice2Stringer(nets...))
			if len(nets) == 0 {
				bt.log.Warnf("add IP-set '%s'/'%s' no any IP%v address is resolved for domain '%s'",
					bt.table.Name, nameOfSet, IPv, domain)
			} else {
				bt.log.Debugf("add IP-set '%s'/'%s' with items:[%s]",
					bt.table.Name, nameOfSet, misc.Slice2Stringer(nets...))
			}
			return nil
		})
	}

	bt.data.ResolvedFQDN.A.Iterate(func(domain domain.FQDN, a dns.DomainAddresses) bool {
		f(iplib.IP4Version, domain, a)
		return true
	})
	bt.data.ResolvedFQDN.AAAA.Iterate(func(domain domain.FQDN, a dns.DomainAddresses) bool {
		f(iplib.IP6Version, domain, a)
		return true
	})
}

func (bt *batch) initRootChains() {
	priority := nftlib.ChainPriorityRef(-130)
	policy := misc.Tern(bt.data.DefPolicyAccept, nftlib.ChainPolicyAccept, nftlib.ChainPolicyDrop)
	if bt.data.LocalAGs.Len() > 0 {
		policy = misc.Tern(bt.data.LocalAGs.HasDenyAction(), nftlib.ChainPolicyDrop, nftlib.ChainPolicyAccept)
	}

	bt.addJob("init root chains", func(tx *Tx) error {
		chains := misc.Sli(
			&nftlib.Chain{
				Name:     chnIngressMain,
				Table:    bt.table,
				Type:     nftlib.ChainTypeFilter,
				Policy:   misc.Val2Ptr(policy),
				Hooknum:  nftlib.ChainHookPrerouting,
				Priority: priority,
			},
			&nftlib.Chain{
				Name:     chnEgressMain,
				Table:    bt.table,
				Type:     nftlib.ChainTypeFilter,
				Policy:   misc.Val2Ptr(policy),
				Hooknum:  nftlib.ChainHookPostrouting,
				Priority: nftlib.ChainPriorityConntrackHelper,
			},
		)
		for i := range chains {
			chain := tx.AddChain(chains[i])
			BeginRule().
				CTState(nfte.CtStateBitESTABLISHED|nfte.CtStateBitRELATED).
				Counter().Accept().ApplyRule(chain, tx.Conn)
			bt.chains.Put(chain.Name, chain)
			bt.log.Debugf("add chain '%s'/'%s'", bt.table.Name, chain.Name)
		}
		return nil
	})
}

func (bt *batch) initBaseRules(dir direction) (err error) {
	rules := bt.baseRules
	for i := 0; i < len(rules) && err == nil; i++ {
		err = bt.initBaseRuleItem(dir, i)
	}
	return err
}

func (bt *batch) initBaseRuleItem(dir direction, ruleN int) (err error) {
	const api = "init base rule item"
	defer func() {
		err = errors.WithMessagef(err, "%s[#%v]", api, ruleN)
	}()
	if !misc.IsIn(dir, misc.Sli(dirIN, dirOUT)...) {
		panic("initBaseRuleItem(unreachable code)")
	}
	rule := bt.baseRules[ruleN]
	haveEgr, haveIngr := !rule.Egress.IsNone(), !rule.Ingress.IsNone()
	if misc.Tern(dir == dirIN, !haveIngr && haveEgr, !haveEgr && haveIngr) {
		return nil
	}
	attrsOpt := misc.Tern(dir == dirIN, rule.Ingress, rule.Egress)
	chnDest := misc.Tern(dirIN == dir,
		chnIngressMain,
		chnEgressMain,
	)
	baseRuleName := fmt.Sprintf("base-rule-%v", ruleN)
	for _, ipV := range misc.Sli(iplib.IP4Version, iplib.IP6Version) {
		jobPrefix := fmt.Sprintf("%s; %sgress; IPV%v", baseRuleName, misc.Tern(dirIN == dir, "in", "e"), ipV)
		isIP4 := iplib.IP4Version == ipV
		setName := nameUtils{}.nameOfBaseRuleNetSet(ipV, ruleN)
		var acts []func(rb ruleBuilder) ruleBuilder
		var jobTitles []string
		if attr, _ := attrsOpt.Maybe(); attr.IsEmpty() {
			jobTitles = append(jobTitles, jobPrefix)
			acts = append(acts, func(rb ruleBuilder) ruleBuilder {
				return rb
			})
		} else {
			for k, opt := range misc.Sli(attr.TCP, attr.UDP) {
				if p, ok := opt.Maybe(); ok {
					var ports accports
					if ports, err = transormPortRanges(p.Ports); err != nil {
						return err
					}
					if len(ports) == 0 {
						ports = append(ports, [2]netrc.PortNumber{1, 65535})
					}
					jobTitles = append(jobTitles, fmt.Sprintf("%s; %s", jobPrefix, misc.Tern(k == 0, "tcp", "udp")))
					acts = append(acts, func(rb ruleBuilder) ruleBuilder {
						return ports.D(
							rb.ProtoIP(misc.Tern(k == 0, netrc.TCP, netrc.UDP)),
						)
					})
				}
			}
			if icmp, ok := misc.Tern(isIP4, &attr.ICMP, &attr.ICMP6).Maybe(); ok {
				jobTitles = append(jobTitles, fmt.Sprintf("%s; icmp", jobPrefix))
				acts = append(acts, func(rb ruleBuilder) ruleBuilder {
					rc := netrc.ICMP{
						IPv: misc.Tern(isIP4, netrc.IPv4, netrc.IPv6),
					}
					rc.Types.PutMany(icmp.Types.Values()...)
					return rb.ProtoICMP(rc)
				})
			}
		}
		for i := range acts {
			act, job := acts[i], jobTitles[i]
			bt.addJob(job, func(tx *Tx) error {
				if netSet := bt.addrsets.At(setName); netSet != nil {
					bt.log.Debugf("add %s into '%s'/'%s'",
						job, bt.table.Name, chnDest)
					rb := BeginRule()
					rb = misc.Tern(dirIN == dir, rb.SAddr, rb.DAddr)(ipV).InSet(netSet)
					act(rb).Accept().
						ApplyRule(bt.chains.At(chnDest), tx.Conn)
				}
				return nil
			})
		}
	}
	return nil
}

func (bt *batch) makeInOutChains(dir direction) (err error) {
	defer func() {
		err = errors.WithMessagef(err, "make %s chain", misc.Tern(dir == dirIN, "in", "out"))
	}()
	gr := [...]func(self *jobGroup, dir direction, ag domain.AddressGroup) error{
		(*jobGroup).populateBothtSvcSvcRules,   // base-pri(-350)
		(*jobGroup).populateBothAgAgIcmpRules,  // base-pri(-300)
		(*jobGroup).populateBothAgAgRules,      // base-pri(-200)
		(*jobGroup).populateIESvcCidrIcmpRules, //
		(*jobGroup).populateIESvcSvcRules,      //
		(*jobGroup).populateIESvcCidrRules,     // base-pri(-150)
		(*jobGroup).populateIEAgAgIcmpRules,    // base-pri(-100)
		(*jobGroup).populateIEAgAgRules,        // base-pri(0)
		(*jobGroup).populateIEAgIcmpRules,      // base-pri(-400) IE
		(*jobGroup).populateOutSvcFqdnRules,    // base-pri(50)
		(*jobGroup).populateOutAgFqdnRules,     // base-pri(100)
		(*jobGroup).populateIEAgCidrIcmpRules,  // base-pri(200)
		(*jobGroup).populateIEAgCidrRules,      // base-pri(300)
		(*jobGroup).populateIEAgSvcRule,
		(*jobGroup).populateIEAgSvcIcmpRule,
		(*jobGroup).populateBothAgSvcRules,
		(*jobGroup).populateBothAgSvcIcmpRules,
		(*jobGroup).populateBothSvcAgRule,
		(*jobGroup).populateBothSvcAgIcmpRule,
		(*jobGroup).populateIESvcAgRule,
		(*jobGroup).populateIESvcAgIcmpRule,
	}
	for _, ag := range bt.data.LocalAGs.Iterate {
		bt.chainInOutProlog(dir, ag)

		if err = bt.withGroup(func(g *jobGroup) (e error) {
			for _, f := range gr {
				if e = f(g, dir, ag); e != nil {
					return e
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func (gp *jobGroup) populateBothAgSvcRules(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-%s-ag-svc-rule", misc.Tern(isIN, "in", "out"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	bt := gp.bt
	targetSGchName := bt.inOutChainName(dir, ag)

	ruleType := domain.Ag2SvcRule

	allRules := bt.data.Rules.At(ruleType).BothTraffic()
	if allRules == nil {
		return nil
	}

	emit := func(rule domain.Rule, svc domain.Service, addrSetForFamily func(ipV int) string) error {
		rlName := rule.Metadata.ID.NamespacedName()
		pri, err := rule.Priority()
		if err != nil {
			return errors.WithMessagef(err, "invalid priority for rule '%s'", rlName)
		}
		for _, transport := range svc.Spec.Transports {
			if _, isL4 := transport.(domain.L4Transport); !isL4 {
				continue
			}
			ipV := transport.GetIPv()
			addrSetName := addrSetForFamily(int(ipV))
			l4proto := transport.GetProto().L4Proto()
			proto, ok := l4proto.Maybe()
			if !ok {
				return errors.Errorf("invalid transport protocol for rule '%s'", rlName)
			}
			ports := setsUtils{}.makeAccPorts(domain.PortFromTransport(transport))
			if len(ports) == 0 {
				ports = append(ports, accports{})
			}
			for i := range ports { //nolint:dupl
				ports := ports[i]
				gp.addJob(pri, api, func(tx *Tx) error {
					chnApplyTo := bt.chains.At(targetSGchName)
					addrSet := bt.addrsets.At(addrSetName)
					if chnApplyTo == nil || addrSet == nil {
						return nil
					}
					bt.log.Debugf("add '%s' %s-ag-svc-rule for accports(%s)/addr-set '%s' into '%s'/'%s' with priority(%v)",
						rlName, misc.Tern(isIN, "in", "out"),
						ports, addrSetName, bt.table.Name, targetSGchName, pri)
					rb := BeginRule()
					rb = ports.D(
						misc.Tern(isIN, rb.SAddr, rb.DAddr)(int(ipV)).InSet(addrSet).
							ProtoIP(proto),
					).MetaNFTRACE(rule.IsTraceOn()).
						Counter()
					if rule.IsLogOn() {
						rb = rb.DLogs(nfte.LogFlagsIPOpt)
					}
					rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
					return nil
				})
			}
		}
		return nil
	}

	// Case A: this AG is the LOCAL endpoint of the rule (rule.local == ag).
	// The other side is the rule.remote Svc. Filter chain by remote-Svc addr-set;
	// since rule.Traffic == BOTH the same rule is installed both into the ingress
	// and the egress chain of `ag` (this function is invoked once per direction).
	if !isIN {
		for _, ref := range ag.GetRefsByTypes(ruleType) {
			rule, ok := allRules.Get(ref)
			if !ok {
				continue
			}
			local, err := domain.LocalFromSpec(rule.Spec.Local)
			if err != nil {
				return errors.WithMessagef(err, "invalid local for rule '%s'", ref)
			}
			if local.ResourceIdentifier != ag.Metadata.ID.ResourceID() {
				continue
			}
			remote, err := domain.RemoteFromSpec(rule.Spec.Remote)
			if err != nil {
				return errors.WithMessagef(err, "invalid remote for rule '%s'", ref)
			}
			remoteSvc, ok := bt.data.Services.Get(remote.ResourceIdentifier)
			if !ok {
				continue
			}
			if err = emit(rule, remoteSvc, func(ipV int) string {
				return nameUtils{}.nameOfSvcNetSet(ipV, remoteSvc.ResourceID().String())
			}); err != nil {
				return err
			}
		}
	}

	// Case B: this AG hosts the Svc that is the REMOTE endpoint of the rule
	// (rule.remote ∈ ag.services). The other side is the rule.local AG, so we
	// filter the chain by that AG's net-set. Ports/proto are still taken from
	// the Svc (which lives in this AG).
	if isIN {
		for _, svcID := range ag.GetServiceRefs() {
			svc, ok := bt.data.Services.Get(svcID)
			if !ok {
				continue
			}
			for _, ref := range svc.GetRefsByTypes(ruleType) {
				rule, ok := allRules.Get(ref)
				if !ok {
					continue
				}
				remote, err := domain.RemoteFromSpec(rule.Spec.Remote)
				if err != nil {
					return errors.WithMessagef(err, "invalid remote for rule '%s'", ref)
				}
				if remote.ResourceIdentifier != svcID {
					continue
				}
				local, err := domain.LocalFromSpec(rule.Spec.Local)
				if err != nil {
					return errors.WithMessagef(err, "invalid local for rule '%s'", ref)
				}
				if err = emit(rule, svc, func(ipV int) string {
					return nameUtils{}.nameOfNetSet(ipV, local.NamespacedName())
				}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (gp *jobGroup) populateBothAgSvcIcmpRules(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-%s-ag-svc-icmp-rule", misc.Tern(isIN, "in", "out"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	bt := gp.bt
	targetSGchName := bt.inOutChainName(dir, ag)

	ruleType := domain.Ag2SvcIcmpRule

	allRules := bt.data.Rules.At(ruleType).BothTraffic()
	if allRules == nil {
		return nil
	}

	emit := func(rule domain.Rule, svc domain.Service, addrSetForFamily func(ipV int) string) error {
		rlName := rule.Metadata.ID.NamespacedName()
		pri, err := rule.Priority()
		if err != nil {
			return errors.WithMessagef(err, "invalid priority for rule '%s'", rlName)
		}
		for _, transport := range svc.Spec.Transports {
			for _, icmp := range domain.IcmpFromTransport(transport) {
				addrSetName := addrSetForFamily(int(icmp.IPv))
				gp.addJob(pri, api, func(tx *Tx) error {
					chnApplyTo := bt.chains.At(targetSGchName)
					addrSet := bt.addrsets.At(addrSetName)
					if chnApplyTo == nil || addrSet == nil {
						return nil
					}
					bt.log.Debugf("add '%s' %s-ag-svc-icmp%v-rule for addr-set '%s' into '%s'/'%s' with priority(%v)",
						rlName, misc.Tern(isIN, "in", "out"),
						misc.Tern(icmp.IPv == netrc.IPv6, "6", ""),
						addrSetName, targetSGchName, bt.table.Name, pri)
					rb := BeginRule()
					rb = misc.Tern(isIN, rb.SAddr, rb.DAddr)(int(icmp.IPv)).
						InSet(addrSet).
						ProtoICMP(icmp).
						MetaNFTRACE(rule.IsTraceOn()).
						Counter()
					if rule.IsLogOn() {
						rb = rb.DLogs(nfte.LogFlagsIPOpt)
					}
					rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
					return nil
				})
			}
		}
		return nil
	}

	// Case A: this AG is the LOCAL endpoint of the rule (rule.local == ag).
	if !isIN {
		for _, ref := range ag.GetRefsByTypes(ruleType) {
			rule, ok := allRules.Get(ref)
			if !ok {
				continue
			}
			local, err := domain.LocalFromSpec(rule.Spec.Local)
			if err != nil {
				return errors.WithMessagef(err, "invalid local for rule '%s'", ref)
			}
			if local.ResourceIdentifier != ag.Metadata.ID.ResourceID() {
				continue
			}
			remote, err := domain.RemoteFromSpec(rule.Spec.Remote)
			if err != nil {
				return errors.WithMessagef(err, "invalid remote for rule '%s'", ref)
			}
			remoteSvc, ok := bt.data.Services.Get(remote.ResourceIdentifier)
			if !ok {
				continue
			}
			if err = emit(rule, remoteSvc, func(ipV int) string {
				return nameUtils{}.nameOfSvcNetSet(ipV, remoteSvc.ResourceID().String())
			}); err != nil {
				return err
			}
		}
	}

	// Case B: this AG hosts the Svc that is the REMOTE endpoint of the rule.
	if isIN {
		for _, svcID := range ag.GetServiceRefs() {
			svc, ok := bt.data.Services.Get(svcID)
			if !ok {
				continue
			}
			for _, ref := range svc.GetRefsByTypes(ruleType) {
				rule, ok := allRules.Get(ref)
				if !ok {
					continue
				}
				remote, err := domain.RemoteFromSpec(rule.Spec.Remote)
				if err != nil {
					return errors.WithMessagef(err, "invalid remote for rule '%s'", ref)
				}
				if remote.ResourceIdentifier != svcID {
					continue
				}
				local, err := domain.LocalFromSpec(rule.Spec.Local)
				if err != nil {
					return errors.WithMessagef(err, "invalid local for rule '%s'", ref)
				}
				if err = emit(rule, svc, func(ipV int) string {
					return nameUtils{}.nameOfNetSet(ipV, local.NamespacedName())
				}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (gp *jobGroup) populateIEAgSvcRule(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-ag-%sgress-svc-rule", misc.Tern(isIN, "in", "e"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	bt := gp.bt
	targetSGchName := bt.inOutChainName(dir, ag)

	ruleType := domain.Ag2SvcRule
	rules := bt.data.Rules.At(ruleType).Traffic(misc.Tern(isIN, domain.INGRESS, domain.EGRESS))
	if rules == nil {
		return nil
	}

	for _, ref := range ag.GetRefsByTypes(ruleType) {
		rule, ok := rules.Get(ref)
		if !ok {
			continue
		}
		rlName := rule.Metadata.ID.NamespacedName()
		local, err := domain.LocalFromSpec(rule.Spec.Local)
		if err != nil {
			return errors.WithMessagef(err, "invalid local for rule '%s'", rlName)
		}
		if local.ResourceIdentifier != ag.Metadata.ID.ResourceID() {
			continue
		}
		remote, err := domain.RemoteFromSpec(rule.Spec.Remote)
		if err != nil {
			return errors.WithMessagef(err, "invalid remote for rule '%s'", rlName)
		}
		remoteSvc, ok := bt.data.Services.Get(remote.ResourceIdentifier)
		if !ok {
			continue
		}
		pri, err := rule.Priority()
		if err != nil {
			return errors.WithMessagef(err, "invalid priority for rule '%s'", rlName)
		}

		emitL4 := func(transport domain.TransportSpec, fromRule bool) error {
			ipV := transport.GetIPv()
			l4proto := transport.GetProto().L4Proto()
			proto, ok := l4proto.Maybe()
			if !ok {
				return errors.Errorf("invalid transport protocol for rule '%s'", rlName)
			}
			addrSetName := nameUtils{}.nameOfSvcNetSet(int(ipV), remoteSvc.ResourceID().String())
			ports := setsUtils{}.makeAccPorts(domain.PortFromTransport(transport))
			if len(ports) == 0 {
				ports = append(ports, accports{})
			}
			for i := range ports { //nolint:dupl
				ports := ports[i]
				gp.addJob(pri, api, func(tx *Tx) error {
					chnApplyTo := bt.chains.At(targetSGchName)
					addrSet := bt.addrsets.At(addrSetName)
					if chnApplyTo == nil || addrSet == nil {
						return nil
					}
					bt.log.Debugf("add '%s' ag-%sgress-svc-rule (transport from %s) for accports(%s)/addr-set '%s' into '%s'/'%s' with priority(%v)",
						rlName, misc.Tern(isIN, "in", "e"),
						misc.Tern(fromRule, "rule", "svc"),
						ports, addrSetName, bt.table.Name, targetSGchName, pri)
					rb := BeginRule()
					rb = ports.D(
						misc.Tern(isIN, rb.SAddr, rb.DAddr)(int(ipV)).InSet(addrSet).
							ProtoIP(proto),
					).MetaNFTRACE(rule.IsTraceOn()).
						Counter()
					if rule.IsLogOn() {
						rb = rb.DLogs(nfte.LogFlagsIPOpt)
					}
					rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
					return nil
				})
			}
			return nil
		}

		if isIN {
			// INGRESS → L4 transport from rule.
			if _, isL4 := rule.Spec.Transport.(domain.L4Transport); !isL4 {
				continue
			}
			if err = emitL4(rule.Spec.Transport, true); err != nil {
				return err
			}
			continue
		}
		// EGRESS → L4 transports from svc (ICMP handled by populateIEAgSvcIcmpRule).
		for _, transport := range remoteSvc.Spec.Transports {
			if _, isL4 := transport.(domain.L4Transport); !isL4 {
				continue
			}
			if err = emitL4(transport, false); err != nil {
				return err
			}
		}
	}
	return nil
}

func (gp *jobGroup) populateIEAgSvcIcmpRule(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-ag-%sgress-svc-icmp-rule", misc.Tern(isIN, "in", "e"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	bt := gp.bt
	targetSGchName := bt.inOutChainName(dir, ag)

	ruleType := domain.Ag2SvcIcmpRule
	rules := bt.data.Rules.At(ruleType).Traffic(misc.Tern(isIN, domain.INGRESS, domain.EGRESS))
	if rules == nil {
		return nil
	}

	for _, ref := range ag.GetRefsByTypes(ruleType) {
		rule, ok := rules.Get(ref)
		if !ok {
			continue
		}
		rlName := rule.Metadata.ID.NamespacedName()
		local, err := domain.LocalFromSpec(rule.Spec.Local)
		if err != nil {
			return errors.WithMessagef(err, "invalid local for rule '%s'", rlName)
		}
		if local.ResourceIdentifier != ag.Metadata.ID.ResourceID() {
			continue
		}
		remote, err := domain.RemoteFromSpec(rule.Spec.Remote)
		if err != nil {
			return errors.WithMessagef(err, "invalid remote for rule '%s'", rlName)
		}
		remoteSvc, ok := bt.data.Services.Get(remote.ResourceIdentifier)
		if !ok {
			continue
		}
		pri, err := rule.Priority()
		if err != nil {
			return errors.WithMessagef(err, "invalid priority for rule '%s'", rlName)
		}

		emitIcmp := func(transport domain.TransportSpec, fromRule bool) {
			for _, icmp := range domain.IcmpFromTransport(transport) {
				addrSetName := nameUtils{}.nameOfSvcNetSet(int(icmp.IPv), remoteSvc.ResourceID().String())
				gp.addJob(pri, api, func(tx *Tx) error {
					chnApplyTo := bt.chains.At(targetSGchName)
					addrSet := bt.addrsets.At(addrSetName)
					if chnApplyTo == nil || addrSet == nil {
						return nil
					}
					bt.log.Debugf("add '%s' ag-%sgress-svc-icmp%v-rule (transport from %s) for addr-set '%s' into '%s'/'%s' with priority(%v)",
						rlName, misc.Tern(isIN, "in", "e"),
						misc.Tern(icmp.IPv == netrc.IPv6, "6", ""),
						misc.Tern(fromRule, "rule", "svc"),
						addrSetName, targetSGchName, bt.table.Name, pri)
					rb := BeginRule()
					rb = misc.Tern(isIN, rb.SAddr, rb.DAddr)(int(icmp.IPv)).
						InSet(addrSet).
						ProtoICMP(icmp).
						MetaNFTRACE(rule.IsTraceOn()).
						Counter()
					if rule.IsLogOn() {
						rb = rb.DLogs(nfte.LogFlagsIPOpt)
					}
					rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
					return nil
				})
			}
		}

		if isIN {
			// INGRESS → ICMP transport from rule.
			if _, isIcmp := rule.Spec.Transport.(domain.IcmpTransport); !isIcmp {
				continue
			}
			emitIcmp(rule.Spec.Transport, true)
			continue
		}
		// EGRESS → ICMP transports from svc.
		for _, transport := range remoteSvc.Spec.Transports {
			if _, isIcmp := transport.(domain.IcmpTransport); !isIcmp {
				continue
			}
			emitIcmp(transport, false)
		}
	}
	return nil
}

func (gp *jobGroup) populateBothSvcAgRule(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-%s-svc-ag-rule", misc.Tern(isIN, "in", "out"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	bt := gp.bt
	targetSGchName := bt.inOutChainName(dir, ag)

	ruleType := domain.Svc2AgRule
	allRules := bt.data.Rules.At(ruleType).BothTraffic()
	if allRules == nil {
		return nil
	}

	// emit installs L4 rule (transport always from rule.Spec.Transport).
	emit := func(rule domain.Rule, addrSetForFamily func(ipV int) string) error {
		rlName := rule.Metadata.ID.NamespacedName()
		if _, isL4 := rule.Spec.Transport.(domain.L4Transport); !isL4 {
			return nil
		}
		pri, err := rule.Priority()
		if err != nil {
			return errors.WithMessagef(err, "invalid priority for rule '%s'", rlName)
		}
		ipV := rule.Spec.Transport.GetIPv()
		l4proto := rule.Spec.Transport.GetProto().L4Proto()
		proto, ok := l4proto.Maybe()
		if !ok {
			return errors.Errorf("invalid transport protocol for rule '%s'", rlName)
		}
		addrSetName := addrSetForFamily(int(ipV))
		ports := setsUtils{}.makeAccPorts(domain.PortFromTransport(rule.Spec.Transport))
		if len(ports) == 0 {
			ports = append(ports, accports{})
		}
		for i := range ports { //nolint:dupl
			ports := ports[i]
			gp.addJob(pri, api, func(tx *Tx) error {
				chnApplyTo := bt.chains.At(targetSGchName)
				addrSet := bt.addrsets.At(addrSetName)
				if chnApplyTo == nil || addrSet == nil {
					return nil
				}
				bt.log.Debugf("add '%s' %s-svc-ag-rule (transport from rule) for accports(%s)/addr-set '%s' into '%s'/'%s' with priority(%v)",
					rlName, misc.Tern(isIN, "in", "out"),
					ports, addrSetName, bt.table.Name, targetSGchName, pri)
				rb := BeginRule()
				rb = ports.D(
					misc.Tern(isIN, rb.SAddr, rb.DAddr)(int(ipV)).InSet(addrSet).
						ProtoIP(proto),
				).MetaNFTRACE(rule.IsTraceOn()).
					Counter()
				if rule.IsLogOn() {
					rb = rb.DLogs(nfte.LogFlagsIPOpt)
				}
				rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
				return nil
			})
		}
		return nil
	}

	// Case A: ag hosts rule.local svc → other side is remote AG.
	if !isIN {
		for _, svcID := range ag.GetServiceRefs() {
			svc, ok := bt.data.Services.Get(svcID)
			if !ok {
				continue
			}
			for _, ref := range svc.GetRefsByTypes(ruleType) {
				rule, ok := allRules.Get(ref)
				if !ok {
					continue
				}
				local, err := domain.LocalFromSpec(rule.Spec.Local)
				if err != nil {
					return errors.WithMessagef(err, "invalid local for rule '%s'", ref)
				}
				if local.ResourceIdentifier != svcID {
					continue
				}
				remote, err := domain.RemoteFromSpec(rule.Spec.Remote)
				if err != nil {
					return errors.WithMessagef(err, "invalid remote for rule '%s'", ref)
				}
				if err = emit(rule, func(ipV int) string {
					return nameUtils{}.nameOfNetSet(ipV, remote.NamespacedName())
				}); err != nil {
					return err
				}
			}
		}
	}

	// Case B: ag == rule.remote AG → other side is local svc.
	if isIN {
		for _, ref := range ag.GetRefsByTypes(ruleType) {
			rule, ok := allRules.Get(ref)
			if !ok {
				continue
			}
			remote, err := domain.RemoteFromSpec(rule.Spec.Remote)
			if err != nil {
				return errors.WithMessagef(err, "invalid remote for rule '%s'", ref)
			}
			if remote.ResourceIdentifier != ag.Metadata.ID.ResourceID() {
				continue
			}
			local, err := domain.LocalFromSpec(rule.Spec.Local)
			if err != nil {
				return errors.WithMessagef(err, "invalid local for rule '%s'", ref)
			}
			localSvc, ok := bt.data.Services.Get(local.ResourceIdentifier)
			if !ok {
				continue
			}
			if err = emit(rule, func(ipV int) string {
				return nameUtils{}.nameOfSvcNetSet(ipV, localSvc.ResourceID().String())
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (gp *jobGroup) populateBothSvcAgIcmpRule(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-%s-svc-ag-icmp-rule", misc.Tern(isIN, "in", "out"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	bt := gp.bt
	targetSGchName := bt.inOutChainName(dir, ag)

	ruleType := domain.Svc2AgIcmpRule
	allRules := bt.data.Rules.At(ruleType).BothTraffic()
	if allRules == nil {
		return nil
	}

	// emit installs ICMP rule (transport always from rule.Spec.Transport).
	emit := func(rule domain.Rule, addrSetForFamily func(ipV int) string) error {
		rlName := rule.Metadata.ID.NamespacedName()
		pri, err := rule.Priority()
		if err != nil {
			return errors.WithMessagef(err, "invalid priority for rule '%s'", rlName)
		}
		for _, icmp := range domain.IcmpFromTransport(rule.Spec.Transport) {
			addrSetName := addrSetForFamily(int(icmp.IPv))
			gp.addJob(pri, api, func(tx *Tx) error {
				chnApplyTo := bt.chains.At(targetSGchName)
				addrSet := bt.addrsets.At(addrSetName)
				if chnApplyTo == nil || addrSet == nil {
					return nil
				}
				bt.log.Debugf("add '%s' %s-svc-ag-icmp%v-rule (transport from rule) for addr-set '%s' into '%s'/'%s' with priority(%v)",
					rlName, misc.Tern(isIN, "in", "out"),
					misc.Tern(icmp.IPv == netrc.IPv6, "6", ""),
					addrSetName, targetSGchName, bt.table.Name, pri)
				rb := BeginRule()
				rb = misc.Tern(isIN, rb.SAddr, rb.DAddr)(int(icmp.IPv)).
					InSet(addrSet).
					ProtoICMP(icmp).
					MetaNFTRACE(rule.IsTraceOn()).
					Counter()
				if rule.IsLogOn() {
					rb = rb.DLogs(nfte.LogFlagsIPOpt)
				}
				rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
				return nil
			})
		}
		return nil
	}

	// Case A: ag hosts rule.local svc → other side is remote AG.
	if !isIN {
		for _, svcID := range ag.GetServiceRefs() {
			svc, ok := bt.data.Services.Get(svcID)
			if !ok {
				continue
			}
			for _, ref := range svc.GetRefsByTypes(ruleType) {
				rule, ok := allRules.Get(ref)
				if !ok {
					continue
				}
				local, err := domain.LocalFromSpec(rule.Spec.Local)
				if err != nil {
					return errors.WithMessagef(err, "invalid local for rule '%s'", ref)
				}
				if local.ResourceIdentifier != svcID {
					continue
				}
				remote, err := domain.RemoteFromSpec(rule.Spec.Remote)
				if err != nil {
					return errors.WithMessagef(err, "invalid remote for rule '%s'", ref)
				}
				if err = emit(rule, func(ipV int) string {
					return nameUtils{}.nameOfNetSet(ipV, remote.NamespacedName())
				}); err != nil {
					return err
				}
			}
		}
	}

	// Case B: ag == rule.remote AG → other side is local svc.
	if isIN {
		for _, ref := range ag.GetRefsByTypes(ruleType) {
			rule, ok := allRules.Get(ref)
			if !ok {
				continue
			}
			remote, err := domain.RemoteFromSpec(rule.Spec.Remote)
			if err != nil {
				return errors.WithMessagef(err, "invalid remote for rule '%s'", ref)
			}
			if remote.ResourceIdentifier != ag.Metadata.ID.ResourceID() {
				continue
			}
			local, err := domain.LocalFromSpec(rule.Spec.Local)
			if err != nil {
				return errors.WithMessagef(err, "invalid local for rule '%s'", ref)
			}
			localSvc, ok := bt.data.Services.Get(local.ResourceIdentifier)
			if !ok {
				continue
			}
			if err = emit(rule, func(ipV int) string {
				return nameUtils{}.nameOfSvcNetSet(ipV, localSvc.ResourceID().String())
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (gp *jobGroup) populateIESvcAgRule(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-svc-%sgress-ag-rule", misc.Tern(isIN, "in", "e"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	bt := gp.bt
	targetSGchName := bt.inOutChainName(dir, ag)

	ruleType := domain.Svc2AgRule
	rules := bt.data.Rules.At(ruleType).Traffic(misc.Tern(isIN, domain.INGRESS, domain.EGRESS))
	if rules == nil {
		return nil
	}

	for _, svcID := range ag.GetServiceRefs() {
		localSvc, ok := bt.data.Services.Get(svcID)
		if !ok {
			continue
		}
		for _, ref := range localSvc.GetRefsByTypes(ruleType) {
			rule, ok := rules.Get(ref)
			if !ok {
				continue
			}
			rlName := rule.Metadata.ID.NamespacedName()
			local, err := domain.LocalFromSpec(rule.Spec.Local)
			if err != nil {
				return errors.WithMessagef(err, "invalid local for rule '%s'", ref)
			}
			if local.ResourceIdentifier != svcID {
				continue
			}
			remote, err := domain.RemoteFromSpec(rule.Spec.Remote)
			if err != nil {
				return errors.WithMessagef(err, "invalid remote for rule '%s'", ref)
			}
			pri, err := rule.Priority()
			if err != nil {
				return errors.WithMessagef(err, "invalid priority for rule '%s'", ref)
			}

			emitL4 := func(transport domain.TransportSpec, fromRule bool) error {
				ipV := transport.GetIPv()
				l4proto := transport.GetProto().L4Proto()
				proto, ok := l4proto.Maybe()
				if !ok {
					return errors.Errorf("invalid transport protocol for rule '%s'", rlName)
				}
				addrSetName := nameUtils{}.nameOfNetSet(int(ipV), remote.NamespacedName())
				ports := setsUtils{}.makeAccPorts(domain.PortFromTransport(transport))
				if len(ports) == 0 {
					ports = append(ports, accports{})
				}
				for i := range ports { //nolint:dupl
					ports := ports[i]
					gp.addJob(pri, api, func(tx *Tx) error {
						chnApplyTo := bt.chains.At(targetSGchName)
						addrSet := bt.addrsets.At(addrSetName)
						if chnApplyTo == nil || addrSet == nil {
							return nil
						}
						bt.log.Debugf("add '%s' svc-%sgress-ag-rule (transport from %s) for accports(%s)/addr-set '%s' into '%s'/'%s' with priority(%v)",
							rlName, misc.Tern(isIN, "in", "e"),
							misc.Tern(fromRule, "rule", "svc"),
							ports, addrSetName, bt.table.Name, targetSGchName, pri)
						rb := BeginRule()
						rb = ports.D(
							misc.Tern(isIN, rb.SAddr, rb.DAddr)(int(ipV)).InSet(addrSet).
								ProtoIP(proto),
						).MetaNFTRACE(rule.IsTraceOn()).
							Counter()
						if rule.IsLogOn() {
							rb = rb.DLogs(nfte.LogFlagsIPOpt)
						}
						rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
						return nil
					})
				}
				return nil
			}

			if isIN {
				// INGRESS → L4 transports from svc (ICMP handled by populateIESvcAgIcmpRule).
				for _, transport := range localSvc.Spec.Transports {
					if _, isL4 := transport.(domain.L4Transport); !isL4 {
						continue
					}
					if err = emitL4(transport, false); err != nil {
						return err
					}
				}
				continue
			}
			// EGRESS → L4 transport from rule.
			if _, isL4 := rule.Spec.Transport.(domain.L4Transport); !isL4 {
				continue
			}
			if err = emitL4(rule.Spec.Transport, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func (gp *jobGroup) populateIESvcAgIcmpRule(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-svc-%sgress-ag-icmp-rule", misc.Tern(isIN, "in", "e"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	bt := gp.bt
	targetSGchName := bt.inOutChainName(dir, ag)

	ruleType := domain.Svc2AgIcmpRule
	rules := bt.data.Rules.At(ruleType).Traffic(misc.Tern(isIN, domain.INGRESS, domain.EGRESS))
	if rules == nil {
		return nil
	}

	for _, svcID := range ag.GetServiceRefs() {
		localSvc, ok := bt.data.Services.Get(svcID)
		if !ok {
			continue
		}
		for _, ref := range localSvc.GetRefsByTypes(ruleType) {
			rule, ok := rules.Get(ref)
			if !ok {
				continue
			}
			rlName := rule.Metadata.ID.NamespacedName()
			local, err := domain.LocalFromSpec(rule.Spec.Local)
			if err != nil {
				return errors.WithMessagef(err, "invalid local for rule '%s'", ref)
			}
			if local.ResourceIdentifier != svcID {
				continue
			}
			remote, err := domain.RemoteFromSpec(rule.Spec.Remote)
			if err != nil {
				return errors.WithMessagef(err, "invalid remote for rule '%s'", ref)
			}
			pri, err := rule.Priority()
			if err != nil {
				return errors.WithMessagef(err, "invalid priority for rule '%s'", ref)
			}

			emitIcmp := func(transport domain.TransportSpec, fromRule bool) {
				for _, icmp := range domain.IcmpFromTransport(transport) {
					addrSetName := nameUtils{}.nameOfNetSet(int(icmp.IPv), remote.NamespacedName())
					gp.addJob(pri, api, func(tx *Tx) error {
						chnApplyTo := bt.chains.At(targetSGchName)
						addrSet := bt.addrsets.At(addrSetName)
						if chnApplyTo == nil || addrSet == nil {
							return nil
						}
						bt.log.Debugf("add '%s' svc-%sgress-ag-icmp%v-rule (transport from %s) for addr-set '%s' into '%s'/'%s' with priority(%v)",
							rlName, misc.Tern(isIN, "in", "e"),
							misc.Tern(icmp.IPv == netrc.IPv6, "6", ""),
							misc.Tern(fromRule, "rule", "svc"),
							addrSetName, targetSGchName, bt.table.Name, pri)
						rb := BeginRule()
						rb = misc.Tern(isIN, rb.SAddr, rb.DAddr)(int(icmp.IPv)).
							InSet(addrSet).
							ProtoICMP(icmp).
							MetaNFTRACE(rule.IsTraceOn()).
							Counter()
						if rule.IsLogOn() {
							rb = rb.DLogs(nfte.LogFlagsIPOpt)
						}
						rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
						return nil
					})
				}
			}

			if isIN {
				// INGRESS → ICMP transports from svc.
				for _, transport := range localSvc.Spec.Transports {
					if _, isIcmp := transport.(domain.IcmpTransport); !isIcmp {
						continue
					}
					emitIcmp(transport, false)
				}
				continue
			}
			// EGRESS → ICMP transport from rule.
			if _, isIcmp := rule.Spec.Transport.(domain.IcmpTransport); !isIcmp {
				continue
			}
			emitIcmp(rule.Spec.Transport, true)
		}
	}
	return nil
}

func (gp *jobGroup) populateBothtSvcSvcRules(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-%s-svc-rule", misc.Tern(isIN, "in", "out"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()

	bt := gp.bt
	targetSGchName := bt.inOutChainName(dir, ag)

	svcRules := bt.data.Rules.At(domain.Svc2SvcRule).BothTraffic()
	if svcRules == nil {
		return nil
	}

	for _, svcID := range ag.GetServiceRefs() {
		localSvc, ok := bt.data.Services.Get(svcID)
		if !ok {
			continue
		}
		rules, err := misc.Tern(isIN,
			svcRules.In, svcRules.Out,
		)(svcID)
		if err != nil {
			return errors.WithMessagef(err, "invalid %s rules for local ag '%s' and svc '%s'",
				misc.Tern(isIN, "in", "out"), ag.Metadata.ID.NamespacedName(), svcID.String())
		}
		for _, rule := range rules {
			rlName := rule.Metadata.ID.NamespacedName()
			pri, err := rule.Priority()
			if err != nil {
				return errors.WithMessagef(err, "invalid priority for rule '%s'", rlName)
			}
			from, err := domain.LocalFromSpec(rule.Spec.Local)
			if err != nil {
				return errors.WithMessagef(err, "invalid local for rule '%s'", rlName)
			}
			to, err := domain.RemoteFromSpec(rule.Spec.Remote)
			if err != nil {
				return errors.WithMessagef(err, "invalid remote for rule '%s'", rlName)
			}
			remoteSvc, ok := bt.data.Services.Get(
				misc.Tern(isIN, from.ResourceIdentifier, to.ResourceIdentifier),
			)
			if !ok {
				continue
			}
			svc := localSvc
			for _, transport := range svc.Spec.Transports {
				ipV := transport.GetIPv()
				addrSetName := nameUtils{}.nameOfSvcNetSet(int(ipV), remoteSvc.ResourceID().String())

				switch transport.(type) {
				case domain.L4Transport:
					l4proto := transport.GetProto().L4Proto()
					proto, ok := l4proto.Maybe()
					if !ok {
						return errors.Errorf("invalid transport protocol for rule '%s'", rlName)
					}
					ports := setsUtils{}.makeAccPorts(domain.PortFromTransport(transport))
					if len(ports) == 0 {
						ports = append(ports, accports{})
					}
					for i := range ports { //nolint:dupl
						ports := ports[i]
						gp.addJob(pri, api, func(tx *Tx) error {
							chnApplyTo := bt.chains.At(targetSGchName)
							addrSet := bt.addrsets.At(addrSetName)
							if chnApplyTo != nil && addrSet != nil {
								bt.log.Debugf("add %s-svc-rule for addr-set '%s' into '%s'/'%s'",
									misc.Tern(isIN, "in", "out"),
									addrSetName, bt.table.Name, targetSGchName)
								r := BeginRule()
								if isIN {
									r = ports.D(
										r.SAddr(int(ipV)).InSet(addrSet).
											ProtoIP(proto),
									)
								} else {
									r = ports.D(
										r.DAddr(int(ipV)).InSet(addrSet).
											ProtoIP(proto),
									)
								}
								r = r.MetaNFTRACE(rule.IsTraceOn()).Counter()
								if rule.IsLogOn() {
									r = r.DLogs(nfte.LogFlagsIPOpt)
								}
								r.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
							}
							return nil
						})
					}
				case domain.IcmpTransport:
					for _, icmp := range domain.IcmpFromTransport(transport) {
						gp.addJob(pri, api, func(tx *Tx) error {
							chnApplyTo := bt.chains.At(targetSGchName)
							addrSet := bt.addrsets.At(addrSetName)
							if chnApplyTo != nil && addrSet != nil {
								bt.log.Debugf("add %s-svc-icmp%v-rule for addr-set '%s' into '%s'/'%s' with priority(%v)",
									misc.Tern(isIN, "in", "out"),
									misc.Tern(icmp.IPv == netrc.IPv6, "6", ""),
									addrSetName, targetSGchName, bt.table.Name, pri)
								rb := BeginRule()
								rb = misc.Tern(isIN, rb.SAddr, rb.DAddr)(int(icmp.IPv)).
									InSet(addrSet).
									ProtoICMP(icmp).
									MetaNFTRACE(rule.IsTraceOn()).
									Counter()
								if rule.IsLogOn() {
									rb = rb.DLogs(nfte.LogFlagsIPOpt)
								}
								rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
							}
							return nil
						})
					}
				}
			}
		}
	}

	return nil
}

func (gp *jobGroup) populateBothAgAgIcmpRules(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-%s-ag-icmp-rule", misc.Tern(isIN, "in", "out"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	bt := gp.bt
	targetChName := bt.inOutChainName(dir, ag)
	allRules := bt.data.Rules.At(domain.Ag2AgIcmpRule).BothTraffic()
	if allRules == nil {
		return nil
	}
	rules, err := misc.Tern(isIN,
		allRules.In, allRules.Out,
	)(ag.Metadata.ID.ResourceID())
	if err != nil {
		return errors.WithMessagef(err, "invalid %s rules for local ag '%s'",
			misc.Tern(isIN, "in", "out"),
			ag.Metadata.ID.NamespacedName(),
		)
	}
	for _, rule := range rules {
		rlName := rule.Metadata.ID.NamespacedName()
		pri, err := rule.Priority()
		if err != nil {
			return errors.WithMessagef(err, "invalid priority for rule '%s'", rlName)
		}
		local, err := domain.LocalFromSpec(rule.Spec.Local)
		if err != nil {
			return errors.WithMessagef(err, "invalid local for rule '%s'", rlName)
		}
		remote, err := domain.RemoteFromSpec(rule.Spec.Remote)
		if err != nil {
			return errors.WithMessagef(err, "invalid remote for rule '%s'", rlName)
		}
		for _, icmp := range domain.IcmpFromTransport(rule.Spec.Transport) {
			gp.addJob(pri, api, func(tx *Tx) error {
				addrSetName := nameUtils{}.nameOfNetSet(
					int(icmp.IPv),
					misc.Tern(isIN, local.ResourceIdentifier.String(), remote.ResourceIdentifier.String()),
				)
				chnApplyTo := bt.chains.At(targetChName)
				addrSet := bt.addrsets.At(addrSetName)
				if addrSet != nil && chnApplyTo != nil {
					bt.log.Debugf("add %s-ag-icmp%v-rule for addr-set '%s' into '%s'/'%s' with priority(%v)",
						misc.Tern(isIN, "in", "out"),
						misc.Tern(icmp.IPv == netrc.IPv6, "6", ""),
						addrSetName, targetChName, bt.table.Name, pri)
					rb := BeginRule()
					rb = misc.Tern(isIN, rb.SAddr, rb.DAddr)(int(icmp.IPv)).
						InSet(addrSet).
						ProtoICMP(icmp).
						MetaNFTRACE(rule.IsTraceOn()).
						Counter()
					if rule.IsLogOn() {
						rb = rb.DLogs(nfte.LogFlagsIPOpt)
					}
					rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
				}
				return nil
			})
		}
	}
	return nil
}

func (gp *jobGroup) populateBothAgAgRules(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-%s-ag-rule", misc.Tern(isIN, "in", "out"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	bt := gp.bt
	targetSGchName := bt.inOutChainName(dir, ag)

	allRules := bt.data.Rules.At(domain.Ag2AgRule).BothTraffic()
	if allRules == nil {
		return nil
	}

	rules, err := misc.Tern(isIN,
		allRules.In, allRules.Out,
	)(ag.Metadata.ID.ResourceID())
	if err != nil {
		return errors.WithMessagef(err, "invalid %s rules for local ag '%s'",
			misc.Tern(isIN, "in", "out"), ag.Metadata.ID.NamespacedName())
	}

	for _, rule := range rules {
		rlName := rule.Metadata.ID.NamespacedName()
		pri, err := rule.Priority()
		if err != nil {
			return errors.WithMessagef(err, "invalid priority for rule '%s'", rlName)
		}
		local, err := domain.LocalFromSpec(rule.Spec.Local)
		if err != nil {
			return errors.WithMessagef(err, "invalid local for rule '%s'", rlName)
		}
		remote, err := domain.RemoteFromSpec(rule.Spec.Remote)
		if err != nil {
			return errors.WithMessagef(err, "invalid remote for rule '%s'", rlName)
		}
		ipV := rule.Spec.Transport.GetIPv()
		addrSetName := nameUtils{}.nameOfNetSet(int(ipV),
			misc.Tern(isIN, local.NamespacedName(), remote.NamespacedName()))

		l4proto := rule.Spec.Transport.GetProto().L4Proto()
		proto, ok := l4proto.Maybe()
		if !ok {
			return errors.Errorf("invalid transport protocol for rule '%s'", rlName)
		}

		accports := setsUtils{}.makeAccPorts(domain.PortFromTransport(rule.Spec.Transport))
		for _, ports := range accports {
			gp.addJob(pri, api, func(tx *Tx) error {
				chnApplyTo := bt.chains.At(targetSGchName)
				addrSet := bt.addrsets.At(addrSetName)

				if chnApplyTo != nil && addrSet != nil {
					bt.log.Debugf("add %s-ag-rule for addr-set '%s' into '%s'/'%s'",
						misc.Tern(isIN, "in", "out"),
						addrSetName, bt.table.Name, targetSGchName)

					r := BeginRule()
					if isIN {
						r = ports.D(
							r.SAddr(int(ipV)).InSet(addrSet).
								ProtoIP(proto),
						)
					} else {
						r = ports.D(
							r.DAddr(int(ipV)).InSet(addrSet).
								ProtoIP(proto),
						)
					}

					r = r.MetaNFTRACE(rule.IsTraceOn()).Counter()
					if rule.IsLogOn() {
						r = r.DLogs(nfte.LogFlagsIPOpt)
					}
					r.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
				}
				return nil
			})
		}
	}
	return nil
}

func (gp *jobGroup) populateIESvcCidrIcmpRules(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-ie-svc-cidr-icmp-%sgress-rule", misc.Tern(isIN, "in", "e"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	bt := gp.bt
	targetSGchName := bt.inOutChainName(dir, ag)

	ruleType := domain.Svc2CidrIcmpRule
	svcRules := bt.data.Rules.At(ruleType).Traffic(misc.Tern(isIN, domain.INGRESS, domain.EGRESS))
	if svcRules == nil {
		return nil
	}
	for _, svcRef := range ag.GetServiceRefs() {
		svc, ok := bt.data.Services.Get(svcRef)
		if !ok {
			continue
		}
		for _, ref := range svc.GetRefsByTypes(ruleType) {
			rule, ok := svcRules.Get(ref)
			if !ok {
				continue
			}
			pri, err := rule.Priority()
			if err != nil {
				return errors.WithMessagef(err, "invalid priority for rule '%s'", ref)
			}
			cidr, err := domain.CidrFromSpec(rule.Spec.Remote)
			if err != nil {
				return errors.WithMessagef(err, "invalid CIDR for rule '%s'", ref)
			}

			applyRule := func(transport domain.TransportSpec) {
				for _, icmp := range domain.IcmpFromTransport(transport) {
					gp.addJob(pri, api, func(tx *Tx) error {
						chnApplyTo := bt.chains.At(targetSGchName)
						if chnApplyTo == nil {
							return nil
						}
						bt.log.Debugf("add svc(%s)-cidr(%s)-icmp-%sgress-rule into '%s'/'%s' with priority(%v)",
							svc.Metadata.ID.NamespacedName(),
							cidr.Value,
							misc.Tern(isIN, "in", "e"),
							bt.table.Name, targetSGchName,
							pri,
						)
						rb := BeginRule().
							SrcOrDstSingleIpNet(cidr.Value.IPNet, isIN).
							ProtoICMP(icmp).
							MetaNFTRACE(rule.IsTraceOn()).
							Counter()
						if rule.IsLogOn() {
							rb = rb.DLogs(nfte.LogFlagsIPOpt)
						}
						rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
						return nil
					})
				}
			}
			if !isIN {
				applyRule(rule.Spec.Transport)
			} else {
				for _, transport := range svc.Spec.Transports {
					applyRule(transport)
				}
			}
		}
	}

	return nil
}

func (gp *jobGroup) populateIESvcSvcRules(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-svc-%sgress-svc-rule(s)", misc.Tern(isIN, "in", "e"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()

	bt := gp.bt
	targetSGchName := bt.inOutChainName(dir, ag)

	ruleType := domain.Svc2SvcRule
	svcRules := bt.data.Rules.At(ruleType).Traffic(misc.Tern(isIN, domain.INGRESS, domain.EGRESS))
	if svcRules == nil {
		return nil
	}

	for _, svcID := range ag.GetServiceRefs() {
		localSvc, ok := bt.data.Services.Get(svcID)
		if !ok {
			continue
		}
		for _, rulRef := range localSvc.GetRefsByTypes(ruleType) {
			rule, ok := svcRules.Get(rulRef)
			if !ok {
				continue
			}
			localRef, err := domain.LocalFromSpec(rule.Spec.Local)
			if err != nil {
				return errors.WithMessagef(err, "invalid local for rule '%s'", rulRef)
			}
			if localRef.ResourceIdentifier != svcID {
				continue
			}
			remoteSvcRef, err := domain.RemoteFromSpec(rule.Spec.Remote)
			if err != nil {
				return errors.WithMessagef(err, "invalid remote for rule '%s'", rulRef)
			}
			remoteSvc, ok := bt.data.Services.Get(remoteSvcRef.ResourceIdentifier)
			if !ok {
				continue
			}
			pri, err := rule.Priority()
			if err != nil {
				return errors.WithMessagef(err, "invalid priority for rule '%s'", rulRef)
			}
			for _, transport := range misc.Tern(isIN, localSvc, remoteSvc).Spec.Transports {
				ipV := transport.GetIPv()
				addrSetName := nameUtils{}.nameOfSvcNetSet(int(ipV), remoteSvc.ResourceID().String())
				switch transport.(type) {
				case domain.L4Transport:
					l4proto := transport.GetProto().L4Proto()
					proto, ok := l4proto.Maybe()
					if !ok {
						return errors.Errorf("invalid transport protocol for rule '%s'", rulRef)
					}
					ports := setsUtils{}.makeAccPorts(domain.PortFromTransport(transport))
					if len(ports) == 0 {
						ports = append(ports, accports{})
					}
					for i := range ports { //nolint:dupl
						ports := ports[i]
						gp.addJob(pri, api, func(tx *Tx) error {
							chnApplyTo := bt.chains.At(targetSGchName)
							addrSet := bt.addrsets.At(addrSetName)
							if chnApplyTo == nil || addrSet == nil {
								return nil
							}
							bt.log.Debugf("add '%s' rule for accports(%s) into '%s'/'%s' with priority(%v)",
								rule.Metadata.ID.NamespacedName(), ports, bt.table.Name, targetSGchName, pri)

							rb := BeginRule()
							rb = ports.D(
								misc.Tern(isIN, rb.SAddr, rb.DAddr)(int(ipV)).InSet(addrSet).
									ProtoIP(proto),
							).MetaNFTRACE(rule.IsTraceOn()).
								Counter()
							if rule.IsLogOn() {
								rb = rb.DLogs(nfte.LogFlagsIPOpt)
							}
							rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
							return nil
						})
					}
				case domain.IcmpTransport:
					for _, icmp := range domain.IcmpFromTransport(transport) {
						gp.addJob(pri, api, func(tx *Tx) error {
							chnApplyTo := bt.chains.At(targetSGchName)
							addrSet := bt.addrsets.At(addrSetName)
							if chnApplyTo != nil && addrSet != nil {
								bt.log.Debugf("add %s-svc-icmp%v-rule for addr-set '%s' into '%s'/'%s' with priority(%v)",
									misc.Tern(isIN, "in", "out"),
									misc.Tern(icmp.IPv == netrc.IPv6, "6", ""),
									addrSetName, targetSGchName, bt.table.Name, pri)
								rb := BeginRule()
								rb = misc.Tern(isIN, rb.SAddr, rb.DAddr)(int(icmp.IPv)).
									InSet(addrSet).
									ProtoICMP(icmp).
									MetaNFTRACE(rule.IsTraceOn()).
									Counter()
								if rule.IsLogOn() {
									rb = rb.DLogs(nfte.LogFlagsIPOpt)
								}
								rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
							}
							return nil
						})
					}
				}
			}
		}
	}

	return nil
}

func (gp *jobGroup) populateIESvcCidrRules(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-ie-svc-cidr-%sgress-rule", misc.Tern(isIN, "in", "e"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	bt := gp.bt
	targetSGchName := bt.inOutChainName(dir, ag)

	ruleType := domain.Svc2CidrRule
	svcRules := bt.data.Rules.At(ruleType).Traffic(misc.Tern(isIN, domain.INGRESS, domain.EGRESS))
	if svcRules == nil {
		return nil
	}
	for _, svcRef := range ag.GetServiceRefs() {
		svc, ok := bt.data.Services.Get(svcRef)
		if !ok {
			continue
		}
		for _, ref := range svc.GetRefsByTypes(ruleType) {
			rule, ok := svcRules.Get(ref)
			if !ok {
				continue
			}
			pri, err := rule.Priority()
			if err != nil {
				return errors.WithMessagef(err, "invalid priority for rule '%s'", ref)
			}
			cidr, err := domain.CidrFromSpec(rule.Spec.Remote)
			if err != nil {
				return errors.WithMessagef(err, "invalid CIDR for rule '%s'", ref)
			}

			applyRule := func(transport domain.TransportSpec) error {
				l4proto := transport.GetProto().L4Proto()
				proto, ok := l4proto.Maybe()
				if !ok {
					return errors.Errorf("invalid transport protocol for rule '%s'", ref)
				}
				accports := setsUtils{}.makeAccPorts(domain.PortFromTransport(transport))
				for _, ports := range accports {
					gp.addJob(pri, api, func(tx *Tx) error {
						bt.log.Debugf("add svc-cidr rule '%s' into '%s'/'%s' with priority(%v)",
							rule.Metadata.ID.NamespacedName(), bt.table.Name, targetSGchName, pri)
						chnApplyTo := bt.chains.At(targetSGchName)
						if chnApplyTo == nil {
							return nil
						}
						rb := BeginRule().
							SrcOrDstSingleIpNet(cidr.Value.IPNet, isIN).
							ProtoIP(proto)
						rb = ports.D(rb).
							MetaNFTRACE(rule.IsTraceOn()).
							Counter()
						if rule.IsLogOn() {
							rb = rb.DLogs(nfte.LogFlagsIPOpt)
						}
						rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
						return nil
					})
				}
				return nil
			}
			if !isIN {
				if err = applyRule(rule.Spec.Transport); err != nil {
					return err
				}
			} else {
				for _, transport := range svc.Spec.Transports {
					if err = applyRule(transport); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (gp *jobGroup) populateIEAgAgIcmpRules(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-ag%s-%sgress-ag%s-icmp-rule(s)",
		misc.Tern(isIN, "", "Local"),
		misc.Tern(isIN, "in", "e"),
		misc.Tern(isIN, "Local", ""),
	)
	defer func() {
		err = errors.WithMessage(err, api)
	}()

	bt := gp.bt
	targetSGchName := bt.inOutChainName(dir, ag)

	ruleType := domain.Ag2AgIcmpRule

	rules := bt.data.Rules.At(ruleType).Traffic(misc.Tern(isIN, domain.INGRESS, domain.EGRESS))
	if rules == nil {
		return nil
	}

	for _, ref := range ag.GetRefsByTypes(ruleType) {
		rule, ok := rules.Get(ref)
		if !ok {
			continue
		}

		local, err := domain.LocalFromSpec(rule.Spec.Local)
		if err != nil {
			return errors.WithMessagef(err, "invalid local for rule '%s'", ref)
		}
		if local.ResourceIdentifier != ag.Metadata.ID.ResourceID() {
			continue
		}
		pri, err := rule.Priority()
		if err != nil {
			return errors.WithMessagef(err, "invalid priority for rule '%s'", ref)
		}
		remote, err := domain.RemoteFromSpec(rule.Spec.Remote)
		if err != nil {
			return errors.WithMessagef(err, "invalid remote for rule '%s'", ref)
		}
		ipV := rule.Spec.Transport.GetIPv()
		addrSetName := nameUtils{}.nameOfNetSet(ipVersion(ipV), remote.NamespacedName())
		for _, icmp := range domain.IcmpFromTransport(rule.Spec.Transport) {
			gp.addJob(pri, api, func(tx *Tx) error {
				chnApplyTo := bt.chains.At(targetSGchName)
				addrSet := bt.addrsets.At(addrSetName)
				if chnApplyTo != nil && addrSet != nil {
					bt.log.Debugf("add %s(%s)-%sgress-%s(%s)-icmp%v rule for addr-set '%s' into '%s'/'%s' with priority(%v)",
						misc.Tern(isIN, "ag", "agLocal"), remote.NamespacedName(),
						misc.Tern(isIN, "in", "e"),
						misc.Tern(isIN, "agLocal", "ag"), local.ResourceIdentifier.String(),
						misc.Tern(icmp.IPv == netrc.IPv6, "6", ""),
						addrSetName, targetSGchName, bt.table.Name,
						pri)
					rb := BeginRule()
					rb = misc.Tern(isIN, rb.SAddr, rb.DAddr)(int(icmp.IPv)).
						InSet(addrSet).
						ProtoICMP(icmp).
						MetaNFTRACE(rule.IsTraceOn()).
						Counter()
					if rule.IsLogOn() {
						rb = rb.DLogs(nfte.LogFlagsIPOpt)
					}
					rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
				}
				return nil
			})
		}
	}
	return nil
}

func (gp *jobGroup) populateIEAgAgRules(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-ag-%sgress-ag-rule(s)", misc.Tern(isIN, "in", "e"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	bt := gp.bt
	targetSGchName := bt.inOutChainName(dir, ag)

	ruleType := domain.Ag2AgRule

	rules := bt.data.Rules.At(ruleType).Traffic(misc.Tern(isIN, domain.INGRESS, domain.EGRESS))
	if rules == nil {
		return nil
	}
	for _, ref := range ag.GetRefsByTypes(ruleType) {
		rule, ok := rules.Get(ref)
		if !ok {
			continue
		}

		local, err := domain.LocalFromSpec(rule.Spec.Local)
		if err != nil {
			return errors.WithMessagef(err, "invalid local for rule '%s'", ref)
		}
		if local.ResourceIdentifier != ag.Metadata.ID.ResourceID() {
			continue
		}

		remote, err := domain.RemoteFromSpec(rule.Spec.Remote)
		if err != nil {
			return errors.WithMessagef(err, "invalid remote for rule '%s'", ref)
		}

		pri, err := rule.Priority()
		if err != nil {
			return errors.WithMessagef(err, "invalid priority for rule '%s'", ref)
		}
		ipV := rule.Spec.Transport.GetIPv()
		addrSetName := nameUtils{}.nameOfNetSet(ipVersion(ipV), remote.NamespacedName())
		l4proto := rule.Spec.Transport.GetProto().L4Proto()
		proto, ok := l4proto.Maybe()
		if !ok {
			return errors.Errorf("invalid transport protocol for rule '%s'", ref)
		}

		accports := setsUtils{}.makeAccPorts(domain.PortFromTransport(rule.Spec.Transport))
		for _, ports := range accports {
			gp.addJob(pri, api, func(tx *Tx) error {
				chnApplyTo := bt.chains.At(targetSGchName)
				addrSet := bt.addrsets.At(addrSetName)
				if chnApplyTo == nil || addrSet == nil {
					return nil
				}
				bt.log.Debugf("add '%s' rule for accports(%s) into '%s'/'%s' with priority(%v)",
					rule.Metadata.ID.NamespacedName(), ports, bt.table.Name, targetSGchName, pri)

				rb := BeginRule()
				rb = ports.D(
					misc.Tern(isIN, rb.SAddr, rb.DAddr)(int(ipV)).InSet(addrSet).
						ProtoIP(proto),
				).MetaNFTRACE(rule.IsTraceOn()).
					Counter()
				if rule.IsLogOn() {
					rb = rb.DLogs(nfte.LogFlagsIPOpt)
				}
				rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
				return nil
			})
		}
	}
	return nil
}

func (gp *jobGroup) populateIEAgIcmpRules(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-ie-ag-icmp-%s-rule", misc.Tern(isIN, "in", "e"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()

	bt := gp.bt
	targetChName := bt.inOutChainName(dir, ag)
	ruleType := domain.Ag2IcmpRule
	rules := bt.data.Rules.At(ruleType).Traffic(misc.Tern(isIN, domain.INGRESS, domain.EGRESS))
	if rules == nil {
		return nil
	}
	for _, ref := range ag.GetRefsByTypes(ruleType) {
		rule, ok := rules.Get(ref)
		if !ok {
			continue
		}
		pri, err := rule.Priority()
		if err != nil {
			return errors.WithMessagef(err, "invalid priority for rule '%s'", ref)
		}
		for _, icmp := range domain.IcmpFromTransport(rule.Spec.Transport) {
			gp.addJob(pri, api, func(tx *Tx) error {
				chnApplyTo := bt.chains.At(targetChName)
				if chnApplyTo != nil {
					bt.log.Debugf("add ag-icmp%v-rule into '%s'/'%s'",
						misc.Tern(icmp.IPv == netrc.IPv6, "6", ""),
						bt.table.Name, targetChName)
					rb := BeginRule().
						ProtoICMP(icmp).
						MetaNFTRACE(rule.IsTraceOn()).
						Counter()
					if rule.IsLogOn() {
						rb = rb.DLogs(nfte.LogFlagsIPOpt)
					}
					rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
				}
				return nil
			})
		}
	}
	return nil
}

func (gp *jobGroup) populateOutSvcFqdnRules(dir direction, ag domain.AddressGroup) (err error) {
	api := "populate-svc-fqdn-rule"
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	if dir != dirOUT {
		return nil
	}
	const (
		useDNS = 1 << iota
		//useNDPI
	)
	var strategy int
	bt := gp.bt
	if bt.fqdnStrategy.Eq(config.FqdnRulesStartegyDNS) {
		strategy = useDNS
	}
	targetChName := bt.inOutChainName(dirOUT, ag)

	ruleType := domain.Svc2FqdnRule
	rules := bt.data.Rules.At(ruleType)
	if rules == nil {
		return nil
	}

	for _, svcRef := range ag.GetServiceRefs() {
		svc, ok := bt.data.Services.Get(svcRef)
		if !ok {
			continue
		}
		for _, ruleRef := range svc.GetRefsByTypes(ruleType) { //nolint:dupl
			rule, ok := rules.Get(ruleRef)
			if !ok {
				continue
			}
			fqdn, err := domain.FqdnFromSpec(rule.Spec.Remote)
			if err != nil {
				return errors.WithMessagef(err, "invalid FQDN for rule '%s'", ruleRef)
			}
			var pri int16
			pri, err = rule.Priority()
			if err != nil {
				return errors.WithMessagef(err, "invalid priority for rule '%s'", ruleRef)
			}
			ipV := rule.Spec.Transport.GetIPv()
			daddrSetName := nameUtils{}.nameOfFqdnNetSet(ipVersion(ipV), fqdn.Value)
			l4proto := rule.Spec.Transport.GetProto().L4Proto()
			proto, ok := l4proto.Maybe()
			if !ok {
				return errors.Errorf("invalid transport protocol for rule '%s'", ruleRef)
			}
			accports := setsUtils{}.makeAccPorts(domain.PortFromTransport(rule.Spec.Transport))
			for _, ports := range accports {
				gp.addJob(pri, api, func(tx *Tx) error {
					chnApplyTo := bt.chains.At(targetChName)
					if chnApplyTo == nil {
						return nil
					}
					daddr := bt.addrsets.At(daddrSetName)
					if daddr == nil && strategy&useDNS != 0 {
						return nil
					}
					if daddr != nil {
						bt.log.Debugf("add svc-fqdn rule '%s' with '%s' strategy into '%s'/'%s' for addr-set '%s' with priority(%v)",
							rule.Metadata.ID.NamespacedName(), string(bt.fqdnStrategy), bt.table.Name, targetChName, daddrSetName, pri)
					} else {
						bt.log.Debugf("add svc-fqdn rule '%s' with '%s' strategy into '%s'/'%s' with priority(%v)",
							rule.Metadata.ID.NamespacedName(), string(bt.fqdnStrategy), bt.table.Name, targetChName, pri)
					}
					r := BeginRule()
					if strategy&useDNS != 0 {
						r = r.DAddr(int(ipV)).InSet(daddr)
					}

					r = ports.D(
						r.ProtoIP(proto),
					)
					r = r.MetaNFTRACE(rule.IsTraceOn()).Counter()
					if rule.IsLogOn() {
						r = r.DLogs(nfte.LogFlagsIPOpt)
					}
					r.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
					return nil
				})
			}
			if strategy&useDNS == 0 {
				break
			}
		}
	}
	return nil
}

func (gp *jobGroup) populateOutAgFqdnRules(dir direction, ag domain.AddressGroup) (err error) {
	api := "populate-ag-fqdn-rule"
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	if dir != dirOUT {
		return nil
	}
	const (
		useDNS = 1 << iota
		//useNDPI
	)
	var strategy int
	bt := gp.bt

	if bt.fqdnStrategy.Eq(config.FqdnRulesStartegyDNS) {
		strategy = useDNS
	}
	targetChName := bt.inOutChainName(dirOUT, ag)

	ruleType := domain.Ag2FqdnRule
	rules := bt.data.Rules.At(ruleType)
	if rules == nil {
		return nil
	}
	for _, ref := range ag.GetRefsByTypes(ruleType) { //nolint:dupl
		rule, ok := rules.Get(ref)
		if !ok {
			continue
		}

		fqdn, err := domain.FqdnFromSpec(rule.Spec.Remote)
		if err != nil {
			return errors.WithMessagef(err, "invalid FQDN for rule '%s'", ref)
		}
		var pri int16
		pri, err = rule.Priority()
		if err != nil {
			return errors.WithMessagef(err, "invalid priority for rule '%s'", ref)
		}
		ipV := rule.Spec.Transport.GetIPv()
		daddrSetName := nameUtils{}.nameOfFqdnNetSet(ipVersion(ipV), fqdn.Value)
		l4proto := rule.Spec.Transport.GetProto().L4Proto()
		proto, ok := l4proto.Maybe()
		if !ok {
			return errors.Errorf("invalid transport protocol for rule '%s'", ref)
		}
		accports := setsUtils{}.makeAccPorts(domain.PortFromTransport(rule.Spec.Transport))
		for _, ports := range accports {
			gp.addJob(pri, api, func(tx *Tx) error {
				chnApplyTo := bt.chains.At(targetChName)
				if chnApplyTo == nil {
					return nil
				}
				daddr := bt.addrsets.At(daddrSetName)
				if daddr == nil && strategy&useDNS != 0 {
					return nil
				}
				if daddr != nil {
					bt.log.Debugf("add ag-fqdn rule '%s' with '%s' strategy into '%s'/'%s' for addr-set '%s' with priority(%v)",
						rule.Metadata.ID.NamespacedName(), string(bt.fqdnStrategy), bt.table.Name, targetChName, daddrSetName, pri)
				} else {
					bt.log.Debugf("add ag-fqdn rule '%s' with '%s' strategy into '%s'/'%s' with priority(%v)",
						rule.Metadata.ID.NamespacedName(), string(bt.fqdnStrategy), bt.table.Name, targetChName, pri)
				}
				r := BeginRule()
				if strategy&useDNS != 0 {
					r = r.DAddr(int(ipV)).InSet(daddr)
				}

				r = ports.D(
					r.ProtoIP(proto),
				)
				r = r.MetaNFTRACE(rule.IsTraceOn()).Counter()
				if rule.IsLogOn() {
					r = r.DLogs(nfte.LogFlagsIPOpt)
				}
				r.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
				return nil
			})
		}
		if strategy&useDNS == 0 {
			break
		}
	}
	return nil
}

func (gp *jobGroup) populateIEAgCidrIcmpRules(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-ie-ag-cidr-icmp-%sgress--rule(s)",
		misc.Tern(isIN, "in", "e"),
	)
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	bt := gp.bt

	targetSGchName := bt.inOutChainName(dir, ag)

	ruleType := domain.Ag2CidrIcmpRule

	rules := bt.data.Rules.At(ruleType).Traffic(misc.Tern(isIN, domain.INGRESS, domain.EGRESS))
	if rules == nil {
		return nil
	}
	for _, ref := range ag.GetRefsByTypes(ruleType) {
		rule, ok := rules.Get(ref)
		if !ok {
			continue
		}
		pri, err := rule.Priority()
		if err != nil {
			return errors.WithMessagef(err, "invalid priority for rule '%s'", ref)
		}
		ipV := rule.Spec.Transport.GetIPv()
		addrSetName := nameUtils{}.nameOfNetSet(ipVersion(ipV), ag.Metadata.ID.NamespacedName())
		cidr, err := domain.CidrFromSpec(rule.Spec.Remote)
		if err != nil {
			return errors.WithMessagef(err, "invalid CIDR for rule '%s'", ref)
		}
		for _, icmp := range domain.IcmpFromTransport(rule.Spec.Transport) {
			gp.addJob(pri, api, func(tx *Tx) error {
				chnApplyTo := bt.chains.At(targetSGchName)
				addrSet := bt.addrsets.At(addrSetName)
				if chnApplyTo != nil && addrSet != nil {
					bt.log.Debugf("add ag(%s)-cidr(%s)-icmp-%sgress-rule for addr-set '%s' into '%s'/'%s' with priority(%v)",
						ag.Metadata.ID.NamespacedName(),
						cidr.Value,
						misc.Tern(isIN, "in", "e"),
						addrSetName,
						bt.table.Name, targetSGchName,
						pri,
					)
				}
				rb := BeginRule().
					SrcOrDstSingleIpNet(cidr.Value.IPNet, isIN).
					ProtoICMP(icmp).
					MetaNFTRACE(rule.IsTraceOn()).
					Counter()
				if rule.IsLogOn() {
					rb = rb.DLogs(nfte.LogFlagsIPOpt)
				}
				rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
				return nil
			})
		}
	}
	return nil
}

func (gp *jobGroup) populateIEAgCidrRules(dir direction, ag domain.AddressGroup) (err error) {
	isIN := dir == dirIN
	api := fmt.Sprintf("populate-ie-ag-cidr-%sgress-rule", misc.Tern(isIN, "in", "e"))
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	bt := gp.bt

	targetSGchName := bt.inOutChainName(dir, ag)
	ruleType := domain.Ag2CidrRule

	rules := bt.data.Rules.At(ruleType).Traffic(misc.Tern(isIN, domain.INGRESS, domain.EGRESS))
	if rules == nil {
		return nil
	}
	for _, ref := range ag.GetRefsByTypes(ruleType) {
		rule, ok := rules.Get(ref)
		if !ok {
			continue
		}

		pri, err := rule.Priority()
		if err != nil {
			return errors.WithMessagef(err, "invalid priority for rule '%s'", ref)
		}
		cidr, err := domain.CidrFromSpec(rule.Spec.Remote)
		if err != nil {
			return errors.WithMessagef(err, "invalid CIDR for rule '%s'", ref)
		}
		l4proto := rule.Spec.Transport.GetProto().L4Proto()
		proto, ok := l4proto.Maybe()
		if !ok {
			return errors.Errorf("invalid transport protocol for rule '%s'", ref)
		}
		accports := setsUtils{}.makeAccPorts(domain.PortFromTransport(rule.Spec.Transport))
		for _, ports := range accports {
			gp.addJob(pri, api, func(tx *Tx) error {
				bt.log.Debugf("add ag-cidr rule '%s' into '%s'/'%s' with priority(%v)",
					rule.Metadata.ID.NamespacedName(), bt.table.Name, targetSGchName, pri)
				chnApplyTo := bt.chains.At(targetSGchName)
				if chnApplyTo == nil {
					return nil
				}
				rb := BeginRule().
					SrcOrDstSingleIpNet(cidr.Value.IPNet, isIN).
					ProtoIP(proto)
				rb = ports.D(rb).
					MetaNFTRACE(rule.IsTraceOn()).
					Counter()
				if rule.IsLogOn() {
					rb = rb.DLogs(nfte.LogFlagsIPOpt)
				}
				rb.RuleAction2Verdict(rule.Spec.Action.ToRuleAction()).ApplyRule(chnApplyTo, tx.Conn)
				return nil
			})
		}
	}

	return nil
}

func (bt *batch) chainInOutProlog(dir direction, ag domain.AddressGroup) {
	chName := bt.inOutChainName(dir, ag)
	isIN := dir == dirIN
	api := fmt.Sprintf("%s-chain-prolog", misc.Tern(isIN, "in", "out"))
	for _, ipV := range misc.Sli(domain.IPv4, domain.IPv6) {
		destChainName := misc.Tern(dir == dirIN, chnIngressMain, chnEgressMain)
		bt.addJob(api, func(tx *Tx) error {
			addrSetName := bt.netSetName(ipV, ag)
			if addrSet := bt.addrsets.At(addrSetName); addrSet != nil {
				if bt.chains.At(chName) == nil {
					chn := tx.AddChain(&nftlib.Chain{
						Name:  chName,
						Table: bt.table,
					})
					bt.chains.Put(chName, chn)
					bt.log.Debugf("chain '%s'/'%s' is in progress",
						bt.table.Name, chName)
				}
				bt.log.Debugf("add jump-rule '%s'/('%s' -> '%s')",
					bt.table.Name, destChainName, chName)
				destChain := bt.chains.At(destChainName)
				rb := BeginRule()
				misc.Tern(isIN, rb.DAddr, rb.SAddr)(int(ipV)).
					InSet(addrSet).
					Counter().
					Jump(chName).
					ApplyRule(destChain, tx.Conn)
			}
			return nil
		})
	}
}

func (bt *batch) withGroup(fg func(g *jobGroup) error) (err error) {
	g := jobGroup{bt: bt}
	if err = fg(&g); err != nil {
		return err
	}
	g.Iterate(func(_ int16, items []jobItem) bool {
		for _, f := range items {
			bt.addJob(f.name, f.jobf)
		}
		return true
	})
	return nil
}

func (bt *batch) addJob(n string, job jobf) {
	if bt.jobs == nil {
		bt.jobs = list.New()
	}
	bt.jobs.PushBack(jobItem{name: n, jobf: job})
}

func (g *jobGroup) addJob(pri int16, n string, job jobf) {
	items := g.At(pri)
	items = append(items, jobItem{name: n, jobf: job})
	g.Put(pri, items)
}

func (bt *batch) fwInOutAddDefaultRules() {
	for _, chName := range misc.Sli(chnIngressMain, chnEgressMain) {
		bt.addJob("add-default-rules", func(tx *Tx) error {
			bt.log.Debugf("add default rules into chain '%s'/'%s'", bt.table.Name, chName)
			BeginRule().Counter().ApplyRule(bt.chains.At(chName), tx.Conn)
			return nil
		})
	}
}

func (bt *batch) switch2NewConfig() {
	bt.addJob("enable-new-config", func(tx *Tx) error {
		if bt.table.Flags&uint32(unix.NFT_TABLE_F_DORMANT) != 0 {
			bt.table.Flags &= ^uint32(unix.NFT_TABLE_F_DORMANT)
			bt.log.Debugf("activate table '%s'", bt.table.Name)
			_ = tx.AddTable(bt.table)
		}
		return nil
	})
	bt.addJob("del-nonactual-configs", func(tx *Tx) error {
		tables, err := tx.ListTables()
		if err != nil {
			return err
		}
		var names nameUtils
		for _, t := range tables {
			if names.isLikeMainTableName(t.Name) && t.Name != bt.table.Name {
				bt.log.Debugf("delete table '%s'", t.Name)
				tx.DelTable(t)
			}
		}
		return nil
	})
}

func (bt *batch) inOutChainName(dir direction, ag domain.AddressGroup) string {
	return nameUtils{}.nameOfInOutChain(
		dir,
		fmt.Sprintf("ag-%s", ag.Metadata.ID.NamespacedName()),
	)
}

func (bt *batch) netSetName(ipV domain.IpFamily, ag domain.AddressGroup) string {
	return nameUtils{}.nameOfNetSet(int(ipV), ag.Metadata.ID.NamespacedName())
}

func (bt *batch) cleanOnFail(_ context.Context) error {
	if bt.table == nil {
		return nil
	}
	tx, err := bt.txProvider()
	if err != nil {
		return err
	}
	defer tx.Close() //nolint
	var tabs []*nftlib.Table
	if tabs, err = tx.ListTables(); err != nil {
		return err
	}
	for _, t := range tabs {
		if t.Name == bt.table.Name && t.Family == bt.table.Family {
			tx.DelTable(t)
			err = tx.Flush()
			break
		}
	}
	return err
}
