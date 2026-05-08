package dto

import (
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	dto.Register[domain.ResourceIdentifier, *common.ResourceIdentifier](resID2proto)
	dto.Register[domain.ResMetadata, *common.Metadata](metadata2proto)
	dto.Register[domain.ResourceRef, *common.ResourceRef](resRef2proto)
	dto.Register[domain.ResFieldSelector, *common.FieldSelector](fieldSelector2proto)
	dto.Register[domain.ResSelector, *common.ResSelector](resSelector2proto)
	dto.Register[domain.EpLocal, *common.Endpoints_Local](epLocal2proto)
	dto.Register[domain.EpRemote, *common.Endpoints_Remote](epRemote2proto)
	dto.Register[domain.EpCIDR, *common.Endpoints_Remote](epCIDR2proto)
	dto.Register[domain.EpFQDN, *common.Endpoints_Remote](epFQDN2proto)
	dto.Register[domain.EndpointSpec, *common.Endpoints_Remote](ep2proto)
	dto.Register[domain.IcmpEntry, *common.Transport_Entry](icmpEntry2proto)
	dto.Register[domain.PortEntry, *common.Transport_Entry](portEntry2proto)
	dto.Register[domain.IcmpTransport, *common.Transport](icmpTransport2proto)
	dto.Register[domain.L4Transport, *common.Transport](l4Transport2proto)
	dto.Register[domain.TransportSpec, *common.Transport](transport2proto)
	dto.Register[domain.IpFamily, common.IpAddrFamily](ipFamily2proto)
	dto.Register[domain.Traffic, common.Session_Traffic](traffic2proto)
	dto.Register[domain.IPproto, common.Transport_Protocol](ipProtocol2proto)
}

// Domain2Proto -
func Domain2Proto[T domain2protoVariants](a T) (err error) {
	return errors.WithMessage(a.Convert(), "common domain -> proto dto convertation")
}

// DTO -
func DTO[tFrom any, tTo any](a tFrom, b *tTo) *dto.Pair[tFrom, tTo] {
	return dto.MakePair(a, b)
}

type domain2protoVariants interface {
	*dto.Pair[domain.ResourceIdentifier, *common.ResourceIdentifier] |
		*dto.Pair[domain.ResMetadata, *common.Metadata] |
		*dto.Pair[domain.ResourceRef, *common.ResourceRef] |
		*dto.Pair[domain.ResFieldSelector, *common.FieldSelector] |
		*dto.Pair[domain.ResSelector, *common.ResSelector] |
		*dto.Pair[domain.EpLocal, *common.Endpoints_Local] |
		*dto.Pair[domain.EpRemote, *common.Endpoints_Remote] |
		*dto.Pair[domain.EpCIDR, *common.Endpoints_Remote] |
		*dto.Pair[domain.EpFQDN, *common.Endpoints_Remote] |
		*dto.Pair[domain.EndpointSpec, *common.Endpoints_Remote] |
		*dto.Pair[domain.IcmpEntry, *common.Transport_Entry] |
		*dto.Pair[domain.PortEntry, *common.Transport_Entry] |
		*dto.Pair[domain.IcmpTransport, *common.Transport] |
		*dto.Pair[domain.L4Transport, *common.Transport] |
		*dto.Pair[domain.TransportSpec, *common.Transport] |
		*dto.Pair[domain.IpFamily, common.IpAddrFamily] |
		*dto.Pair[domain.Traffic, common.Session_Traffic] |
		*dto.Pair[domain.IPproto, common.Transport_Protocol]
	Convert() error
}

func resID2proto(src domain.ResourceIdentifier) (dest *common.ResourceIdentifier, err error) {
	dest = &common.ResourceIdentifier{
		Name:      src.Name.String(),
		Namespace: src.Namespace.String(),
	}
	return dest, nil
}

func metadata2proto(src domain.ResMetadata) (dest *common.Metadata, err error) {
	dest = &common.Metadata{
		Uid:               src.ID.UID.String(),
		Name:              src.ID.Name.String(),
		Namespace:         src.ID.Namespace.String(),
		Labels:            src.Labels,
		Annotations:       src.Annotations,
		CreationTimestamp: timestamppb.New(src.CreationTimestamp),
		ResourceVersion:   src.ResourceVersion,
	}
	return dest, nil
}

func resRef2proto(src domain.ResourceRef) (dest *common.ResourceRef, err error) {
	dest = &common.ResourceRef{
		Name:      src.Name.String(),
		Namespace: src.Namespace.String(),
		ResType:   src.ResType.String(),
	}
	return dest, nil
}

