package ss

import (
	"net/netip"

	ss "github.com/PRO-Robotech/sgroups/internal/sg-agent/socketscan"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	cdto "github.com/PRO-Robotech/sgroups/internal/shared/dto/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/pkg/filter"
	netrc "github.com/H-BF/corlib/pkg/net/resources"
	"github.com/H-BF/corlib/pkg/option"
	agentv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/agent/v1"
	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	"github.com/pkg/errors"
)

func scopeFromeReq[T interface {
	GetSelectors() []*agentv1.SockStat_Selectors
}](req T) (ret filter.Scope, err error) {
	var sc []ss.Selector
	for _, s := range req.GetSelectors() {
		var sel ss.Selector
		if sel, err = selectorFromReq(s); err != nil {
			return nil, err
		}
		sc = append(sc, sel)
	}
	return ss.ScopeBySocketSelectors{
		Selectors: sc,
	}, nil
}

func selectorFromReq(req *agentv1.SockStat_Selectors) (ss.Selector, error) {
	var err error

	ret := ss.Selector{
		Ifname: req.GetIfname(),
		Comm:   req.GetComm(),
	}
	if req != nil && req.Inode != nil {
		ret.Inode.Set(uint64(req.GetInode())) // nolint:gosec
	}
	if req != nil && req.Pid != nil {
		ret.PID.Set(int(req.GetPid())) // nolint:gosec
	}
	if req != nil && req.LocalPort != nil {
		ret.LocalPort.Set(uint16(req.GetLocalPort())) // nolint:gosec
	}
	if req != nil && req.RemotePort != nil {
		ret.RemotePort.Set(uint16(req.GetRemotePort())) // nolint:gosec
	}
	if ret.LocalAddr, err = toAddr(req.GetLocalAddr()); err != nil {
		return ret, errors.WithMessage(err, "invalid local address")
	}
	if ret.RemoteAddr, err = toAddr(req.GetRemoteAddr()); err != nil {
		return ret, errors.WithMessage(err, "invalid remote address")
	}
	if p := req.GetProtocol(); len(p) > 0 {
		var v netrc.NetworkTransport
		if err = v.FromString(p); err != nil {
			return ret, err
		}
		ret.L4Proto.Set(v)
	}

	ret.State = connStateFromReq(req.GetState(), ret.L4Proto)

	var fam domain.IpFamily
	if pbFam := req.GetFamily(); pbFam != common.IpAddrFamily_IPV_UNDEF {
		if err = cdto.Proto2Domain(cdto.DTO(pbFam, &fam)); err == nil {
			ret.Family.Set(fam)
		}
	}

	return ret, err
}

func connStateFromReq(req agentv1.ConnState, proto option.ValueOf[netrc.NetworkTransport]) (ret option.ValueOf[ss.ConnState]) {
	switch req {
	case agentv1.ConnState_ESTABLISHED:
		option.Match(proto).Some(func(p netrc.NetworkTransport) {
			switch p {
			case netrc.TCP:
				ret.Set(ss.TCPEstablished)
			case netrc.UDP:
				ret.Set(ss.UDPEstablished)
			}
		})
		if proto.IsNone() {
			ret.Set(ss.TCPUDPEstablished)
		}
	case agentv1.ConnState_SYN_SENT:
		ret.Set(ss.TCPSynSent)
	case agentv1.ConnState_SYN_RECV:
		ret.Set(ss.TCPSynRecv)
	case agentv1.ConnState_FIN_WAIT1:
		ret.Set(ss.TCPFinWait1)
	case agentv1.ConnState_FIN_WAIT2:
		ret.Set(ss.TCPFinWait2)
	case agentv1.ConnState_TIME_WAIT:
		ret.Set(ss.TCPTimeWait)
	case agentv1.ConnState_CLOSE:
		ret.Set(ss.TCPUDPClose)
	case agentv1.ConnState_CLOSE_WAIT:
		ret.Set(ss.TCPCloseWait)
	case agentv1.ConnState_LAST_ACK:
		ret.Set(ss.TCPLastAck)
	case agentv1.ConnState_LISTEN:
		ret.Set(ss.TCPListen)
	case agentv1.ConnState_CLOSING:
		ret.Set(ss.TCPClosing)
	case agentv1.ConnState_NEW_SYN_RECV:
		ret.Set(ss.TCPNewSynRecv)
	}
	return ret
}

func connStateToPb(src ss.ConnState) (ret agentv1.ConnState) {
	switch src {
	case ss.TCPEstablished, ss.UDPEstablished, ss.TCPUDPEstablished:
		ret = agentv1.ConnState_ESTABLISHED
	case ss.TCPSynSent:
		ret = agentv1.ConnState_SYN_SENT
	case ss.TCPSynRecv:
		ret = agentv1.ConnState_SYN_RECV
	case ss.TCPFinWait1:
		ret = agentv1.ConnState_FIN_WAIT1
	case ss.TCPFinWait2:
		ret = agentv1.ConnState_FIN_WAIT2
	case ss.TCPTimeWait:
		ret = agentv1.ConnState_TIME_WAIT
	case ss.TCPUDPClose:
		ret = agentv1.ConnState_CLOSE
	case ss.TCPCloseWait:
		ret = agentv1.ConnState_CLOSE_WAIT
	case ss.TCPLastAck:
		ret = agentv1.ConnState_LAST_ACK
	case ss.TCPListen:
		ret = agentv1.ConnState_LISTEN
	case ss.TCPNewSynRecv:
		ret = agentv1.ConnState_NEW_SYN_RECV
	}

	return ret
}

func toAddr(addr string) (ret option.ValueOf[netip.Addr], err error) {
	if len(addr) == 0 {
		return ret, nil
	}
	var ip netip.Addr
	if ip, err = netip.ParseAddr(addr); err == nil {
		ret.Set(ip)
	}
	return ret, err
}

func socketInfoToPb(si []ss.SocketInfo) (ret []*agentv1.SockStat, err error) {
	ret = misc.Tern(len(si) > 0, make([]*agentv1.SockStat, len(si)), nil)
	for i, s := range si {
		ret[i] = &agentv1.SockStat{
			Protocol:   s.L4Proto.String(),
			State:      connStateToPb(s.State),
			LocalAddr:  s.LocalAddr.Addr().String(),
			LocalPort:  int32(s.LocalAddr.Port()),
			RemoteAddr: s.RemoteAddr.Addr().String(),
			RemotePort: int32(s.RemoteAddr.Port()),
			Ifname:     s.Ifname,
			Inode:      int64(s.Inode), // nolint:gosec
		}
		if err = cdto.Domain2Proto(cdto.DTO(s.Family, &ret[i].Family)); err != nil {
			return nil, err
		}
		ret[i].Processes = misc.Tern(len(s.Process) > 0, make([]*agentv1.ProcessInfo, len(s.Process)), nil)
		for j, p := range s.Process {
			ret[i].Processes[j] = procInfoToPb(p)
		}
	}
	return ret, nil
}

func procInfoToPb(src ss.ProcessInfo) (ret *agentv1.ProcessInfo) {
	ret = &agentv1.ProcessInfo{
		Pid:     int32(src.PID), // nolint:gosec
		Fd:      int32(src.FD),  // nolint:gosec
		Comm:    src.Comm,
		CmdLine: src.Cmdline,
		Exe:     src.Exe,
	}
	return ret
}
