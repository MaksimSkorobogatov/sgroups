package sgroups

import (
	"fmt"
	"strings"

	netrc "github.com/H-BF/corlib/pkg/net/resources"
	"github.com/H-BF/corlib/pkg/option"
	"github.com/pkg/errors"
)

// IPproto -
type IPproto uint8

const (
	// PROTOCOL_UNDEF is not set.
	PROTOCOL_UNDEF IPproto = iota
	// TCP -
	TCP
	// UDP -
	UDP
	// ICMP -
	ICMP
)

var protoToStr = map[IPproto]string{
	TCP:  "tcp",
	UDP:  "udp",
	ICMP: "icmp",
}

var strToProto = map[string]IPproto{
	protoToStr[TCP]:  TCP,
	protoToStr[UDP]:  UDP,
	protoToStr[ICMP]: ICMP,
}

// String impl Stringer
func (p IPproto) String() string {
	if s, ok := protoToStr[p]; ok {
		return s
	}
	return fmt.Sprintf("Undef(%v)", int(p))
}

// FromString init from string
func (p *IPproto) FromString(s string) error {
	v, ok := strToProto[strings.ToLower(s)]
	if !ok {
		return errors.WithMessage(fmt.Errorf("unknown value '%s'", s), "IPproto")
	}
	*p = v
	return nil
}

// IsEq -
func (p IPproto) IsEq(other IPproto) bool {
	return p == other
}

// L4Proto -
func (p IPproto) L4Proto() (ret option.ValueOf[netrc.NetworkTransport]) {
	switch p {
	case TCP:
		ret.Set(netrc.TCP)
	case UDP:
		ret.Set(netrc.UDP)
	}
	return ret
}