func fieldSelector2proto(src domain.ResFieldSelector) (dest *common.FieldSelector, err error) {
	dest = &common.FieldSelector{
		Name:      src.Name.String(),
		Namespace: src.Namespace.String(),
		Refs:      make([]*common.ResourceRef, len(src.Refs)),
	}
	for i, r := range src.Refs {
		if err = Domain2Proto(DTO(r, &dest.Refs[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func resSelector2proto(src domain.ResSelector) (dest *common.ResSelector, err error) {
	dest = &common.ResSelector{
		LabelSelector: src.LabelSelector,
	}
	err = Domain2Proto(DTO(src.FieldSelector, &dest.FieldSelector))
	return dest, err
}

func epLocal2proto(src domain.EpLocal) (dest *common.Endpoints_Local, err error) {
	dest = &common.Endpoints_Local{
		Name:      src.Name.String(),
		Namespace: src.Namespace.String(),
		Type:      common.Endpoints_Type(src.Type),
		Labels:    src.Labels,
	}
	return dest, nil
}

func epRemote2proto(src domain.EpRemote) (dest *common.Endpoints_Remote, err error) {
	dest = &common.Endpoints_Remote{
		Name:      src.Name.String(),
		Namespace: src.Namespace.String(),
		Type:      common.Endpoints_Type(src.Type),
		Labels:    src.Labels,
	}
	return dest, nil
}

func epCIDR2proto(src domain.EpCIDR) (dest *common.Endpoints_Remote, err error) {
	dest = &common.Endpoints_Remote{
		Type:  common.Endpoints_Type(src.Type),
		Value: src.Value.String(),
	}
	return dest, nil
}

func epFQDN2proto(src domain.EpFQDN) (dest *common.Endpoints_Remote, err error) {
	dest = &common.Endpoints_Remote{
		Type:  common.Endpoints_Type(src.Type),
		Value: src.Value.String(),
	}
	return dest, nil
}

func ep2proto(src domain.EndpointSpec) (dest *common.Endpoints_Remote, err error) {
	if src == nil {
		return dest, err
	}
	switch remote := src.(type) {
	case domain.EpRemote:
		err = Domain2Proto(DTO(remote, &dest))
	case domain.EpCIDR:
		err = Domain2Proto(DTO(remote, &dest))
	case domain.EpFQDN:
		err = Domain2Proto(DTO(remote, &dest))
	case domain.EpNull:
	default:
		err = errors.Errorf("unsupported endpoint type '%T'", src)
	}
	return dest, err
}

func icmpEntry2proto(src domain.IcmpEntry) (dest *common.Transport_Entry, err error) {
	dest = &common.Transport_Entry{
		Description: src.Description,
		Comment:     src.Comment,
	}
	for v := range src.Value.Iterate {
		dest.Types = append(dest.Types, uint32(v))
	}
	return dest, nil
}

func portEntry2proto(src domain.PortEntry) (dest *common.Transport_Entry, err error) {
	var ps domain.PortSource
	if err = ps.FromPortRanges(src.Value); err != nil {
		return nil, err
	}
	dest = &common.Transport_Entry{
		Description: src.Description,
		Comment:     src.Comment,
		Ports:       string(ps),
	}

	return dest, nil
}

func icmpTransport2proto(src domain.IcmpTransport) (dest *common.Transport, err error) {
	dest = &common.Transport{
		Entries: misc.Tern(len(src.Entries) > 0, make([]*common.Transport_Entry, len(src.Entries)), nil),
	}
	if err = Domain2Proto(DTO(src.Proto, &dest.Protocol)); err != nil {
		return dest, err
	}
	if err = Domain2Proto(DTO(src.IPv, &dest.Ipv)); err != nil {
		return dest, err
	}
	for i, et := range src.Entries {
		if err = Domain2Proto(DTO(et, &dest.Entries[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func ipFamily2proto(ipv domain.IpFamily) (dst common.IpAddrFamily, err error) {
	switch ipv {
	case domain.IPv4:
		dst = common.IpAddrFamily_IPV4
	case domain.IPv6:
		dst = common.IpAddrFamily_IPV6
	default:
		err = errors.Errorf("unsupported domain ip family: %v", ipv)
	}
	return dst, err
}

func l4Transport2proto(src domain.L4Transport) (dest *common.Transport, err error) {
	dest = &common.Transport{
		Entries: misc.Tern(len(src.Entries) > 0, make([]*common.Transport_Entry, len(src.Entries)), nil),
	}
	if err = Domain2Proto(DTO(src.Proto, &dest.Protocol)); err != nil {
		return dest, err
	}
	if err = Domain2Proto(DTO(src.IPv, &dest.Ipv)); err != nil {
		return dest, err
	}
	for i, et := range src.Entries {
		if err = Domain2Proto(DTO(et, &dest.Entries[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func transport2proto(src domain.TransportSpec) (dest *common.Transport, err error) {
	switch transport := src.(type) {
	case nil, domain.NullTransport:
	case domain.IcmpTransport:
		err = Domain2Proto(DTO(transport, &dest))
	case domain.L4Transport:
		err = Domain2Proto(DTO(transport, &dest))
	default:
		err = errors.Errorf("unsupported transport type '%T'", src)
	}
	return dest, err
}

func traffic2proto(t domain.Traffic) (dst common.Session_Traffic, err error) {
	switch t {
	case domain.BOTH:
		dst = common.Session_BOTH
	case domain.INGRESS:
		dst = common.Session_INGRESS
	case domain.EGRESS:
		dst = common.Session_EGRESS
	default:
		err = errors.Errorf("unsupported domain traffic value: %v", t)
	}
	return dst, err
}

func ipProtocol2proto(p domain.IPproto) (dst common.Transport_Protocol, err error) {
	switch p {
	case domain.TCP:
		dst = common.Transport_TCP
	case domain.UDP:
		dst = common.Transport_UDP
	case domain.ICMP:
		dst = common.Transport_ICMP
	default:
		err = errors.Errorf("unsupported domain protocol value: %v", p)
	}
	return dst, err
}
