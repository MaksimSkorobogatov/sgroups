package socketscan

import (
	"cmp"
	"net"
	"net/netip"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

type ifAddrMap map[netip.Prefix][]string

func (m ifAddrMap) lookup(ip net.IP) []string {
	a, ok := netip.AddrFromSlice(ip)
	if !ok || !a.IsValid() {
		return nil
	}
	a = a.Unmap()
	var ret []string
	for p, names := range m {
		if p.Contains(a) {
			ret = append(ret, names...)
		}
	}
	return ret
}

func (m *ifAddrMap) reload() error {
	ifs, err := net.Interfaces()
	if err != nil {
		return errors.WithMessage(err, "list interfaces")
	}
	v4 := netip.IPv4Unspecified()
	v6 := netip.IPv6Unspecified()
	out := ifAddrMap{
		netip.PrefixFrom(v4, v4.BitLen()): {"any"},
		netip.PrefixFrom(v6, v6.BitLen()): {"any"},
	}

	for i := range ifs {
		addrs, err := ifs[i].Addrs()
		if err != nil {
			return errors.WithMessagef(err, "list addrs of %s", ifs[i].Name)
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			addr, ok := netip.AddrFromSlice(ipnet.IP)
			if !ok {
				continue
			}
			ones, _ := ipnet.Mask.Size()
			p := netip.PrefixFrom(addr.Unmap(), ones)
			out[p] = append(out[p], ifs[i].Name)
		}
	}
	*m = out
	return nil
}

func toAddr(ip net.IP, port uint16, fam uint8) (ret netip.AddrPort) {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return ret
	}
	if fam == unix.AF_INET {
		addr = addr.Unmap()
	}

	return netip.AddrPortFrom(addr, port)
}

func cmpAddrPort(a, b netip.AddrPort) int {
	if c := a.Addr().Compare(b.Addr()); c != 0 {
		return c
	}
	return cmp.Compare(a.Port(), b.Port())
}

func fam2domain(fam uint8) (domain.IpFamily, error) {
	if f, ok := toFam[fam]; ok {
		return f, nil
	}
	return 0, errors.Errorf("unknown family: %d", fam)
}

var toFam = map[uint8]domain.IpFamily{
	unix.AF_INET:  domain.IPv4,
	unix.AF_INET6: domain.IPv6,
}
