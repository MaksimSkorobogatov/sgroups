package sgroups

import (
	"slices"

	"github.com/H-BF/corlib/pkg/dict"
	netrc "github.com/H-BF/corlib/pkg/net/resources"
)

type (
	// TransportSpec -
	TransportSpec interface {
		isTransport()
		Validate() error
		GetProto() IPproto
		GetIPv() IpFamily
		IsEq(other TransportSpec) bool
	}

	entryVariants[T any] interface {
		IcmpTypes | PortRanges
		Eq(T) bool
	}

	// Transport -
	Transport[T entryVariants[T]] struct {
		Proto   IPproto
		IPv     IpFamily
		Entries []Entry[T]
	}

	// Entry -
	Entry[T entryVariants[T]] struct {
		Description string
		Comment     string
		Value       T
	}

	// IcmpEntry -
	IcmpEntry = Entry[IcmpTypes]

	// PortEntry -
	PortEntry = Entry[PortRanges]

	// IcmpTransport -
	IcmpTransport = Transport[IcmpTypes]

	// L4Transport -
	L4Transport = Transport[PortRanges]

	// NullTransport -
	NullTransport struct{}
)

var (
	_ TransportSpec = IcmpTransport{}
	_ TransportSpec = L4Transport{}
	_ TransportSpec = NullTransport{}
)

// GetProto returns the IPproto of the transport
func (t Transport[T]) GetProto() IPproto {
	return t.Proto
}

// GetIPv returns transport IP family.
func (t Transport[T]) GetIPv() IpFamily {
	return t.IPv
}

// IsEq -
func (t Transport[T]) IsEq(other TransportSpec) bool {
	o, ok := other.(Transport[T])
	if !ok {
		return false
	}
	return t.Proto == o.Proto && t.IPv == o.IPv &&
		slices.EqualFunc(t.Entries, o.Entries, func(a, b Entry[T]) bool {
			return a.IsEq(b)
		})
}

// IsEq -
func (et Entry[T]) IsEq(other Entry[T]) bool {
	return et.Description == other.Description &&
		et.Comment == other.Comment &&
		et.Value.Eq(other.Value)
}

func (t Transport[T]) isTransport() {}

func (NullTransport) isTransport() {}

// Validate impl TransportSpec
func (NullTransport) Validate() error {
	return nil
}

// GetProto impl TransportSpec
func (NullTransport) GetProto() (ret IPproto) {
	return ret
}

// GetIPv impl TransportSpec
func (NullTransport) GetIPv() (ret IpFamily) {
	return ret
}

// IsEq impl TransportSpec
func (NullTransport) IsEq(other TransportSpec) bool {
	_, ok := other.(NullTransport)
	return ok
}

// IsNullableTransport -
func IsNullableTransport(t TransportSpec) bool {
	if t == nil {
		return true
	}
	_, ok := t.(NullTransport)
	return ok
}

// IcmpFromTransport -
func IcmpFromTransport(t TransportSpec) (ret []netrc.ICMP) {
	it, ok := t.(IcmpTransport)
	if !ok {
		return nil
	}
	for _, et := range it.Entries {
		ret = append(ret, netrc.ICMP{
			IPv:   uint8(t.GetIPv()),
			Types: dict.RBSet[uint8](et.Value),
		})
	}
	return ret
}

// PortFromTransport -
func PortFromTransport(t TransportSpec) (ret []PortRanges) {
	lt, ok := t.(L4Transport)
	if !ok {
		return nil
	}
	for _, et := range lt.Entries {
		ret = append(ret, et.Value)
	}
	return ret
}
