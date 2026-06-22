package dto

import (
	"encoding/json"

	pg "github.com/PRO-Robotech/sgroups/internal/sg-server/repository/pg/domain"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/pkg/ranges"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

func init() {
	dto.Register[pg.ResourceIdentifier, domain.ResourceIdentifier](resIDToDomain)
	dto.Register[pg.ResourceEvent, domain.NamespaceEvent](nsEventToDomain)
	dto.Register[pg.ResourceEvent, domain.AddressGroupEvent](agEventToDomain)
	dto.Register[pg.ResourceEvent, domain.NetworkEvent](networkEventToDomain)
	dto.Register[pg.ResourceEvent, domain.HostEvent](hostEventToDomain)
	dto.Register[pg.ResourceEvent, domain.HostBindingEvent](hbEventToDomain)
	dto.Register[pg.ResourceEvent, domain.NetworkBindingEvent](nbEventToDomain)
	dto.Register[pg.ResourceEvent, domain.RuleEvent](rlEventToDomain)
	dto.Register[pg.NsPK, domain.ClusterScopeMetadataIdentity](nsIdToDomain)
	dto.Register[pg.ResPK, domain.NamespacedMetadataIdentity](resIdToDomain)
	dto.Register[pg.ResourceEvent, domain.ServiceEvent](serviceEventToDomain)
	dto.Register[pg.ResourceEvent, domain.ServiceBindingEvent](sbEventToDomain)
	dto.Register[pg.Namespace, domain.Namespace](namespaceToDomain)
	dto.Register[pg.AddressGroup, domain.AddressGroup](agToDomain)
	dto.Register[pg.Network, domain.Network](networkToDomain)
	dto.Register[*pg.HostEndpoints, *domain.HostEndpoints](hostEndpointsToDomain)
	dto.Register[pg.HostInfo, domain.HostInfo](hostInfoToDomain)
	dto.Register[pg.Host, domain.Host](hostToDomain)
	dto.Register[pg.HostBinding, domain.HostBinding](hbToDomain)
	dto.Register[pg.NetworkBinding, domain.NetworkBinding](nbToDomain)
	dto.Register[pg.PortRange, domain.PortRange](portRangeToDomain)
	dto.Register[pg.PortMultirange, domain.PortRanges](portRangesToDomain)
	dto.Register[pg.IcmpEntries, domain.IcmpEntry](icmpEntriesToDomain)
	dto.Register[pg.PortEntries, domain.PortEntry](portEntriesToDomain)
	dto.Register[pg.Rule, domain.Rule](rlToDomain)
	dto.Register[pg.UniRuleL4, domain.Rule](uniRuleL4ToDomain)
	dto.Register[pg.UniRuleIcmp, domain.Rule](uniRuleIcmpToDomain)
	dto.Register[pg.Res2ResRule, domain.Rule](res2resRuleToDomain)
	dto.Register[pg.Res2ResIcmpRule, domain.Rule](res2resIcmpRuleToDomain)
	dto.Register[pg.Res2IcmpRule, domain.Rule](res2IcmpRuleToDomain)
	dto.Register[pg.Res2CidrRule, domain.Rule](res2CidrRuleToDomain)
	dto.Register[pg.Res2CidrIcmpRule, domain.Rule](res2CidrIcmpRuleToDomain)
	dto.Register[pg.Res2FqdnRule, domain.Rule](res2FqdnRuleToDomain)
	dto.Register[pg.TransportEntry, domain.PortEntry](transportEntryToPortEntry)
	dto.Register[pg.TransportEntry, domain.IcmpEntry](transportEntryToIcmpEntry)
	dto.Register[pg.Transport, domain.TransportSpec](transportToDomain)
	dto.Register[[]pg.TransportEntry, []domain.PortEntry](transportEntriesToPortEntries)
	dto.Register[[]pg.TransportEntry, []domain.IcmpEntry](transportEntriesToIcmpEntries)
	dto.Register[pg.Service, domain.Service](serviceToDomain)
	dto.Register[pg.ServiceBinding, domain.ServiceBinding](sbToDomain)
	dto.Register[pg.ResourceRef, domain.ResourceRef](resourceRefToDomain)
}

// Pg2Domain -
func Pg2Domain[T pg2domainVariants](a T) (err error) {
	return errors.WithMessage(a.Convert(), "pg -> domain dto convertation")
}

type pg2domainVariants interface {
	*dto.Pair[pg.ResourceIdentifier, domain.ResourceIdentifier] |
		*dto.Pair[pg.ResourceEvent, domain.NamespaceEvent] |
		*dto.Pair[pg.ResourceEvent, domain.AddressGroupEvent] |
		*dto.Pair[pg.ResourceEvent, domain.NetworkEvent] |
		*dto.Pair[pg.ResourceEvent, domain.HostEvent] |
		*dto.Pair[pg.ResourceEvent, domain.HostBindingEvent] |
		*dto.Pair[pg.ResourceEvent, domain.NetworkBindingEvent] |
		*dto.Pair[pg.ResourceEvent, domain.RuleEvent] |
		*dto.Pair[pg.NsPK, domain.ClusterScopeMetadataIdentity] |
		*dto.Pair[pg.ResPK, domain.NamespacedMetadataIdentity] |
		*dto.Pair[pg.ResourceEvent, domain.ServiceEvent] |
		*dto.Pair[pg.ResourceEvent, domain.ServiceBindingEvent] |
		*dto.Pair[pg.Namespace, domain.Namespace] |
		*dto.Pair[pg.AddressGroup, domain.AddressGroup] |
		*dto.Pair[pg.Network, domain.Network] |
		*dto.Pair[*pg.HostEndpoints, *domain.HostEndpoints] |
		*dto.Pair[pg.HostInfo, domain.HostInfo] |
		*dto.Pair[pg.Host, domain.Host] |
		*dto.Pair[pg.HostBinding, domain.HostBinding] |
		*dto.Pair[pg.NetworkBinding, domain.NetworkBinding] |
		*dto.Pair[pg.PortRange, domain.PortRange] |
		*dto.Pair[pg.PortMultirange, domain.PortRanges] |
		*dto.Pair[pg.IcmpEntries, domain.IcmpEntry] |
		*dto.Pair[pg.PortEntries, domain.PortEntry] |
		*dto.Pair[pg.Rule, domain.Rule] |
		*dto.Pair[pg.UniRuleL4, domain.Rule] |
		*dto.Pair[pg.UniRuleIcmp, domain.Rule] |
		*dto.Pair[pg.Res2ResRule, domain.Rule] |
		*dto.Pair[pg.Res2ResIcmpRule, domain.Rule] |
		*dto.Pair[pg.Res2IcmpRule, domain.Rule] |
		*dto.Pair[pg.Res2CidrRule, domain.Rule] |
		*dto.Pair[pg.Res2CidrIcmpRule, domain.Rule] |
		*dto.Pair[pg.Res2FqdnRule, domain.Rule] |
		*dto.Pair[pg.TransportEntry, domain.PortEntry] |
		*dto.Pair[pg.TransportEntry, domain.IcmpEntry] |
		*dto.Pair[pg.Transport, domain.TransportSpec] |
		*dto.Pair[[]pg.TransportEntry, []domain.PortEntry] |
		*dto.Pair[[]pg.TransportEntry, []domain.IcmpEntry] |
		*dto.Pair[pg.Service, domain.Service] |
		*dto.Pair[pg.ServiceBinding, domain.ServiceBinding] |
		*dto.Pair[pg.ResourceRef, domain.ResourceRef]
	Convert() error
}

func resIDToDomain(src pg.ResourceIdentifier) (domain.ResourceIdentifier, error) {
	return domain.ResourceIdentifier{
		Name:      domain.ResourceName(src.Name),
		Namespace: domain.ResourceNamespace(src.Namespace),
	}, nil
}

func nsEventToDomain(src pg.ResourceEvent) (dst domain.NamespaceEvent, err error) { //nolint:dupl
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	dst = domain.NamespaceEvent{
		TS:              src.TS,
		ResourceVersion: src.ResourceVersion,
		ResourceType:    domain.ResourceType(src.ResourceType),
	}
	if err = dst.EventType.FromString(src.EventType); err != nil {
		return dst, errors.WithMessage(err, "event type convertation")
	}
	var pgNs pg.Namespace
	err = json.Unmarshal(src.Object, &pgNs)
	if err != nil {
		return dst, errors.WithMessage(err, "json unmarshal")
	}
	err = Pg2Domain(DTO(pgNs, &dst.Object))
	return dst, err
}

func agEventToDomain(src pg.ResourceEvent) (dst domain.AddressGroupEvent, err error) { //nolint:dupl
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	dst = domain.AddressGroupEvent{
		TS:              src.TS,
		ResourceVersion: src.ResourceVersion,
		ResourceType:    domain.ResourceType(src.ResourceType),
	}
	if err = dst.EventType.FromString(src.EventType); err != nil {
		return dst, errors.WithMessage(err, "event type convertation")
	}
	var pgAg pg.AddressGroup
	err = json.Unmarshal(src.Object, &pgAg)
	if err != nil {
		return dst, errors.WithMessage(err, "json unmarshal")
	}
	err = Pg2Domain(DTO(pgAg, &dst.Object))
	return dst, err
}

func networkEventToDomain(src pg.ResourceEvent) (dst domain.NetworkEvent, err error) { //nolint:dupl
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	dst = domain.NetworkEvent{
		TS:              src.TS,
		ResourceVersion: src.ResourceVersion,
		ResourceType:    domain.ResourceType(src.ResourceType),
	}
	if err = dst.EventType.FromString(src.EventType); err != nil {
		return dst, errors.WithMessage(err, "event type convertation")
	}
	var pgNw pg.Network
	err = json.Unmarshal(src.Object, &pgNw)
	if err != nil {
		return dst, errors.WithMessage(err, "json unmarshal")
	}
	err = Pg2Domain(DTO(pgNw, &dst.Object))
	return dst, err
}

func hostEventToDomain(src pg.ResourceEvent) (dst domain.HostEvent, err error) { //nolint:dupl
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	dst = domain.HostEvent{
		TS:              src.TS,
		ResourceVersion: src.ResourceVersion,
		ResourceType:    domain.ResourceType(src.ResourceType),
	}
	if err = dst.EventType.FromString(src.EventType); err != nil {
		return dst, errors.WithMessage(err, "event type convertation")
	}
	var pgHost pg.Host
	err = json.Unmarshal(src.Object, &pgHost)
	if err != nil {
		return dst, errors.WithMessage(err, "json unmarshal")
	}
	err = Pg2Domain(DTO(pgHost, &dst.Object))
	return dst, err
}

func hbEventToDomain(src pg.ResourceEvent) (dst domain.HostBindingEvent, err error) { //nolint:dupl
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	dst = domain.HostBindingEvent{
		TS:              src.TS,
		ResourceVersion: src.ResourceVersion,
		ResourceType:    domain.ResourceType(src.ResourceType),
	}
	if err = dst.EventType.FromString(src.EventType); err != nil {
		return dst, errors.WithMessage(err, "event type convertation")
	}
	var pgHostBinding pg.HostBinding
	err = json.Unmarshal(src.Object, &pgHostBinding)
	if err != nil {
		return dst, errors.WithMessage(err, "json unmarshal")
	}
	err = Pg2Domain(DTO(pgHostBinding, &dst.Object))
	return dst, err
}

func rlEventToDomain(src pg.ResourceEvent) (dst domain.RuleEvent, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	dst = domain.RuleEvent{
		TS:              src.TS,
		ResourceVersion: src.ResourceVersion,
		ResourceType:    domain.ResourceType(src.ResourceType),
	}
	if err = dst.EventType.FromString(src.EventType); err != nil {
		return dst, errors.WithMessage(err, "event type convertation")
	}
	switch dst.ResourceType {
	case domain.Ag2AgRule, domain.Svc2SvcRule, domain.Ag2SvcRule, domain.Svc2AgRule:
		err = rlEvtConv(src, func(pgRule pg.Res2ResRule) error {
			return Pg2Domain(DTO(pgRule, &dst.Object))
		})
	case domain.Ag2AgIcmpRule, domain.Svc2AgIcmpRule, domain.Ag2SvcIcmpRule:
		err = rlEvtConv(src, func(pgRule pg.Res2ResIcmpRule) error {
			return Pg2Domain(DTO(pgRule, &dst.Object))
		})
	case domain.Ag2IcmpRule:
		err = rlEvtConv(src, func(pgRule pg.Res2IcmpRule) error {
			return Pg2Domain(DTO(pgRule, &dst.Object))
		})
	case domain.Ag2CidrRule, domain.Svc2CidrRule:
		err = rlEvtConv(src, func(pgRule pg.Res2CidrRule) error {
			return Pg2Domain(DTO(pgRule, &dst.Object))
		})
	case domain.Ag2CidrIcmpRule, domain.Svc2CidrIcmpRule:
		err = rlEvtConv(src, func(pgRule pg.Res2CidrIcmpRule) error {
			return Pg2Domain(DTO(pgRule, &dst.Object))
		})
	case domain.Ag2FqdnRule, domain.Svc2FqdnRule:
		err = rlEvtConv(src, func(pgRule pg.Res2FqdnRule) error {
			return Pg2Domain(DTO(pgRule, &dst.Object))
		})
	default:
		return dst, errors.Errorf("unexpected rule '%s'", dst.ResourceType)
	}
	return dst, err
}

func nsIdToDomain(src pg.NsPK) (domain.ClusterScopeMetadataIdentity, error) {
	return domain.ClusterScopeMetadataIdentity{
		UID:  src.UID,
		Name: domain.ResourceName(src.Name),
	}, nil
}

func resIdToDomain(src pg.ResPK) (dst domain.NamespacedMetadataIdentity, err error) {
	err = Pg2Domain(DTO(src.NsPK, &dst.ClusterScopeMetadataIdentity))
	if err != nil {
		return dst, err
	}
	dst.Namespace = domain.ResourceNamespace(src.Namespace)
	return dst, nil
}

func namespaceToDomain(src pg.Namespace) (domain.Namespace, error) {
	dst := domain.Namespace{
		Metadata: domain.NsMetadata{
			ID: domain.ClusterScopeMetadataIdentity{
				UID:  src.UID,
				Name: domain.ResourceName(src.Name),
			},
			Labels:            src.Labels,
			Annotations:       src.Annotations,
			CreationTimestamp: src.CreationTimestamp,
			ResourceVersion:   src.ResourceVersion,
		},
		Spec: domain.CommonSpec{
			DisplayName: domain.DisplayName(src.DisplayName),
			Comment:     src.Comment,
			Description: src.Description,
		},
	}
	return dst, nil
}

func agToDomain(src pg.AddressGroup) (dst domain.AddressGroup, err error) {
	dst = domain.AddressGroup{
		Metadata: domain.ResMetadata{
			ID: domain.NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
					UID:  src.UID,
					Name: domain.ResourceName(src.Name),
				},
				Namespace: domain.ResourceNamespace(src.Namespace),
			},
			Labels:            src.Labels,
			Annotations:       src.Annotations,
			CreationTimestamp: src.CreationTimestamp,
			ResourceVersion:   src.ResourceVersion,
		},
		Spec: domain.AgSpec{
			CommonSpec: domain.CommonSpec{
				DisplayName: domain.DisplayName(src.DisplayName),
				Comment:     src.Comment,
				Description: src.Description,
			},
			Logs:  src.Logs,
			Trace: src.Trace,
		},
		Refs: misc.Tern(len(src.Refs) > 0, make([]domain.ResourceRef, len(src.Refs)), nil),
	}
	for i, r := range src.Refs {
		if err = Pg2Domain(DTO(r, &dst.Refs[i])); err != nil {
			return dst, err
		}
	}
	err = dst.Spec.DefaultAction.FromString(string(src.DefaultAction))
	return dst, err
}

