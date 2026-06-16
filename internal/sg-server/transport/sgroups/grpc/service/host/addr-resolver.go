package host

import (
	"context"
	"net"
	"net/netip"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/pkg/errors"
	"google.golang.org/grpc/peer"
)

func resolveHostAddr(ctx context.Context, hosts domain.Hosts) (err error) {
	for i := range hosts {
		ep := hosts[i].Spec.Endpoints
		if ep == nil {
			continue
		}
		if !ep.Address.IsValid() {
			if ep.Address, err = peerAddr(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

func peerAddr(ctx context.Context) (ret netip.Addr, err error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ret, errors.New("no peer info in context")
	}
	tcp, ok := p.Addr.(*net.TCPAddr)
	if !ok {
		return ret, errors.Errorf("unsupported peer address type %T", p.Addr)
	}
	a, ok := netip.AddrFromSlice(tcp.IP)
	if !ok {
		return ret, errors.Errorf("invalid peer IP %q", tcp.IP)
	}
	return a.Unmap(), nil
}
