package sgroups

import "github.com/PRO-Robotech/sgroups/internal/shared/misc"

// ResourceType -
type ResourceType string

// Available resource types
const (
	UnknownResource        ResourceType = "Unknown"
	NamespaceResource      ResourceType = "Namespace"
	AddressGroupResource   ResourceType = "AddressGroup"
	NetworkResource        ResourceType = "Network"
	HostResource           ResourceType = "Host"
	HostBindingResource    ResourceType = "HostBinding"
	NetworkBindingResource ResourceType = "NetworkBinding"
	Ag2AgRule              ResourceType = "Ag2AgRule"
	Ag2AgIcmpRule          ResourceType = "Ag2AgIcmpRule"
	Ag2IcmpRule            ResourceType = "Ag2IcmpRule"
	Ag2CidrRule            ResourceType = "Ag2CidrRule"
	Ag2CidrIcmpRule        ResourceType = "Ag2CidrIcmpRule"
	Ag2FqdnRule            ResourceType = "Ag2FqdnRule"
	Ag2SvcRule             ResourceType = "Ag2SvcRule"
	Ag2SvcIcmpRule         ResourceType = "Ag2SvcIcmpRule"
	Svc2SvcRule            ResourceType = "Svc2SvcRule"
	Svc2FqdnRule           ResourceType = "Svc2FqdnRule"
	Svc2CidrRule           ResourceType = "Svc2CidrRule"
	Svc2CidrIcmpRule       ResourceType = "Svc2CidrIcmpRule"
	Svc2AgRule             ResourceType = "Svc2AgRule"
	Svc2AgIcmpRule         ResourceType = "Svc2AgIcmpRule"
	ServiceResource        ResourceType = "Service"
	ServiceBindingResource ResourceType = "ServiceBinding"
)

var (
	// IsRule -
	IsRule = func(rt ResourceType) bool {
		return misc.IsIn(rt, RuleTypes...)
	}

	// RuleTypes -
	RuleTypes = misc.Sli(
		Ag2AgRule,
		Ag2AgIcmpRule,
		Ag2IcmpRule,
		Ag2CidrRule,
		Ag2CidrIcmpRule,
		Ag2FqdnRule,
		Ag2SvcRule,
		Ag2SvcIcmpRule,
		Svc2SvcRule,
		Svc2FqdnRule,
		Svc2CidrRule,
		Svc2CidrIcmpRule,
		Svc2AgRule,
		Svc2AgIcmpRule,
	)

	// DefRulePriority -
	DefRulePriority = map[ResourceType]int16{
		Ag2AgRule:        -200,
		Ag2AgIcmpRule:    -300,
		Ag2IcmpRule:      -400,
		Ag2CidrRule:      300,
		Ag2CidrIcmpRule:  200,
		Ag2FqdnRule:      100,
		Ag2SvcRule:       -50,
		Ag2SvcIcmpRule:   -60,
		Svc2SvcRule:      -350,
		Svc2FqdnRule:     50,
		Svc2CidrRule:     -150,
		Svc2CidrIcmpRule: -180,
		Svc2AgRule:       -100,
		Svc2AgIcmpRule:   -110,
	}
)

var availableResourceTypes = map[ResourceType]struct{}{
	NamespaceResource:      {},
	AddressGroupResource:   {},
	NetworkResource:        {},
	HostResource:           {},
	HostBindingResource:    {},
	NetworkBindingResource: {},
	Ag2AgRule:              {},
	Ag2AgIcmpRule:          {},
	Ag2IcmpRule:            {},
	Ag2CidrRule:            {},
	Ag2CidrIcmpRule:        {},
	Ag2FqdnRule:            {},
	Ag2SvcRule:             {},
	Ag2SvcIcmpRule:         {},
	Svc2SvcRule:            {},
	Svc2FqdnRule:           {},
	Svc2CidrRule:           {},
	Svc2CidrIcmpRule:       {},
	Svc2AgRule:             {},
	Svc2AgIcmpRule:         {},
	ServiceResource:        {},
	ServiceBindingResource: {},
}

// String impl Stringer
func (rt ResourceType) String() string {
	if _, ok := availableResourceTypes[rt]; ok {
		return string(rt)
	}
	return string(UnknownResource)
}

// IsEq -
func (rt ResourceType) IsEq(other ResourceType) bool {
	return rt == other
}