func networkToDomain(src pg.Network) (dst domain.Network, err error) {
	dst = domain.Network{
		Metadata: domain.ResMetadata{
			ID: domain.NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
					UID:  src.UID,
					Name: domain.ResourceName(src.Name),
				},
				Namespace: domain.ResourceNamespace(src.Namespace),
			},
			Labels:            src.Labels,
			Annotations:       src.Annotations,
			CreationTimestamp: src.CreationTimestamp,
			ResourceVersion:   src.ResourceVersion,
		},
		Spec: domain.NetworkSpec{
			CommonSpec: domain.CommonSpec{
				DisplayName: domain.DisplayName(src.DisplayName),
				Comment:     src.Comment,
				Description: src.Description,
			},
			CIDR: domain.IPNet{IPNet: src.Network.IPNet},
		},
		Refs: misc.Tern(len(src.Refs) > 0, make([]domain.ResourceRef, len(src.Refs)), nil),
	}

	for i, r := range src.Refs {
		if err = Pg2Domain(DTO(r, &dst.Refs[i])); err != nil {
			return dst, err
		}
	}
	return dst, nil
}

func hostEndpointsToDomain(src *pg.HostEndpoints) (dst *domain.HostEndpoints, err error) {
	if src == nil {
		return dst, err
	}
	dst = &domain.HostEndpoints{
		Address: src.Address,
	}
	if len(src.Ports) > 0 {
		dst.Ports = lo.Map(src.Ports, func(p pg.NamedPort, _ int) domain.NamedPort {
			return domain.NamedPort{
				Name: p.Name,
				Port: domain.PortNumber(p.Port), //nolint:gosec
			}
		})
	}

	return dst, nil
}

