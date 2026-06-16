package socketscan

import (
	"fmt"

	netrc "github.com/H-BF/corlib/pkg/net/resources"
	"golang.org/x/sys/unix"
)

// ConnState - connection state for TCP/UDP socket. Values correspond to sk_state in kernel.
type ConnState uint8

// TCP/UDP connection state variants
const (
	TCPEstablished ConnState = iota + 1
	UDPEstablished
	TCPUDPEstablished
	TCPSynSent
	TCPSynRecv
	TCPFinWait1
	TCPFinWait2
	TCPTimeWait
	TCPUDPClose
	TCPCloseWait
	TCPLastAck
	TCPListen
	TCPClosing
	TCPNewSynRecv
)

var conStateToStr = map[ConnState]string{
	TCPEstablished:    "ESTABLISHED",
	UDPEstablished:    "ESTABLISHED",
	TCPUDPEstablished: "ESTABLISHED",
	TCPSynSent:        "SYN_SENT",
	TCPSynRecv:        "SYN_RECV",
	TCPFinWait1:       "FIN_WAIT1",
	TCPFinWait2:       "FIN_WAIT2",
	TCPTimeWait:       "TIME_WAIT",
	TCPUDPClose:       "CLOSE",
	TCPCloseWait:      "CLOSE_WAIT",
	TCPLastAck:        "LAST_ACK",
	TCPListen:         "LISTEN",
	TCPClosing:        "CLOSING",
	TCPNewSynRecv:     "NEW_SYN_RECV",
}

// IsEq returns true if other is the same state
func (s ConnState) IsEq(other ConnState) bool {
	if s == other {
		return true
	}
	if s == TCPUDPEstablished {
		return other == TCPEstablished || other == UDPEstablished
	}
	return false
}

// String -
func (s ConnState) String() string {
	if name, ok := conStateToStr[s]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(%d)", uint8(s))
}

func toConnState(proto netrc.NetworkTransport, state uint8) (ret ConnState) {
	if st, ok := stateToConnState[proto][state]; ok {
		return st
	}

	return ret
}

var stateToConnState = map[netrc.NetworkTransport]map[uint8]ConnState{
	netrc.TCP: {
		unix.BPF_TCP_ESTABLISHED:  TCPEstablished,
		unix.BPF_TCP_SYN_SENT:     TCPSynSent,
		unix.BPF_TCP_SYN_RECV:     TCPSynRecv,
		unix.BPF_TCP_FIN_WAIT1:    TCPFinWait1,
		unix.BPF_TCP_FIN_WAIT2:    TCPFinWait2,
		unix.BPF_TCP_TIME_WAIT:    TCPTimeWait,
		unix.BPF_TCP_CLOSE:        TCPUDPClose,
		unix.BPF_TCP_CLOSE_WAIT:   TCPCloseWait,
		unix.BPF_TCP_LAST_ACK:     TCPLastAck,
		unix.BPF_TCP_LISTEN:       TCPListen,
		unix.BPF_TCP_CLOSING:      TCPClosing,
		unix.BPF_TCP_NEW_SYN_RECV: TCPNewSynRecv,
	},
	netrc.UDP: {
		unix.BPF_TCP_CLOSE:       TCPUDPClose,
		unix.BPF_TCP_ESTABLISHED: UDPEstablished,
	},
}
