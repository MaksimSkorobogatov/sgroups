package nft

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/PRO-Robotech/sgroups/internal/sg-agent/resources"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/pkg/dict"
	"github.com/c-robinson/iplib"
	nftlib "github.com/google/nftables"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// RuleApplierProvider -
type RuleApplierProvider interface {
	NewApplier(ctx context.Context) (RuleApplier, error)
}

// RuleApplier -
type RuleApplier interface {
	ApplyConfig(ctx context.Context, data resources.LocalData) (AppliedRules, error)
	ApplyPatch(ctx context.Context, applied AppliedRules, patch Patch) error
	Close() error
}

type (
	// AppliedRules -
	AppliedRules struct {
		ID          uuid.UUID
		NetNS       string
		TargetTable string
		BaseRules   resources.BaseRuleList
		LocalData   resources.LocalData
	}

	// Patch -
	Patch interface {
		String() string
		Apply(context.Context, *AppliedRules) error
		isAppliedRulesPatch()
	}

	// UpdateFqdnNetsets - is kind of Patch
	UpdateFqdnNetsets struct {
		IPVersion int
		FQDN      domain.FQDN
		Addresses []net.IP
	}
)

// LastAppliedRules -
func LastAppliedRules(netNS string) *AppliedRules {
	lastAppliedRulesMx.RLock()
	defer lastAppliedRulesMx.RUnlock()
	return lastAppliedRules.At(netNS)
}

// LastAppliedRulesUpd -
func LastAppliedRulesUpd(netNS string, data *AppliedRules) {
	lastAppliedRulesMx.Lock()
	defer lastAppliedRulesMx.Unlock()
	lastAppliedRules.Put(netNS, data)
}

var (
	lastAppliedRules   dict.HDict[string, *AppliedRules]
	lastAppliedRulesMx sync.RWMutex

	_ Patch = (*UpdateFqdnNetsets)(nil)
)

func (UpdateFqdnNetsets) isAppliedRulesPatch() {}

// String impl Stringer interface
func (p UpdateFqdnNetsets) String() string {
	return fmt.Sprintf("patch/fqdn-netset(IPv: %v; domain: '%s'; addrs: %s)",
		p.IPVersion, p.FQDN, misc.Slice2Stringer(p.Addresses...))
}

// NetSet -
func (ns UpdateFqdnNetsets) NetSet() []net.IPNet {
	isV6 := ns.IPVersion == iplib.IP6Version
	bits := misc.Tern(isV6, net.IPv6len, net.IPv4len) * 8
	mask := net.CIDRMask(bits, bits)
	ret := make([]net.IPNet, len(ns.Addresses))
	for i, ip := range ns.Addresses {
		ret[i] = net.IPNet{IP: ip, Mask: mask}
	}
	return ret
}

// Apply -
func (ns UpdateFqdnNetsets) Apply(ctx context.Context, rules *AppliedRules) error {
	const api = "apply"

	if !misc.IsIn(ns.IPVersion, misc.Sli(iplib.IP4Version, iplib.IP6Version)...) {
		return errors.WithMessagef(ErrPatchNotApplicable,
			"%s/%s failed cause it has bad IPv(%v)", ns, api, ns.IPVersion)
	}
	tx, err := NewTx(rules.NetNS)
	if err != nil {
		return err
	}
	defer tx.Close() //nolint
	var nftConf NFTablesConf
	if nftConf, err = NFTconfLoad(tx.Conn); err != nil {
		return err
	}
	targetTable := NfTableKey{
		TableFamily: nftlib.TableFamilyINet,
		Name:        rules.TargetTable,
	}
	netSets := nftConf.Sets.At(targetTable)
	netsetName := nameUtils{}.
		nameOfFqdnNetSet(ns.IPVersion, ns.FQDN)
	set := netSets.At(netsetName)
	if set.Set == nil {
		return errors.WithMessagef(ErrPatchNotApplicable,
			"%s/%s failed cause targed netset '%s' does not exist", ns, api, netsetName)
	}
	elements := setsUtils{}.nets2SetElements(ns.NetSet(), ns.IPVersion)
	if err = tx.SetAddElements(set.Set, elements); err != nil {
		panic(err)
	}
	err = tx.FlushAndClose()
	return err
}