func hostInfoToDomain(src pg.HostInfo) (domain.HostInfo, error) {
	return domain.HostInfo{
		HostName:        src.HostName,
		OS:              src.OS,
		Platform:        src.Platform,
		PlatformFamily:  src.PlatformFamily,
		PlatformVersion: src.PlatformVersion,
		KernelVersion:   src.KernelVersion,
	}, nil
}

func hostToDomain(src pg.Host) (dst domain.Host, err error) {
	dst = domain.Host{
		Metadata: domain.ResMetadata{
			ID: domain.NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
					UID:  src.UID,
					Name: domain.ResourceName(src.Name),
				},
				Namespace: domain.ResourceNamespace(src.Namespace),
			},
			Labels:            src.Labels,
			Annotations:       src.Annotations,
			CreationTimestamp: src.CreationTimestamp,
			ResourceVersion:   src.ResourceVersion,
		},
		Spec: domain.HostSpec{
			CommonSpec: domain.CommonSpec{
				DisplayName: domain.DisplayName(src.DisplayName),
				Comment:     src.Comment,
				Description: src.Description,
			},
		},
		Refs: misc.Tern(len(src.Refs) > 0, make([]domain.ResourceRef, len(src.Refs)), nil),
	}
	for _, addr := range src.IPs {
		if addr.Is4() {
			dst.Spec.IPs.IPv4.Put(addr)
		} else {
			dst.Spec.IPs.IPv6.Put(addr)
		}
	}
	for i, r := range src.Refs {
		if err = Pg2Domain(DTO(r, &dst.Refs[i])); err != nil {
			return dst, err
		}
	}
	if err = Pg2Domain(DTO(src.MetaInfo, &dst.Spec.MetaInfo)); err == nil {
		err = Pg2Domain(DTO(src.Endpoints, &dst.Spec.Endpoints))
	}
	dst.Spec.Healthy = src.Healthy
	return dst, err
}

