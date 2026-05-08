package dto

import (
	"net"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func init() {
	dto.Register[*common.ResourceIdentifier, domain.ResourceIdentifier](resID2domain)
	dto.Register[*common.Metadata, domain.ResMetadata](metadata2domain)
	dto.Register[*common.MetadataScope, domain.ResMetadata](metadataScope2domain)
	dto.Register[*common.ResourceRef, domain.ResourceRef](resRef2domain)
	dto.Register[*common.FieldSelector, domain.ResFieldSelector](fieldSelector2domain)
	dto.Register[*common.ResSelector, domain.ResSelector](resSelector2domain)
	dto.Register[*common.Endpoints_Local, domain.EpLocal](epLocal2domain)
	dto.Register[*common.Endpoints_Remote, domain.EpRemote](epRemote2domain)
	dto.Register[*common.Endpoints_Remote, domain.EpCIDR](epCIDR2domain)
	dto.Register[*common.Endpoints_Remote, domain.EpFQDN](epFQDN2domain)
	dto.Register[*common.Endpoints_Remote, domain.EndpointSpec](ep2domain)
	dto.Register[*common.Transport_Entry, domain.IcmpEntry](icmpEntry2domain)
	dto.Register[*common.Transport_Entry, domain.PortEntry](portEntry2domain)
	dto.Register[*common.Transport, domain.IcmpTransport](icmpTransport2domain)
	dto.Register[*common.Transport, domain.L4Transport](l4Transport2domain)
	dto.Register[*common.Transport, domain.TransportSpec](transport2domain)
	dto.Register[common.IpAddrFamily, domain.IpFamily](ipFamily2domain)
	dto.Register[common.Session_Traffic, domain.Traffic](traffic2domain)
	dto.Register[common.Transport_Protocol, domain.IPproto](ipProtocol2domain)
}

// Proto2Domain -
func Proto2Domain[T proto2domainVariants](a T) (err error) {
	return errors.WithMessage(a.Convert(), "common proto -> domain dto convertation")
}

type proto2domainVariants interface {
	*dto.Pair[*common.ResourceIdentifier, domain.ResourceIdentifier] |
		*dto.Pair[*common.Metadata, domain.ResMetadata] |
		*dto.Pair[*common.MetadataScope, domain.ResMetadata] |
		*dto.Pair[*common.ResourceRef, domain.ResourceRef] |
		*dto.Pair[*common.FieldSelector, domain.ResFieldSelector] |
		*dto.Pair[*common.ResSelector, domain.ResSelector] |
		*dto.Pair[*common.Endpoints_Local, domain.EpLocal] |
		*dto.Pair[*common.Endpoints_Remote, domain.EpRemote] |
		*dto.Pair[*common.Endpoints_Remote, domain.EpCIDR] |
		*dto.Pair[*common.Endpoints_Remote, domain.EpFQDN] |
		*dto.Pair[*common.Endpoints_Remote, domain.EndpointSpec] |
		*dto.Pair[*common.Transport_Entry, domain.IcmpEntry] |
		*dto.Pair[*common.Transport_Entry, domain.PortEntry] |
		*dto.Pair[*common.Transport, domain.IcmpTransport] |
		*dto.Pair[*common.Transport, domain.L4Transport] |
		*dto.Pair[*common.Transport, domain.TransportSpec] |
		*dto.Pair[common.IpAddrFamily, domain.IpFamily] |
		*dto.Pair[common.Session_Traffic, domain.Traffic] |
		*dto.Pair[common.Transport_Protocol, domain.IPproto]
	Convert() error
}

func resID2domain(src *common.ResourceIdentifier) (dest domain.ResourceIdentifier, err error) {
	dest = domain.ResourceIdentifier{
		Name:      domain.ResourceName(src.GetName()),
		Namespace: domain.ResourceNamespace(src.GetNamespace()),
	}
	return dest, nil
}

func metadata2domain(src *common.Metadata) (dest domain.ResMetadata, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = domain.ResMetadata{
		ID: domain.NamespacedMetadataIdentity{
			ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
				Name: domain.ResourceName(src.GetName()),
			},
			Namespace: domain.ResourceNamespace(src.GetNamespace()),
		},
		Labels:            src.GetLabels(),
		Annotations:       src.GetAnnotations(),
		CreationTimestamp: src.GetCreationTimestamp().AsTime(),
		ResourceVersion:   src.GetResourceVersion(),
	}
	if len(src.GetUid()) > 0 {
		if dest.ID.UID, err = uuid.Parse(src.GetUid()); err != nil {
			err = errors.WithMessagef(err, "bad 'UUID' '%s'", src.GetUid())
		}
	}

	return dest, err
}

