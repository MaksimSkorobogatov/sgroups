package misc

import (
	"net"
)

// SeparateNetworks it selerates source into IPv4 and IPv4 networks
func SeparateNetworks(nws []net.IPNet, scopeIPs ...net.IP) (netIPv4, netIPv6 []net.IPNet) {
	netIPv4, netIPv6 = make([]net.IPNet, 0, len(nws)), make([]net.IPNet, 0, len(nws))

	pass := func(nw net.IPNet) bool {
		for i := range scopeIPs {
			if nw.Contains(scopeIPs[i]) {
				return true
			}
		}
		return len(scopeIPs) == 0
	}
	for _, nw := range nws {
		if pass(nw) {
			switch len(nw.IP) {
			case net.IPv6len:
				netIPv6 = append(netIPv6, nw)
			case net.IPv4len:
				netIPv4 = append(netIPv4, nw)
			}
		}
	}

	return netIPv4, netIPv6
}