func resourceRefToDomain(src pg.ResourceRef) (domain.ResourceRef, error) {
	dst := domain.ResourceRef{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.Name),
			Namespace: domain.ResourceNamespace(src.Namespace),
		},
		ResType: domain.ResourceType(src.ResType),
	}
	return dst, nil
}

func nbEventToDomain(src pg.ResourceEvent) (dst domain.NetworkBindingEvent, err error) { //nolint:dupl
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	dst = domain.NetworkBindingEvent{
		TS:              src.TS,
		ResourceVersion: src.ResourceVersion,
		ResourceType:    domain.ResourceType(src.ResourceType),
	}
	if err = dst.EventType.FromString(src.EventType); err != nil {
		return dst, errors.WithMessage(err, "event type convertation")
	}
	var pgNetworkBinding pg.NetworkBinding
	err = json.Unmarshal(src.Object, &pgNetworkBinding)
	if err != nil {
		return dst, errors.WithMessage(err, "json unmarshal")
	}
	err = Pg2Domain(DTO(pgNetworkBinding, &dst.Object))
	return dst, err
}

func hbToDomain(src pg.HostBinding) (dst domain.HostBinding, err error) { //nolint:dupl
	dst = domain.HostBinding{
		Metadata: domain.ResMetadata{
			ID: domain.NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
					UID:  src.UID,
					Name: domain.ResourceName(src.Name),
				},
				Namespace: domain.ResourceNamespace(src.Namespace),
			},
			Labels:            src.Labels,
			Annotations:       src.Annotations,
			CreationTimestamp: src.CreationTimestamp,
			ResourceVersion:   src.ResourceVersion,
		},
		Spec: domain.HostBindingSpec{
			CommonSpec: domain.CommonSpec{
				DisplayName: domain.DisplayName(src.DisplayName),
				Comment:     src.Comment,
				Description: src.Description,
			},
		},
	}
	if err = Pg2Domain(DTO(src.AddressGroup, &dst.Spec.AddressGroup)); err != nil {
		return dst, err
	}
	err = Pg2Domain(DTO(src.Host, &dst.Spec.Host))
	return dst, err
}

