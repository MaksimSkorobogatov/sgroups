package socketscan

import (
	"net/netip"
	"slices"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/pkg/filter"
	netrc "github.com/H-BF/corlib/pkg/net/resources"
	"github.com/H-BF/corlib/pkg/option"
	"github.com/samber/lo"
)

type (
	// ScopeBySocketSelectors defines a filter scope based on socket information
	ScopeBySocketSelectors struct {
		filter.Scope
		Selectors []Selector
	}
	// Selector defines criteria for selecting socket information
	Selector struct {
		L4Proto    option.ValueOf[netrc.NetworkTransport]
		Family     option.ValueOf[domain.IpFamily]
		State      option.ValueOf[ConnState]
		LocalAddr  option.ValueOf[netip.Addr]
		LocalPort  option.ValueOf[uint16]
		RemoteAddr option.ValueOf[netip.Addr]
		RemotePort option.ValueOf[uint16]
		Ifname     string
		Inode      option.ValueOf[uint64]
		PID        option.ValueOf[int]
		Comm       string
	}
)

func (sc ScopeBySocketSelectors) match(in SocketInfo) bool {
	if len(sc.Selectors) == 0 {
		return true
	}
	return slices.ContainsFunc(sc.Selectors, func(s Selector) bool {
		if v, ok := s.L4Proto.Maybe(); ok && !v.IsEq(in.L4Proto) {
			return false
		}
		if v, ok := s.Family.Maybe(); ok && !v.IsEq(in.Family) {
			return false
		}
		if v, ok := s.State.Maybe(); ok && !v.IsEq(in.State) {
			return false
		}
		if v, ok := s.LocalAddr.Maybe(); ok && v != in.LocalAddr.Addr() {
			return false
		}
		if v, ok := s.LocalPort.Maybe(); ok && v != in.LocalAddr.Port() {
			return false
		}
		if v, ok := s.RemoteAddr.Maybe(); ok && v != in.RemoteAddr.Addr() {
			return false
		}
		if v, ok := s.RemotePort.Maybe(); ok && v != in.RemoteAddr.Port() {
			return false
		}

		if len(s.Ifname) > 0 && s.Ifname != in.Ifname {
			return false
		}
		if v, ok := s.Inode.Maybe(); ok && v != in.Inode {
			return false
		}
		if v, ok := s.PID.Maybe(); ok && !misc.IsIn(v, lo.Map(in.Process, func(p ProcessInfo, _ int) int {
			return p.PID
		})...) {
			return false
		}
		if len(s.Comm) > 0 && !misc.IsIn(s.Comm, lo.Map(in.Process, func(p ProcessInfo, _ int) string {
			return p.Comm
		})...) {
			return false
		}
		return true
	})
}