func metadataScope2domain(src *common.MetadataScope) (dest domain.ResMetadata, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = domain.ResMetadata{
		ID: domain.NamespacedMetadataIdentity{
			ClusterScopeMetadataIdentity: domain.ClusterScopeMetadataIdentity{
				Name: domain.ResourceName(src.GetName()),
			},
			Namespace: domain.ResourceNamespace(src.GetNamespace()),
		},
	}
	if len(src.GetUid()) > 0 {
		if dest.ID.UID, err = uuid.Parse(src.GetUid()); err != nil {
			err = errors.WithMessagef(err, "bad 'UUID' '%s'", src.GetUid())
		}
	}

	return dest, err
}

func resRef2domain(src *common.ResourceRef) (dest domain.ResourceRef, err error) {
	dest = domain.ResourceRef{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.GetName()),
			Namespace: domain.ResourceNamespace(src.GetNamespace()),
		},
		ResType: domain.ResourceType(src.GetResType()),
	}
	return dest, nil
}

func fieldSelector2domain(src *common.FieldSelector) (dest domain.ResFieldSelector, err error) {
	dest = domain.ResFieldSelector{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.GetName()),
			Namespace: domain.ResourceNamespace(src.GetNamespace()),
		},
	}
	dest.Refs = misc.Tern(len(src.GetRefs()) > 0, make([]domain.ResourceRef, len(src.GetRefs())), nil)
	for i, r := range src.GetRefs() {
		if err = Proto2Domain(DTO(r, &dest.Refs[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func resSelector2domain(src *common.ResSelector) (dest domain.ResSelector, err error) {
	dest = domain.ResSelector{
		LabelSelector: src.GetLabelSelector(),
	}
	err = Proto2Domain(DTO(src.GetFieldSelector(), &dest.FieldSelector))
	return dest, err
}

func epLocal2domain(src *common.Endpoints_Local) (dest domain.EpLocal, err error) {
	dest = domain.EpLocal{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.GetName()),
			Namespace: domain.ResourceNamespace(src.GetNamespace()),
		},
		Type:   domain.EndpointType(src.GetType()), //nolint:gosec
		Labels: src.GetLabels(),
	}
	return dest, nil
}

func epRemote2domain(src *common.Endpoints_Remote) (dest domain.EpRemote, err error) {
	dest = domain.EpRemote{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.GetName()),
			Namespace: domain.ResourceNamespace(src.GetNamespace()),
		},
		Type:   domain.EndpointType(src.GetType()), //nolint:gosec
		Labels: src.GetLabels(),
	}
	return dest, nil
}

func epCIDR2domain(src *common.Endpoints_Remote) (dest domain.EpCIDR, err error) {
	dest = domain.EpCIDR{
		Type: domain.EndpointType(src.GetType()), //nolint:gosec
	}
	ip, ipnet, e := net.ParseCIDR(src.GetValue())
	if e != nil {
		return dest, errors.WithMessagef(e, "bad CIDR '%s'", src.GetValue())
	}

	origIP := ip.To4()
	if origIP == nil {
		origIP = ip.To16()
	}
	dest.Value = domain.IPNet{IPNet: net.IPNet{IP: origIP, Mask: ipnet.Mask}}
	return dest, nil
}

func epFQDN2domain(src *common.Endpoints_Remote) (dest domain.EpFQDN, err error) {
	dest = domain.EpFQDN{
		Type:  domain.EndpointType(src.GetType()), //nolint:gosec
		Value: domain.FQDN(src.GetValue()),
	}

	return dest, nil
}

func ep2domain(src *common.Endpoints_Remote) (dst domain.EndpointSpec, err error) {
	if src == nil {
		return domain.EpNull{}, nil
	}
	epType := domain.EndpointType(src.GetType()) //nolint:gosec
	switch epType {
	case domain.AddressGroupEp, domain.ServiceEp:
		var remote domain.EpRemote
		err = Proto2Domain(DTO(src, &remote))
		if err != nil {
			return dst, err
		}
		dst = remote
	case domain.CidrEp:
		var remote domain.EpCIDR
		err = Proto2Domain(DTO(src, &remote))
		if err != nil {
			return dst, err
		}
		dst = remote
	case domain.FqdnEp:
		var remote domain.EpFQDN
		err = Proto2Domain(DTO(src, &remote))
		if err != nil {
			return dst, err
		}
		dst = remote
	default:
		dst = domain.EpNull{}
	}

	return dst, nil
}