func nbToDomain(src pg.NetworkBinding) (dst domain.NetworkBinding, err error) { //nolint:dupl
	dst = domain.NetworkBinding{
		Metadata: domain.ResMetadata{
			ID: domain.NamespacedMetadataIdentity{
				ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
					UID:  src.UID,
					Name: domain.ResourceName(src.Name),
				},
				Namespace: domain.ResourceNamespace(src.Namespace),
			},
			Labels:            src.Labels,
			Annotations:       src.Annotations,
			CreationTimestamp: src.CreationTimestamp,
			ResourceVersion:   src.ResourceVersion,
		},
		Spec: domain.NetworkBindingSpec{
			CommonSpec: domain.CommonSpec{
				DisplayName: domain.DisplayName(src.DisplayName),
				Comment:     src.Comment,
				Description: src.Description,
			},
		},
	}
	if err = Pg2Domain(DTO(src.AddressGroup, &dst.Spec.AddressGroup)); err != nil {
		return dst, err
	}
	err = Pg2Domain(DTO(src.Network, &dst.Spec.Network))
	return dst, err
}

func portRangeToDomain(src pg.PortRange) (dst domain.PortRange, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	if src.IsNull() {
		return nil, domain.ErrUnexpectedNullPortRange
	}
	lower, upper := src.Lower, src.Upper
	bounds := []struct {
		val *pg.PortNumber
		typ pgtype.BoundType
		adj pg.PortNumber
	}{
		{&lower, src.LowerType, +1},
		{&upper, src.UpperType, -1},
	}
	for _, n := range bounds {
		switch n.typ {
		case pgtype.Inclusive:
		case pgtype.Exclusive:
			*n.val += n.adj
		default:
			return nil, errors.Errorf("unexpected port range bound type '%s'", n.typ)
		}
	}
	if lower < 0 || lower > upper || upper > pg.PortNumber(^domain.PortNumber(0)) {
		return nil, errors.Errorf("invalid port range: %v - %v", lower, upper)
	}
	dst = domain.PortRangeFactory.Range(
		domain.PortNumber(lower), false, //nolint
		domain.PortNumber(upper), false, //nolint
	)
	err = domain.ValidatePortRange(dst, false)
	return dst, err
}

func portRangesToDomain(src pg.PortMultirange) (dst domain.PortRanges, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()

	dst = ranges.NewMultiRange(domain.PortRangeFactory)
	rr := make([]domain.PortRange, len(src.Multirange))
	for i, pr := range src.Multirange {
		if err = Pg2Domain(DTO(pr, &rr[i])); err != nil {
			return dst, err
		}
	}

	dst.Update(ranges.CombineMerge, rr...)
	return dst, nil
}

func icmpEntriesToDomain(src pg.IcmpEntries) (dst domain.IcmpEntry, err error) {
	dst = domain.IcmpEntry{
		Description: src.Description,
		Comment:     src.Comment,
	}
	for _, icmp := range src.Types {
		dst.Value.Put(uint8(icmp)) //nolint:gosec
	}
	return dst, nil
}

