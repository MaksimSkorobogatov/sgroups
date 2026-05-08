package app

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-agent/dns"

	"github.com/H-BF/corlib/pkg/atomic"
)

// GetDnsResolver -
func GetDnsResolver() dns.DomainAddressQuerier {
	ret, ok := appDnsResolver.Load()
	if !ok {
		panic("Need call 'SetupDnsResolver'")
	}
	return ret
}

// SetupDnsResolver -
func SetupDnsResolver(ctx context.Context) error {
	dnsResolver, err := dns.NewDomainAddressQuerier(ctx)
	if err != nil {
		return err
	}
	cached := dns.NewDomainAddressQuerierCache(dnsResolver)
	appDnsResolver.Store(cached, func(o dns.DomainAddressQuerierCacheWrapper) {
		_ = o.Close()
	})
	return nil
}

var (
	appDnsResolver atomic.Value[dns.DomainAddressQuerierCacheWrapper]
)
