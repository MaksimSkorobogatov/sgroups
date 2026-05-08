package job

import (
	"time"

	"github.com/PRO-Robotech/sgroups/internal/sg-agent/dns"
	"github.com/PRO-Robotech/sgroups/internal/sg-agent/nft"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/pkg/host"
	"github.com/H-BF/corlib/pkg/nl"
	"github.com/H-BF/corlib/pkg/patterns/observer"
)

type (
	// AppliedConfEvent -
	AppliedConfEvent struct {
		NetConf      host.NetConf
		AppliedRules nft.AppliedRules

		observer.EventType
	}

	// Ask2ResolveDomainAddressesEvent -
	Ask2ResolveDomainAddressesEvent struct {
		IpVersion   int
		FQDN        domain.FQDN
		ValidBefore time.Time

		observer.EventType
	}

	// DomainAddressesEvent -
	DomainAddressesEvent struct {
		IpVersion int
		FQDN      domain.FQDN
		DnsAnswer dns.DomainAddresses

		observer.EventType
	}

	// NetlinkUpdatesEvent -
	NetlinkUpdatesEvent struct {
		Updates []nl.WatcherMsg

		observer.EventType
	}

	// NetlinkErrorEvent -
	NetlinkErrorEvent struct {
		nl.ErrMsg

		observer.EventType
	}

	// SyncStatusErrorEvent -
	SyncStatusErrorEvent struct {
		error
		observer.EventType
	}

	// SyncStatusValueEvent -
	SyncStatusValueEvent struct {
		domain.SyncStatus
		observer.EventType
	}
)

// Cause -
func (e SyncStatusErrorEvent) Cause() error {
	return e.error
}