func portEntriesToDomain(src pg.PortEntries) (dst domain.PortEntry, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	dst = domain.PortEntry{
		Description: src.Description,
		Comment:     src.Comment,
	}

	err = Pg2Domain(DTO(src.Ports, &dst.Value))
	return dst, err
}

func rlToDomain(src pg.Rule) (dst domain.Rule, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	err = Pg2Domain(DTO(src.ResPK, &dst.Metadata.ID))
	if err != nil {
		return dst, err
	}

	dst.Spec.DisplayName = domain.DisplayName(src.DisplayName)
	dst.Spec.Comment = src.Comment
	dst.Spec.Description = src.Description
	dst.Metadata.Labels = src.Labels
	dst.Metadata.Annotations = src.Annotations
	if err = dst.Spec.Action.FromString(string(src.Action)); err != nil {
		return dst, err
	}
	err = dst.Spec.Traffic.FromString(string(src.Traffic))

	return dst, err
}

func uniRuleL4ToDomain(src pg.UniRuleL4) (dst domain.Rule, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	err = Pg2Domain(DTO(src.UniRule.Rule, &dst))
	if err != nil {
		return dst, err
	}

	if src.Proto == "" && src.IPv == "" && len(src.Entries) == 0 {
		dst.Spec.Transport = domain.NullTransport{}
		return dst, err
	}

	var l4 domain.L4Transport
	if err = l4.IPv.FromString(string(src.IPv)); err != nil {
		return dst, err
	}
	if err = l4.Proto.FromString(string(src.Proto)); err != nil {
		return dst, err
	}
	l4.Entries = misc.Tern(len(src.Entries) > 0, make([]domain.PortEntry, len(src.Entries)), nil)

	for i, e := range src.Entries {
		if err = Pg2Domain(DTO(e, &l4.Entries[i])); err != nil {
			return dst, err
		}
	}
	dst.Spec.Transport = l4

	return dst, err
}

func uniRuleIcmpToDomain(src pg.UniRuleIcmp) (dst domain.Rule, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	err = Pg2Domain(DTO(src.Rule, &dst))
	if err != nil {
		return dst, err
	}

	l4 := domain.IcmpTransport{Proto: domain.ICMP}
	if err = l4.IPv.FromString(string(src.IPv)); err != nil {
		return dst, err
	}

	l4.Entries = misc.Tern(len(src.Entries) > 0, make([]domain.IcmpEntry, len(src.Entries)), nil)

	for i, e := range src.Entries {
		if err = Pg2Domain(DTO(e, &l4.Entries[i])); err != nil {
			return dst, err
		}
	}
	dst.Spec.Transport = l4

	return dst, err
}

func res2resRuleToDomain(src pg.Res2ResRule) (dst domain.Rule, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	err = Pg2Domain(DTO(src.UniRuleL4, &dst))
	if err != nil {
		return dst, err
	}

	dst.Metadata.CreationTimestamp = src.CreationTimestamp
	dst.Metadata.ResourceVersion = src.ResourceVersion

	var ltype domain.EndpointType
	if err = ltype.FromString(string(src.Local.ResType)); err != nil {
		return dst, err
	}
	dst.Spec.Local = domain.EpLocal{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.Local.Name),
			Namespace: domain.ResourceNamespace(src.Local.Namespace),
		},
		Labels: src.Local.Labels,
		Type:   ltype,
	}

	var rtype domain.EndpointType
	if err = rtype.FromString(string(src.Remote.ResType)); err != nil {
		return dst, err
	}
	dst.Spec.Remote = domain.EpRemote{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.Remote.Name),
			Namespace: domain.ResourceNamespace(src.Remote.Namespace),
		},
		Labels: src.Remote.Labels,
		Type:   rtype,
	}

	return dst, err
}

func res2resIcmpRuleToDomain(src pg.Res2ResIcmpRule) (dst domain.Rule, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	err = Pg2Domain(DTO(src.UniRuleIcmp, &dst))
	if err != nil {
		return dst, err
	}

	dst.Metadata.CreationTimestamp = src.CreationTimestamp
	dst.Metadata.ResourceVersion = src.ResourceVersion

	var ltype domain.EndpointType
	if err = ltype.FromString(string(src.Local.ResType)); err != nil {
		return dst, err
	}
	var rtype domain.EndpointType
	if err = rtype.FromString(string(src.Remote.ResType)); err != nil {
		return dst, err
	}

	dst.Spec.Local = domain.EpLocal{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.Local.Name),
			Namespace: domain.ResourceNamespace(src.Local.Namespace),
		},
		Labels: src.Local.Labels,
		Type:   ltype,
	}

	dst.Spec.Remote = domain.EpRemote{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.Remote.Name),
			Namespace: domain.ResourceNamespace(src.Remote.Namespace),
		},
		Labels: src.Remote.Labels,
		Type:   rtype,
	}

	return dst, err
}

