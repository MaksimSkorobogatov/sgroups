package resources

import (
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
)

type (
	// ICMPv4 -
	ICMPv4 = icmpV
	// ICMPv6 -
	ICMPv6 = icmpV

	icmpV struct {
		Types domain.IcmpTypes `json:"types"`
	}
)

// IsEq -
func (icmp icmpV) IsEq(other icmpV) bool {
	return icmp.Types.Eq(other.Types)
}
