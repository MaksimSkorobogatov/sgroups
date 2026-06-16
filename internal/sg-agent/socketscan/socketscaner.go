package socketscan

import (
	"cmp"
	"net/netip"
	"slices"
	"strings"
	"time"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/pkg/atomic"
	netrc "github.com/H-BF/corlib/pkg/net/resources"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"golang.org/x/sync/singleflight"
	"golang.org/x/sys/unix"
)

type (
	// SocketInfo holds information about a socket
	SocketInfo struct {
		L4Proto    netrc.NetworkTransport
		Family     domain.IpFamily
		State      ConnState
		LocalAddr  netip.AddrPort
		RemoteAddr netip.AddrPort
		Ifname     string
		Inode      uint64
		Process    []ProcessInfo
	}

	socketsSnapshot struct {
		at   time.Time
		data []SocketInfo
	}
)

var (
	snapshotHolder atomic.Value[*socketsSnapshot]
	ssGroup        singleflight.Group
)

// IsEq -
func (s SocketInfo) IsEq(other SocketInfo) bool {
	return s.L4Proto.IsEq(other.L4Proto) &&
		s.Family == other.Family &&
		s.State == other.State &&
		s.LocalAddr == other.LocalAddr &&
		s.RemoteAddr == other.RemoteAddr &&
		s.Ifname == other.Ifname &&
		s.Inode == other.Inode &&
		slices.EqualFunc(s.Process, other.Process, func(a, b ProcessInfo) bool {
			return a.IsEq(b)
		})
}

// Cmp -
func (s SocketInfo) Cmp(other SocketInfo) int {
	if c := cmp.Compare(s.L4Proto, other.L4Proto); c != 0 {
		return c
	}
	if c := cmp.Compare(s.Family, other.Family); c != 0 {
		return c
	}
	if c := cmp.Compare(s.State, other.State); c != 0 {
		return c
	}
	if c := cmpAddrPort(s.LocalAddr, other.LocalAddr); c != 0 {
		return c
	}
	if c := cmpAddrPort(s.RemoteAddr, other.RemoteAddr); c != 0 {
		return c
	}
	if c := strings.Compare(s.Ifname, other.Ifname); c != 0 {
		return c
	}
	if c := cmp.Compare(s.Inode, other.Inode); c != 0 {
		return c
	}
	for i := range s.Process {
		if c := s.Process[i].Cmp(other.Process[i]); c != 0 {
			return c
		}
	}
	return 0
}

// ScanSockets - scan sockets based on provided options and return matching SocketInfo instances.
func ScanSockets(opts ...SSopt) (ret []SocketInfo, err error) {
	var cfg ssOpts
	for _, opt := range opts {
		if err = opt.apply(&cfg); err != nil {
			return nil, err
		}
	}

	var raw []SocketInfo
	if raw, err = loadOrScan(cfg); err != nil {
		return nil, err
	}

	sc, ok := cfg.scope.(ScopeBySocketSelectors)
	if !ok {
		return raw, nil
	}
	for _, si := range raw {
		if sc.match(si) {
			ret = append(ret, si)
		}
	}
	return ret, nil
}

func loadOrScan(cfg ssOpts) ([]SocketInfo, error) {
	if cfg.cacheTTL > 0 {
		if snap, _ := snapshotHolder.Load(); snap != nil &&
			time.Since(snap.at) < cfg.cacheTTL {
			return snap.data, nil
		}
	}

	const key = "scan"
	v, err, _ := ssGroup.Do(key, func() (any, error) {
		if cfg.cacheTTL > 0 {
			if snap, _ := snapshotHolder.Load(); snap != nil &&
				time.Since(snap.at) < cfg.cacheTTL {
				return snap.data, nil
			}
		}
		data, e := scanAll()
		if e != nil {
			return nil, e
		}
		if cfg.cacheTTL > 0 {
			snapshotHolder.Store(&socketsSnapshot{at: time.Now(), data: data}, nil)
		}
		return data, nil
	})

	return misc.Tern(err == nil, v.([]SocketInfo), nil), err
}

func scanAll() ([]SocketInfo, error) {
	var ret []SocketInfo
	procs, err := makeInodeToProc()
	if err != nil {
		return nil, err
	}

	var ifByAddr ifAddrMap
	if err = ifByAddr.reload(); err != nil {
		return nil, err
	}

	for _, src := range [...]struct {
		proto  string
		family uint8
		fn     func(uint8) ([]*netlink.Socket, error)
	}{
		{"TCP", unix.AF_INET, netlink.SocketDiagTCP},
		{"TCP", unix.AF_INET6, netlink.SocketDiagTCP},
		{"UDP", unix.AF_INET, netlink.SocketDiagUDP},
		{"UDP", unix.AF_INET6, netlink.SocketDiagUDP},
	} {
		socks, e := src.fn(src.family)
		if e != nil {
			return nil, errors.WithMessagef(e, "sock_diag %s/%d", src.proto, src.family)
		}
		for _, s := range socks {
			inode := uint64(s.INode)
			procInfo, ok := procs[inode]
			if !ok {
				continue
			}
			var proto netrc.NetworkTransport
			if err = proto.FromString(src.proto); err != nil {
				return nil, errors.WithMessagef(err, "invalid protocol '%s'", src.proto)
			}

			var fam domain.IpFamily
			if fam, err = fam2domain(s.Family); err != nil {
				return nil, errors.WithMessagef(err, "mapping family %d to domain", s.Family)
			}

			si := SocketInfo{
				L4Proto:    proto,
				Family:     fam,
				State:      toConnState(proto, s.State),
				LocalAddr:  toAddr(s.ID.Source, s.ID.SourcePort, s.Family),
				RemoteAddr: toAddr(s.ID.Destination, s.ID.DestinationPort, s.Family),
				Ifname:     strings.Join(ifByAddr.lookup(s.ID.Source), ","),
				Inode:      inode,
				Process:    procInfo,
			}
			ret = append(ret, si)
		}
	}

	return ret, nil
}