func res2IcmpRuleToDomain(src pg.Res2IcmpRule) (dst domain.Rule, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	err = Pg2Domain(DTO(src.UniRuleIcmp, &dst))
	if err != nil {
		return dst, err
	}

	dst.Metadata.CreationTimestamp = src.CreationTimestamp
	dst.Metadata.ResourceVersion = src.ResourceVersion

	var ltype domain.EndpointType
	if err = ltype.FromString(string(src.Local.ResType)); err != nil {
		return dst, err
	}

	dst.Spec.Local = domain.EpLocal{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.Local.Name),
			Namespace: domain.ResourceNamespace(src.Local.Namespace),
		},
		Labels: src.Local.Labels,
		Type:   ltype,
	}

	dst.Spec.Remote = domain.EpNull{}

	return dst, err
}

func res2CidrRuleToDomain(src pg.Res2CidrRule) (dst domain.Rule, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	err = Pg2Domain(DTO(src.UniRuleL4, &dst))
	if err != nil {
		return dst, err
	}

	dst.Metadata.CreationTimestamp = src.CreationTimestamp
	dst.Metadata.ResourceVersion = src.ResourceVersion

	var ltype domain.EndpointType
	if err = ltype.FromString(string(src.Local.ResType)); err != nil {
		return dst, err
	}

	dst.Spec.Local = domain.EpLocal{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.Local.Name),
			Namespace: domain.ResourceNamespace(src.Local.Namespace),
		},
		Labels: src.Local.Labels,
		Type:   ltype,
	}

	dst.Spec.Remote = domain.EpCIDR{
		Type:  domain.CidrEp,
		Value: domain.IPNet{IPNet: src.CIDR.IPNet},
	}

	return dst, err
}

func res2CidrIcmpRuleToDomain(src pg.Res2CidrIcmpRule) (dst domain.Rule, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	err = Pg2Domain(DTO(src.UniRuleIcmp, &dst))
	if err != nil {
		return dst, err
	}

	dst.Metadata.CreationTimestamp = src.CreationTimestamp
	dst.Metadata.ResourceVersion = src.ResourceVersion

	var ltype domain.EndpointType
	if err = ltype.FromString(string(src.Local.ResType)); err != nil {
		return dst, err
	}

	dst.Spec.Local = domain.EpLocal{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.Local.Name),
			Namespace: domain.ResourceNamespace(src.Local.Namespace),
		},
		Labels: src.Local.Labels,
		Type:   ltype,
	}

	dst.Spec.Remote = domain.EpCIDR{
		Type:  domain.CidrEp,
		Value: domain.IPNet{IPNet: src.CIDR.IPNet},
	}

	return dst, err
}

func res2FqdnRuleToDomain(src pg.Res2FqdnRule) (dst domain.Rule, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	err = Pg2Domain(DTO(src.UniRuleL4, &dst))
	if err != nil {
		return dst, err
	}

	dst.Metadata.CreationTimestamp = src.CreationTimestamp
	dst.Metadata.ResourceVersion = src.ResourceVersion

	var ltype domain.EndpointType
	if err = ltype.FromString(string(src.Local.ResType)); err != nil {
		return dst, err
	}

	dst.Spec.Local = domain.EpLocal{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.Local.Name),
			Namespace: domain.ResourceNamespace(src.Local.Namespace),
		},
		Labels: src.Local.Labels,
		Type:   ltype,
	}

	dst.Spec.Remote = domain.EpFQDN{
		Type:  domain.FqdnEp,
		Value: domain.FQDN(src.FQDN),
	}

	return dst, err
}

func transportEntryToPortEntry(src pg.TransportEntry) (dst domain.PortEntry, err error) {
	dst.Description = src.Description
	dst.Comment = src.Comment
	err = Pg2Domain(DTO(src.Ports, &dst.Value))

	return dst, err
}

func transportEntryToIcmpEntry(src pg.TransportEntry) (dst domain.IcmpEntry, err error) {
	dst.Description = src.Description
	dst.Comment = src.Comment
	for _, icmp := range src.IcmpTypes {
		dst.Value.Put(uint8(icmp)) //nolint:gosec
	}

	return dst, nil
}

func transportEntriesToPortEntries(src []pg.TransportEntry) ([]domain.PortEntry, error) {
	dst := misc.Tern(len(src) > 0, make([]domain.PortEntry, len(src)), nil)
	for i, e := range src {
		if err := Pg2Domain(DTO(e, &dst[i])); err != nil {
			return nil, err
		}
	}

	return dst, nil
}

func transportEntriesToIcmpEntries(src []pg.TransportEntry) ([]domain.IcmpEntry, error) {
	dst := misc.Tern(len(src) > 0, make([]domain.IcmpEntry, len(src)), nil)
	for i, e := range src {
		if err := Pg2Domain(DTO(e, &dst[i])); err != nil {
			return nil, err
		}
	}

	return dst, nil
}