func icmpEntry2domain(src *common.Transport_Entry) (dest domain.IcmpEntry, err error) {
	dest = domain.IcmpEntry{
		Description: src.GetDescription(),
		Comment:     src.GetComment(),
	}
	for _, v := range src.GetTypes() {
		dest.Value.Put(uint8(v)) //nolint:gosec
	}
	return dest, nil
}

func portEntry2domain(src *common.Transport_Entry) (dest domain.PortEntry, err error) {
	dest = domain.PortEntry{
		Description: src.GetDescription(),
		Comment:     src.GetComment(),
	}
	if dest.Value, err = domain.PortSource(src.GetPorts()).ToPortRanges(); err != nil {
		err = errors.Errorf( //TODO: fix corlib error message
			"invalid destination port range %q: value out of range", src.GetPorts())
	}
	return dest, err
}

func traffic2domain(t common.Session_Traffic) (dst domain.Traffic, err error) {
	switch t {
	case common.Session_BOTH:
		dst = domain.BOTH
	case common.Session_INGRESS:
		dst = domain.INGRESS
	case common.Session_EGRESS:
		dst = domain.EGRESS
	case common.Session_TRAFFIC_UNDEF:
		err = errors.New("traffic must be set")
	default:
		err = errors.Errorf("unsupported traffic: %v", t)
	}
	return dst, err
}

func ipProtocol2domain(p common.Transport_Protocol) (dst domain.IPproto, err error) {
	switch p {
	case common.Transport_TCP:
		dst = domain.TCP
	case common.Transport_UDP:
		dst = domain.UDP
	case common.Transport_ICMP:
		dst = domain.ICMP
	default:
		err = errors.Errorf("unsupported protocol: %v", p)
	}
	return dst, err
}

func icmpTransport2domain(src *common.Transport) (dest domain.IcmpTransport, err error) {
	dest = domain.IcmpTransport{
		Entries: misc.Tern(len(src.GetEntries()) > 0, make([]domain.IcmpEntry, len(src.GetEntries())), nil),
	}
	if err = Proto2Domain(DTO(src.GetProtocol(), &dest.Proto)); err != nil {
		return dest, err
	}
	if err = Proto2Domain(DTO(src.GetIpv(), &dest.IPv)); err != nil {
		return dest, err
	}
	for i, et := range src.GetEntries() {
		if err = Proto2Domain(DTO(et, &dest.Entries[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func l4Transport2domain(src *common.Transport) (dest domain.L4Transport, err error) {
	dest = domain.L4Transport{
		Entries: misc.Tern(len(src.GetEntries()) > 0, make([]domain.PortEntry, len(src.GetEntries())), nil),
	}
	if err = Proto2Domain(DTO(src.GetProtocol(), &dest.Proto)); err != nil {
		return dest, err
	}
	if err = Proto2Domain(DTO(src.GetIpv(), &dest.IPv)); err != nil {
		return dest, err
	}
	for i, et := range src.GetEntries() {
		if err = Proto2Domain(DTO(et, &dest.Entries[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func transport2domain(src *common.Transport) (dest domain.TransportSpec, err error) {
	if src == nil {
		return domain.NullTransport{}, nil
	}
	var proto domain.IPproto
	if err = Proto2Domain(DTO(src.GetProtocol(), &proto)); err != nil {
		return dest, err
	}
	switch proto {
	case domain.ICMP:
		var transport domain.IcmpTransport
		err = Proto2Domain(DTO(src, &transport))
		if err != nil {
			return dest, err
		}
		dest = transport
	case domain.TCP, domain.UDP:
		var transport domain.L4Transport
		err = Proto2Domain(DTO(src, &transport))
		if err != nil {
			return dest, err
		}
		dest = transport
	default:
		err = errors.WithMessagef(err, "unsupported protocol '%s'", proto)
	}
	return dest, err
}

func ipFamily2domain(ipv common.IpAddrFamily) (dst domain.IpFamily, err error) {
	switch ipv {
	case common.IpAddrFamily_IPV4:
		dst = domain.IPv4
	case common.IpAddrFamily_IPV6:
		dst = domain.IPv6
	default:
		err = errors.Errorf("unsupported ip family: %v", ipv)
	}
	return dst, err
}
