package sgroups

import (
	"fmt"
	"maps"
	"net"
	"net/netip"
	"regexp"
	"slices"
	"strings"

	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/pkg/dict"
	oz "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

// Validate validates the ResourceName according to the specified rules.
func (rn ResourceName) Validate() error {
	return oz.Validate(
		string(rn), oz.Required.Error("resource name is required"), isName,
	)
}

// Validate validates the ResourceNamespace according to the specified rules.
func (rn ResourceNamespace) Validate() error {
	return oz.Validate(
		string(rn), oz.Required.Error("resource namespace is required"), isName,
	)
}

// Validate validates the DisplayName according to the specified rules.
func (dn DisplayName) Validate() error {
	const maxLen = 63
	return oz.Validate(
		string(dn),
		oz.RuneLength(0, maxLen).
			Error(fmt.Sprintf("display name must be no longer than %d characters", maxLen)),
	)
}

// Validate validates cluster scoped metadata
func (m ClusterScopeMetadataIdentity) Validate() error {
	hasUID := m.UID != uuid.Nil
	hasName := m.Name != ""
	if !hasUID && !hasName {
		return errors.New("at least one of UID or Name must be set")
	}
	return oz.ValidateStruct(
		&m,
		oz.Field(&m.UID, oz.Skip.When(!hasUID), uuidValidator),
		oz.Field(&m.Name, oz.Skip.When(!hasName)),
	)
}

// Validate validates namespaced metadata
func (n NamespacedMetadataIdentity) Validate() error {
	hasUID := n.UID != uuid.Nil
	hasResIdentity := n.Name != "" && n.Namespace != ""
	if !hasUID && !hasResIdentity {
		return errors.New("at least one of UID or Name and Namespace must be set")
	}
	return oz.ValidateStruct(
		&n,
		oz.Field(&n.UID, oz.Skip.When(!hasUID), uuidValidator),
		oz.Field(&n.Name, oz.Skip.When(!hasResIdentity)),
		oz.Field(&n.Namespace, oz.Skip.When(!hasResIdentity)),
	)
}

// Validate validates the metadata of a resource.
func (m Metadata[T]) Validate() error {
	return oz.ValidateStruct(&m, oz.Field(&m.ID))
}

// Validate -
func (nw IPNet) Validate() error {
	return oz.ValidateStruct(&nw,
		oz.Field(&nw.IP,
			oz.Required.Error("IP of CIDR is not set"),
			validateRule(func(ip net.IP) error {
				switch len(ip) {
				case net.IPv4len, net.IPv6len:
				default:
					return errors.New("IP of net is invalid")
				}
				return nil
			}),
			validateRule(func(ip net.IP) error {
				masked := ip.Mask(nw.Mask)
				if !ip.Equal(masked) {
					canonical := net.IPNet{IP: masked, Mask: nw.Mask}
					return errors.Errorf("CIDR '%s' is not in canonical form, use '%s'",
						nw.String(), canonical.String())
				}
				return nil
			}),
		),
		oz.Field(&nw.Mask, validateRule(func(mask net.IPMask) error {
			return misc.Tern(len(mask) != len(nw.IP), errors.New("net mask is invalid"), nil)
		})),
	)
}

// Validate validates the Namespace according to the specified rules.
func (n Namespace) Validate() error {
	return oz.ValidateStruct(&n,
		oz.Field(&n.Metadata),
		oz.Field(&n.Spec),
	)
}

// Validate validates the NamespaceSpec according to the specified rules.
func (n NamespaceSpec) Validate() error {
	return oz.ValidateStruct(&n, oz.Field(&n.DisplayName))
}

// Validate available events for resource watch
func (nt ResourceEventType) Validate() error {
	return requireOneOf(nt, evtToStr)
}

// Validate PolicyAction validator
func (a PolicyAction) Validate() error {
	return requireOneOf(a, actionToStr)
}

// Validate Traffic enum
func (tfc Traffic) Validate() error {
	return requireOneOf(tfc, trafficToStr)
}

// Validate IPproto enum
func (p IPproto) Validate() error {
	return requireOneOf(p, protoToStr)
}

// Validate IpFamily enum
func (p IpFamily) Validate() error {
	return requireOneOf(p, ipvToStr)
}

// Validate available resource types
func (rt ResourceType) Validate() error {
	sli := slices.Collect(maps.Keys(availableResourceTypes))
	return oz.Validate(rt,
		oz.In(misc.SliceToAny(sli)...).
			Error(
				fmt.Sprintf("must be one of [%s]",
					strings.Join(
						misc.SliceToString(sli),
						",",
					),
				),
			),
	)
}

// Validate available endpoint types
func (et EndpointType) Validate() error {
	return requireOneOf(et, epToStr)
}

// Validate validates the AddressGroup according to the specified rules.
func (a AddressGroup) Validate() error {
	return oz.ValidateStruct(&a,
		oz.Field(&a.Metadata),
		oz.Field(&a.Spec),
	)
}

// Validate validates the AgSpec according to the specified rules.
func (a AgSpec) Validate() error {
	return oz.ValidateStruct(&a,
		oz.Field(&a.DisplayName),
		oz.Field(&a.DefaultAction),
	)
}

// Validate validates the Network according to the specified rules.
func (n Network) Validate() error {
	return oz.ValidateStruct(&n,
		oz.Field(&n.Metadata),
		oz.Field(&n.Spec))
}

// Validate validates the NetworkSpec according to the specified rules.
func (s NetworkSpec) Validate() error {
	return oz.ValidateStruct(&s,
		oz.Field(&s.DisplayName),
		oz.Field(&s.CIDR),
	)
}

// Validate validates the Host according to the specified rules.
func (n Host) Validate() error {
	return oz.ValidateStruct(&n,
		oz.Field(&n.Metadata),
		oz.Field(&n.Spec),
	)
}

// Validate validates the HostSpec according to the specified rules.
func (h HostSpec) Validate() error {
	return oz.ValidateStruct(&h,
		oz.Field(&h.DisplayName),
		oz.Field(&h.IPs),
	)
}

// Validate validates the dual stack IPs according to the specified rules.
func (d DualStackIPs) Validate() error {
	onlyV4 := validateRule(func(set dict.HSet[netip.Addr]) error {
		for a := range set.Iterate {
			if !a.IsValid() {
				return errors.New("invalid IP address")
			}
			if !a.Is4() {
				return errors.Errorf("must contain only IPv4 addresses, got %s", a.String())
			}
		}
		return nil
	})

	onlyV6 := validateRule(func(set dict.HSet[netip.Addr]) error {
		for a := range set.Iterate {
			if !a.IsValid() {
				return errors.New("invalid IP address")
			}
			if !a.Is6() || a.Is4In6() {
				return errors.Errorf("must contain only IPv6 addresses, got %s", a.String())
			}
		}
		return nil
	})

	return oz.ValidateStruct(
		&d,
		oz.Field(&d.IPv4, onlyV4),
		oz.Field(&d.IPv6, onlyV6),
	)
}

// Validate validates the HostBinding according to the specified rules.
func (h HostBinding) Validate() error {
	return oz.ValidateStruct(&h,
		oz.Field(&h.Metadata),
		oz.Field(&h.Spec),
	)
}

// Validate validates the HostBindingSpec according to the specified rules.
func (h HostBindingSpec) Validate() error {
	return oz.ValidateStruct(&h,
		oz.Field(&h.DisplayName),
		oz.Field(&h.AddressGroup),
		oz.Field(&h.Host),
	)
}

// Validate validates the ResourceIdentifier according to the specified rules.
func (r ResourceIdentifier) Validate() error {
	return oz.ValidateStruct(&r,
		oz.Field(&r.Name),
		oz.Field(&r.Namespace),
	)
}

// Validate validates the NetworkBinding according to the specified rules.
func (n NetworkBinding) Validate() error {
	return oz.ValidateStruct(&n,
		oz.Field(&n.Metadata),
		oz.Field(&n.Spec),
	)
}

// Validate validates the NetworkBindingSpec according to the specified rules.
func (n NetworkBindingSpec) Validate() error {
	return oz.ValidateStruct(&n,
		oz.Field(&n.DisplayName),
		oz.Field(&n.AddressGroup),
		oz.Field(&n.Network),
	)
}

// Validate validates the Rule according to the specified rules.
func (r Rule) Validate() error {
	return oz.ValidateStruct(&r,
		oz.Field(&r.Metadata),
		oz.Field(&r.Spec),
	)
}

type ruleSpecCheck = func(RuleSpec) error

// Validate validates the RuleSpec according to the specified rules.
func (r RuleSpec) Validate() error {
	ruleSpecChecks := [...]ruleSpecCheck{
		validateLocalEndpoint,
		validateRemoteEndpoint,
		validateEpNullRequiresAgLocal,
		validateTrafficByRemote,
		validateTransportTypeByRemote,
		validateTransportByMatrix,
	}

	for _, check := range ruleSpecChecks {
		if err := check(r); err != nil {
			return err
		}
	}
	return oz.ValidateStruct(&r,
		oz.Field(&r.DisplayName),
		oz.Field(&r.Action),
		oz.Field(&r.Traffic),
		oz.Field(&r.Local),
		oz.Field(&r.Remote),
		oz.Field(&r.Transport),
	)
}

// Validate validates the Transport according to the specified rules.
func (t Transport[T]) Validate() error {
	switch any(t).(type) {
	case L4Transport:
		if !misc.IsIn(t.Proto, TCP, UDP) {
			return errors.Errorf("l4 transport protocol must be TCP or UDP, got %v", t.Proto)
		}
	case IcmpTransport:
		if t.Proto != ICMP {
			return errors.Errorf("icmp transport protocol must be ICMP, got %v", t.Proto)
		}
	}
	if !misc.IsIn(t.IPv, IPv4, IPv6) {
		return errors.Errorf("ip address family is required; must be IPv4 or IPv6, got %v", t.IPv)
	}
	return oz.ValidateStruct(&t,
		oz.Field(&t.Proto),
		oz.Field(&t.IPv),
		oz.Field(&t.Entries, oz.Required, oz.Each(validateRule(func(e Entry[T]) error {
			return e.Validate()
		}))),
	)
}

// Validate validates the Entry according to the specified rules.
func (e Entry[T]) Validate() error {
	return oz.ValidateStruct(&e,
		oz.Field(&e.Value),
		oz.Field(&e.Value, validateRule(func(v T) error {
			return validateEntryVal(v)
		})),
	)
}

// Validate validates the Endpoint according to the specified rules.
func (e ep) Validate() error {
	return oz.ValidateStruct(&e,
		oz.Field(&e.Name),
		oz.Field(&e.Namespace),
		oz.Field(&e.Type),
	)
}

// Validate validates the remote Endpoint according to the specified rules.
func (e epRemote[T]) Validate() error {
	return oz.ValidateStruct(&e,
		oz.Field(&e.Type),
		oz.Field(&e.Value),
	)
}

// Validate validates the Service according to the specified rules.
func (s Service) Validate() error {
	return oz.ValidateStruct(&s,
		oz.Field(&s.Metadata),
		oz.Field(&s.Spec),
	)
}

// Validate validates the ServiceSpec according to the specified rules.
func (s ServiceSpec) Validate() error {
	return oz.ValidateStruct(&s,
		oz.Field(&s.DisplayName),
		oz.Field(&s.Transports,
			validateRule(validateUniqueTransports),
			oz.Each(validateRule(func(t TransportSpec) error {
				return t.Validate()
			})),
		),
	)
}

// Validate validates the ServiceBinding according to the specified rules.
func (sb ServiceBinding) Validate() error {
	return oz.ValidateStruct(&sb,
		oz.Field(&sb.Metadata),
		oz.Field(&sb.Spec),
	)
}

// Validate validates the ServiceBindingSpec according to the specified rules.
func (sb ServiceBindingSpec) Validate() error {
	return oz.ValidateStruct(&sb,
		oz.Field(&sb.DisplayName),
		oz.Field(&sb.AddressGroup),
		oz.Field(&sb.Service),
	)
}

// ValidatePortRange portrange model validate
func ValidatePortRange(pr PortRange, canBeNull bool) error {
	if pr.IsNull() && !canBeNull {
		return ErrUnexpectedNullPortRange
	}
	return nil
}

func validateEntryVal[T entryVariants[T]](val T) error {
	switch v := any(val).(type) {
	case PortRanges:
		return oz.Validate(v, portRangesValidator)
	case IcmpTypes:
		return oz.Validate(v, icmpTypesValidator)
	}
	return nil
}

func requireOneOf[K comparable](v K, validNames map[K]string) error {
	if _, ok := validNames[v]; ok {
		return nil
	}
	names := slices.Sorted(maps.Values(validNames))
	return errors.Errorf("must be one of [%s]", strings.Join(names, ","))
}

func validateUniqueTransports(ts []TransportSpec) error {
	type key struct {
		Proto IPproto
		IPv   IpFamily
	}

	if len(lo.UniqBy(ts,
		func(t TransportSpec) key {
			return key{t.GetProto(), t.GetIPv()}
		})) != len(ts) {
		return errors.New("duplicate transports: same protocol and IP family")
	}

	return nil
}

func validateLocalEndpoint(r RuleSpec) error {
	local, err := LocalFromSpec(r.Local)
	if err != nil {
		return err
	}
	if !misc.IsIn(local.Type, AddressGroupEp, ServiceEp) {
		err = errors.Errorf("unknown local endpoint resource type %T", local.Type)
	}
	return err
}

func validateRemoteEndpoint(r RuleSpec) error {
	switch t := r.Remote.(type) {
	case EpRemote:
		if !misc.IsIn(t.Type, AddressGroupEp, ServiceEp) {
			return errors.Errorf("not valid remote endpoint resource type %T", t.Type)
		}
	case EpCIDR:
		if t.Type != CidrEp {
			return errors.Errorf("not valid remote endpoint resource type %T", t.Type)
		}
	case EpFQDN:
		if t.Type != FqdnEp {
			return errors.Errorf("not valid remote endpoint resource type %T", t.Type)
		}
	case EpNull:
	default:
		return errors.Errorf("unknown remote endpoint type %T", r.Remote)
	}
	return nil
}

func validateEpNullRequiresAgLocal(r RuleSpec) error {
	if _, ok := r.Remote.(EpNull); !ok {
		return nil
	}
	local, err := LocalFromSpec(r.Local)
	if err != nil {
		return err
	}
	if local.Type != AddressGroupEp {
		err = errors.Errorf(
			"remote endpoint of type EpNull is only allowed for local endpoints of type AddressGroupEp, got %T",
			local.Type)
	}
	return err
}

func validateTrafficByRemote(r RuleSpec) error {
	switch r.Remote.(type) {
	case EpFQDN:
		if r.Traffic != EGRESS {
			return errors.Errorf("FQDN rules must have traffic EGRESS, got %s", r.Traffic)
		}
	case EpCIDR:
		if r.Traffic == BOTH {
			return errors.New("CIDR rules must have traffic INGRESS or EGRESS, BOTH is not supported")
		}
	case EpNull:
		if r.Traffic == BOTH {
			return errors.New(
				"Ag2IcmpRule (AG, ICMP, no remote) must have traffic INGRESS or EGRESS, BOTH is not supported")
		}
	}
	return nil
}

func validateTransportTypeByRemote(r RuleSpec) error {
	switch r.Remote.(type) {
	case EpFQDN:
		if _, ok := r.Transport.(L4Transport); !ok {
			return errors.Errorf(
				"L4 transport expected for remote endpoint of type FqdnEp, got %T",
				r.Transport)
		}
	case EpNull:
		if _, ok := r.Transport.(IcmpTransport); !ok {
			return errors.Errorf(
				"ICMP transport expected for remote endpoint of type EpNull, got %T",
				r.Transport)
		}
	}
	return nil
}

func validateTransportByMatrix(r RuleSpec) error {
	switch r.Remote.(type) {
	case EpFQDN, EpNull:
		return nil
	}

	local, err := LocalFromSpec(r.Local)
	if err != nil {
		return err
	}
	var remoteType EndpointType
	switch rem := r.Remote.(type) {
	case EpRemote:
		remoteType = rem.Type
	case EpCIDR:
		remoteType = rem.Type
	}

	destinationIsService := false
	switch r.Traffic {
	case INGRESS:
		destinationIsService = local.Type == ServiceEp
	case EGRESS, BOTH:
		destinationIsService = remoteType == ServiceEp
	}

	isNullableTransport := IsNullableTransport(r.Transport)

	switch {
	case destinationIsService && !isNullableTransport:
		return errors.Errorf(
			"spec.transport must not be set when traffic destination is a Service "+
				"(traffic=%s, local=%s, remote=%s), got %T",
			r.Traffic, local.Type, remoteType, r.Transport)
	case !destinationIsService && isNullableTransport:
		return errors.Errorf(
			"spec.transport is required when traffic destination is not a Service "+
				"(traffic=%s, local=%s, remote=%s)",
			r.Traffic, local.Type, remoteType)
	}
	return nil
}

type validatorFunc[T any] func(T) error

func (x validatorFunc[T]) Validate(v any) error {
	return x(v.(T))
}

func validateRule[T any](f func(T) error) validatorFunc[T] {
	return f
}

var (
	reName = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
	isName = oz.Match(reName).Error("must consist of lower-case alphanumeric characters or '-', must start and end with an alphanumeric character, and be no longer than 63 characters")

	_             oz.Rule = (validateRule[UUID])(nil)
	uuidValidator         = validateRule(func(v UUID) error {
		if v == uuid.Nil {
			return errors.New("UUID is zero")
		}
		return nil
	})
	portRangesValidator = validateRule(func(pr PortRanges) error {
		if pr.Len() == 0 {
			return errors.New("ports are required for TCP/UDP transport entry")
		}

		for p := range pr.Iterate {
			if err := oz.Validate(p, portRangeValidator); err != nil {
				return errors.WithMessagef(err, "on validate ports(%s) found bad range(%s)", pr, p)
			}
		}
		return nil
	})
	portRangeValidator = validateRule(func(pr PortRange) error {
		if pr.IsNull() {
			return ErrUnexpectedNullPortRange
		}
		return nil
	})
	icmpTypesValidator = validateRule(func(t IcmpTypes) error {
		if t.Len() == 0 {
			return ErrUnexpectedEmptyIcmpTypes
		}
		return nil
	})
)

var (
	// ErrUnexpectedNullPortRange -
	ErrUnexpectedNullPortRange = errors.New("unexpected null port range")

	// ErrUnexpectedEmptyIcmpTypes -
	ErrUnexpectedEmptyIcmpTypes = errors.New("unexpected empty ICMP types")
)