func transportToDomain(src pg.Transport) (dst domain.TransportSpec, err error) {
	var proto domain.IPproto
	if err = proto.FromString(string(src.Proto)); err != nil {
		return nil, err
	}
	var ipv domain.IpFamily
	if string(src.IPv) != "" {
		if err = ipv.FromString(string(src.IPv)); err != nil {
			return nil, err
		}
	}

	switch proto {
	case domain.ICMP:
		t := domain.IcmpTransport{
			Proto: proto,
			IPv:   ipv,
		}
		if err = Pg2Domain(DTO(src.Entries, &t.Entries)); err != nil {
			return nil, err
		}
		dst = t
	case domain.TCP, domain.UDP:
		t := domain.L4Transport{
			Proto: proto,
			IPv:   ipv,
		}
		if err = Pg2Domain(DTO(src.Entries, &t.Entries)); err != nil {
			return nil, err
		}
		dst = t
	}

	return dst, nil
}

func serviceToDomain(src pg.Service) (dst domain.Service, err error) {
	if err = Pg2Domain(DTO(src.ResPK, &dst.Metadata.ID)); err != nil {
		return dst, err
	}

	dst.Metadata.Labels = src.Labels
	dst.Metadata.Annotations = src.Annotations
	dst.Metadata.CreationTimestamp = src.CreationTimestamp
	dst.Metadata.ResourceVersion = src.ResourceVersion
	dst.Spec.DisplayName = domain.DisplayName(src.DisplayName)
	dst.Spec.Comment = src.Comment
	dst.Spec.Description = src.Description
	dst.Spec.Transports = misc.Tern(len(src.Transports) > 0, make([]domain.TransportSpec, len(src.Transports)), nil)

	for i, tr := range src.Transports {
		if err = Pg2Domain(DTO(tr, &dst.Spec.Transports[i])); err != nil {
			return dst, err
		}
	}

	dst.Refs = misc.Tern(len(src.Refs) > 0, make([]domain.ResourceRef, len(src.Refs)), nil)
	for i, r := range src.Refs {
		if err = Pg2Domain(DTO(r, &dst.Refs[i])); err != nil {
			return dst, err
		}
	}

	return dst, nil
}

func serviceEventToDomain(src pg.ResourceEvent) (dst domain.ServiceEvent, err error) { //nolint:dupl
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	dst = domain.ServiceEvent{
		TS:              src.TS,
		ResourceVersion: src.ResourceVersion,
		ResourceType:    domain.ResourceType(src.ResourceType),
	}
	if err = dst.EventType.FromString(src.EventType); err != nil {
		return dst, errors.WithMessage(err, "event type convertation")
	}

	var pgService pg.Service
	err = json.Unmarshal(src.Object, &pgService)
	if err != nil {
		return dst, errors.WithMessage(err, "json unmarshal")
	}
	err = Pg2Domain(DTO(pgService, &dst.Object))

	return dst, err
}

func sbToDomain(src pg.ServiceBinding) (dst domain.ServiceBinding, err error) { //nolint:dupl
	if err = Pg2Domain(DTO(src.ResPK, &dst.Metadata.ID)); err != nil {
		return dst, err
	}

	dst.Metadata.Labels = src.Labels
	dst.Metadata.Annotations = src.Annotations
	dst.Metadata.CreationTimestamp = src.CreationTimestamp
	dst.Metadata.ResourceVersion = src.ResourceVersion
	dst.Spec.DisplayName = domain.DisplayName(src.DisplayName)
	dst.Spec.Comment = src.Comment
	dst.Spec.Description = src.Description

	if err = Pg2Domain(DTO(src.AddressGroup, &dst.Spec.AddressGroup)); err != nil {
		return dst, err
	}

	err = Pg2Domain(DTO(src.Service, &dst.Spec.Service))

	return dst, err
}

func sbEventToDomain(src pg.ResourceEvent) (dst domain.ServiceBindingEvent, err error) { //nolint:dupl
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dst)
	}()
	dst = domain.ServiceBindingEvent{
		TS:              src.TS,
		ResourceVersion: src.ResourceVersion,
		ResourceType:    domain.ResourceType(src.ResourceType),
	}
	if err = dst.EventType.FromString(src.EventType); err != nil {
		return dst, errors.WithMessage(err, "event type convertation")
	}
	var pgServiceBinding pg.ServiceBinding
	err = json.Unmarshal(src.Object, &pgServiceBinding)
	if err != nil {
		return dst, errors.WithMessage(err, "json unmarshal")
	}
	err = Pg2Domain(DTO(pgServiceBinding, &dst.Object))
	return dst, err
}

func rlEvtConv[pgT any](src pg.ResourceEvent, conv func(pgT) error) error {
	var pgRule pgT
	err := json.Unmarshal(src.Object, &pgRule)
	if err != nil {
		return errors.WithMessage(err, "json unmarshal")
	}
	return conv(pgRule)
}
