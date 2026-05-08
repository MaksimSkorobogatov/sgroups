package resources

import (
	"context"
	"sync"

	"github.com/PRO-Robotech/sgroups/internal/sg-agent/dns"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/pkg/dict"
	"github.com/H-BF/corlib/pkg/parallel"
)

type (
	// ResolvedFQDN -
	ResolvedFQDN struct {
		sync.RWMutex
		A    dict.RBDict[domain.FQDN, dns.DomainAddresses]
		AAAA dict.RBDict[domain.FQDN, dns.DomainAddresses]
	}
)

// UpdA -
func (r *ResolvedFQDN) UpdA(domain domain.FQDN, addr dns.DomainAddresses) {
	r.Lock()
	defer r.Unlock()
	r.A.Put(domain, addr)
}

// UpdAAAA -
func (r *ResolvedFQDN) UpdAAAA(domain domain.FQDN, addr dns.DomainAddresses) {
	r.Lock()
	defer r.Unlock()
	r.AAAA.Put(domain, addr)
}

// Resolve -
func (r *ResolvedFQDN) Resolve(ctx context.Context, fqdns []domain.FQDN, dnsRes dns.DomainAddressQuerier) {
	const parallelism = 7

	type item = struct {
		domain domain.FQDN
		up     func(domain.FQDN, dns.DomainAddresses)
		re     func(context.Context, string) dns.DomainAddresses
	}
	var items []item
	for _, f := range fqdns {
		items = append(items,
			item{domain: f, up: r.UpdA, re: dnsRes.A},
		)
	}
	_ = parallel.ExecAbstract(len(items), parallelism, func(i int) error {
		item := items[i]
		res := item.re(ctx, item.domain.String())
		item.up(item.domain, res)
		return nil
	})
}
