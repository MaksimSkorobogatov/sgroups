package dto

import (
	"net/netip"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	cdto "github.com/PRO-Robotech/sgroups/internal/shared/dto/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
)

func init() {
	dto.Register[*pb.Host_Spec_MetaInfo, domain.HostInfo](hostInfoToDomain)
	dto.Register[*pb.Host_Spec, domain.HostSpec](specToDomain)
	dto.Register[*pb.Host, domain.Host](hostToDomain)
	dto.Register[*pb.HostReq_Upsert, domain.Hosts](upsertReqToDomain)
	dto.Register[*pb.HostReq_List, domain.ResSelectorList](hostListToDomain)
	dto.Register[*pb.HostReq_Delete_Host, domain.Host](deleteReqHostToDomain)
	dto.Register[*pb.HostReq_Delete, domain.Hosts](deleteReqToDomain)

	dto.Register[*pb.Host_Spec_Endpoints_Port, domain.NamedPort](hostPortToDomain)
	dto.Register[*pb.Host_Spec_Endpoints, *domain.HostEndpoints](hostEndpointsToDomain)
	dto.Register[*pb.HostReq_UpdIPs_Host_Spec, domain.HostSpec](updIpSpecToDomain)
	dto.Register[*pb.HostReq_UpdIPs_Host, domain.Host](updIpHostToDomain)
	dto.Register[*pb.HostReq_UpdIPs, domain.Hosts](updIpToDomain)

	dto.Register[*pb.HostReq_UpdMetaInfo_HostInfo_Spec, domain.HostSpec](updMetaInfoSpecToDomain)
	dto.Register[*pb.HostReq_UpdMetaInfo_HostInfo, domain.Host](updMetaInfoHostToDomain)
	dto.Register[*pb.HostReq_UpdMetaInfo, domain.Hosts](updMetaInfoToDomain)

	// TODO: Раскоментировать после создания типов proto:
	// dto.Register[*pb.HostReq_UpdHealthStatus_Host_Spec, domain.HostSpec](updHealthSpecToDomain)
	// dto.Register[*pb.HostReq_UpdHealthStatus_Host, domain.Host](updHealthHostToDomain)
	// dto.Register[*pb.HostReq_UpdHealthStatus, domain.Hosts](updHealthToDomain)

	dto.Register[*pb.HostReq_SocketStatistics_FieldSelector, domain.ResSelector](ssSelectorToDomain)
	dto.Register[*pb.HostReq_SocketStatistics_List, domain.ResSelectorList](ssListReqToDomain)
	dto.Register[*pb.HostReq_SocketStatistics_Watch, domain.ResSelectorList](ssWatchReqToDomain)
	dto.Register[*pb.HostReq_Nft_FieldSelector, domain.ResSelector](nftSelectorToDomain)
	dto.Register[*pb.HostReq_Nft_List, domain.ResSelectorList](nftListReqToDomain)
	dto.Register[*pb.HostReq_Nft_Watch, domain.ResSelectorList](nftWatchReqToDomain)
}

// Proto2Domain -
func Proto2Domain[v proto2domainVariants](obj v) error {
	return errors.WithMessage(obj.Convert(), "proto -> domain dto convertation")
}

// DTO -
func DTO[tFrom any, tTo any](a tFrom, b *tTo) *dto.Pair[tFrom, tTo] {
	return dto.MakePair(a, b)
}

type proto2domainVariants interface {
	*dto.Pair[*pb.Host_Spec_MetaInfo, domain.HostInfo] |
		*dto.Pair[*pb.Host_Spec, domain.HostSpec] |
		*dto.Pair[*pb.Host, domain.Host] |
		*dto.Pair[*pb.HostReq_Upsert, domain.Hosts] |
		*dto.Pair[*pb.HostReq_List, domain.ResSelectorList] |
		*dto.Pair[*pb.HostReq_Delete_Host, domain.Host] |
		*dto.Pair[*pb.HostReq_Delete, domain.Hosts] |

		*dto.Pair[*pb.Host_Spec_Endpoints_Port, domain.NamedPort] |
		*dto.Pair[*pb.Host_Spec_Endpoints, *domain.HostEndpoints] |
		*dto.Pair[*pb.HostReq_UpdIPs_Host_Spec, domain.HostSpec] |
		*dto.Pair[*pb.HostReq_UpdIPs_Host, domain.Host] |
		*dto.Pair[*pb.HostReq_UpdIPs, domain.Hosts] |

		*dto.Pair[*pb.HostReq_UpdMetaInfo_HostInfo_Spec, domain.HostSpec] |
		*dto.Pair[*pb.HostReq_UpdMetaInfo_HostInfo, domain.Host] |
		*dto.Pair[*pb.HostReq_UpdMetaInfo, domain.Hosts] |
		*dto.Pair[*pb.HostReq_SocketStatistics_FieldSelector, domain.ResSelector] |
		*dto.Pair[*pb.HostReq_SocketStatistics_List, domain.ResSelectorList] |
		*dto.Pair[*pb.HostReq_SocketStatistics_Watch, domain.ResSelectorList] |
		*dto.Pair[*pb.HostReq_Nft_FieldSelector, domain.ResSelector] |
		*dto.Pair[*pb.HostReq_Nft_List, domain.ResSelectorList] |
		*dto.Pair[*pb.HostReq_Nft_Watch, domain.ResSelectorList]

	Convert() error
}

func hostInfoToDomain(src *pb.Host_Spec_MetaInfo) (dest domain.HostInfo, err error) {
	dest = domain.HostInfo{
		HostName:        src.GetHostName(),
		OS:              src.GetOs(),
		Platform:        src.GetPlatform(),
		PlatformFamily:  src.GetPlatformFamily(),
		PlatformVersion: src.GetPlatformVersion(),
		KernelVersion:   src.GetKernelVersion(),
	}
	return dest, nil
}

func specToDomain(src *pb.Host_Spec) (dest domain.HostSpec, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = domain.HostSpec{
		CommonSpec: domain.CommonSpec{
			DisplayName: domain.DisplayName(src.GetDisplayName()),
			Comment:     src.GetComment(),
			Description: src.GetDescription(),
		},
	}

	for _, ip := range src.GetIps().GetIpv4() {
		var addr netip.Addr
		if addr, err = netip.ParseAddr(ip); err != nil {
			return dest, err
		}
		dest.IPs.IPv4.Put(addr)
	}
	for _, ip := range src.GetIps().GetIpv6() {
		var addr netip.Addr
		if addr, err = netip.ParseAddr(ip); err != nil {
			return dest, err
		}
		dest.IPs.IPv6.Put(addr)
	}

	err = Proto2Domain(DTO(src.GetMetaInfo(), &dest.MetaInfo))

	return dest, err
}

func hostToDomain(src *pb.Host) (dest domain.Host, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	err = cdto.Proto2Domain(cdto.DTO(src.GetMetadata(), &dest.Metadata))
	if err != nil {
		return dest, err
	}
	err = Proto2Domain(DTO(src.GetSpec(), &dest.Spec))
	return dest, err
}

func upsertReqToDomain(src *pb.HostReq_Upsert) (dest domain.Hosts, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetHosts()) > 0, make(domain.Hosts, len(src.GetHosts())), nil)
	for i, host := range src.GetHosts() {
		if err = Proto2Domain(DTO(host, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func hostPortToDomain(src *pb.Host_Spec_Endpoints_Port) (dest domain.NamedPort, err error) {
	return domain.NamedPort{
		Name: src.GetName(),
		Port: domain.PortNumber(src.GetPort()), //nolint:gosec
	}, nil
}

func hostEndpointsToDomain(src *pb.Host_Spec_Endpoints) (dest *domain.HostEndpoints, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	if src == nil {
		return dest, err
	}
	dest = new(domain.HostEndpoints)
	if addr := src.GetAddress(); len(addr) > 0 {
		if dest.Address, err = netip.ParseAddr(addr); err != nil {
			return dest, err
		}
	}
	dest.Ports = misc.Tern(len(src.GetPorts()) > 0, make([]domain.NamedPort, len(src.GetPorts())), nil)
	for i, port := range src.GetPorts() {
		if err = Proto2Domain(DTO(port, &dest.Ports[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func updIpSpecToDomain(src *pb.HostReq_UpdIPs_Host_Spec) (dest domain.HostSpec, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()

	for _, ip := range src.GetIps().GetIpv4() {
		var addr netip.Addr
		if addr, err = netip.ParseAddr(ip); err != nil {
			return dest, err
		}
		dest.IPs.IPv4.Put(addr)
	}
	for _, ip := range src.GetIps().GetIpv6() {
		var addr netip.Addr
		if addr, err = netip.ParseAddr(ip); err != nil {
			return dest, err
		}
		dest.IPs.IPv6.Put(addr)
	}
	err = Proto2Domain(DTO(src.GetEndpoints(), &dest.Endpoints))
	return dest, err
}

func updIpHostToDomain(src *pb.HostReq_UpdIPs_Host) (dest domain.Host, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	err = Proto2Domain(DTO(src.GetSpec(), &dest.Spec))
	if err != nil {
		return dest, err
	}
	err = cdto.Proto2Domain(cdto.DTO(src.GetMetadata(), &dest.Metadata))
	return dest, err
}

func updIpToDomain(src *pb.HostReq_UpdIPs) (dest domain.Hosts, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetHosts()) > 0, make(domain.Hosts, len(src.GetHosts())), nil)
	for i, host := range src.GetHosts() {
		if err = Proto2Domain(DTO(host, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func updMetaInfoSpecToDomain(src *pb.HostReq_UpdMetaInfo_HostInfo_Spec) (dest domain.HostSpec, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()

	err = Proto2Domain(DTO(src.GetMetaInfo(), &dest.MetaInfo))
	return dest, err
}

func updMetaInfoHostToDomain(src *pb.HostReq_UpdMetaInfo_HostInfo) (dest domain.Host, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	err = Proto2Domain(DTO(src.GetSpec(), &dest.Spec))
	if err != nil {
		return dest, err
	}
	err = cdto.Proto2Domain(cdto.DTO(src.GetMetadata(), &dest.Metadata))
	return dest, err
}

func updMetaInfoToDomain(src *pb.HostReq_UpdMetaInfo) (dest domain.Hosts, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetHosts()) > 0, make(domain.Hosts, len(src.GetHosts())), nil)
	for i, host := range src.GetHosts() {
		if err = Proto2Domain(DTO(host, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func hostListToDomain(src *pb.HostReq_List) (dest domain.ResSelectorList, err error) {
	dest = misc.Tern(len(src.GetSelectors()) > 0, make(domain.ResSelectorList, len(src.GetSelectors())), nil)
	for i, sel := range src.GetSelectors() {
		if err = cdto.Proto2Domain(cdto.DTO(sel, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func deleteReqHostToDomain(src *pb.HostReq_Delete_Host) (dest domain.Host, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	err = cdto.Proto2Domain(cdto.DTO(src.GetMetadata(), &dest.Metadata))
	return dest, err
}

func deleteReqToDomain(src *pb.HostReq_Delete) (dest domain.Hosts, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetHosts()) > 0, make(domain.Hosts, len(src.GetHosts())), nil)
	for i, host := range src.GetHosts() {
		if err = Proto2Domain(DTO(host, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func ssSelectorToDomain(src *pb.HostReq_SocketStatistics_FieldSelector) (dest domain.ResSelector, err error) {
	return domain.ResSelector{
		FieldSelector: domain.ResFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name:      domain.ResourceName(src.GetName()),
				Namespace: domain.ResourceNamespace(src.GetNamespace()),
			},
		},
	}, nil
}

func ssListReqToDomain(src *pb.HostReq_SocketStatistics_List) (dest domain.ResSelectorList, err error) {
	dest = misc.Tern(len(src.GetSelectors()) > 0, make(domain.ResSelectorList, len(src.GetSelectors())), nil)
	for i, sel := range src.GetSelectors() {
		if err = Proto2Domain(DTO(sel, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func ssWatchReqToDomain(src *pb.HostReq_SocketStatistics_Watch) (dest domain.ResSelectorList, err error) {
	dest = misc.Tern(len(src.GetSelectors()) > 0, make(domain.ResSelectorList, len(src.GetSelectors())), nil)
	for i, sel := range src.GetSelectors() {
		if err = Proto2Domain(DTO(sel, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func nftSelectorToDomain(src *pb.HostReq_Nft_FieldSelector) (dest domain.ResSelector, err error) {
	return domain.ResSelector{
		FieldSelector: domain.ResFieldSelector{
			ResourceIdentifier: domain.ResourceIdentifier{
				Name:      domain.ResourceName(src.GetName()),
				Namespace: domain.ResourceNamespace(src.GetNamespace()),
			},
		},
	}, nil
}

func nftListReqToDomain(src *pb.HostReq_Nft_List) (dest domain.ResSelectorList, err error) {
	dest = misc.Tern(len(src.GetSelectors()) > 0, make(domain.ResSelectorList, len(src.GetSelectors())), nil)
	for i, sel := range src.GetSelectors() {
		if err = Proto2Domain(DTO(sel, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func nftWatchReqToDomain(src *pb.HostReq_Nft_Watch) (dest domain.ResSelectorList, err error) {
	dest = misc.Tern(len(src.GetSelectors()) > 0, make(domain.ResSelectorList, len(src.GetSelectors())), nil)
	for i, sel := range src.GetSelectors() {
		if err = Proto2Domain(DTO(sel, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

// TODO: Раскоментировать после добавления proto.
//
// func updHealthSpecToDomain(src *pb.HostReq_UpdHealthStatus_Host_Spec) (dest domain.HostSpec, err error) {
// 	defer func() {
// 		err = errors.WithMessagef(err, "%T -> %T", src, dest)
// 	}()
// 	v := src.GetHealthy()
// 	dest.HealthStatus = &v
// 	return dest, err
// }
//
// func updHealthHostToDomain(src *pb.HostReq_UpdHealthStatus_Host) (dest domain.Host, err error) {
// 	defer func() {
// 		err = errors.WithMessagef(err, "%T -> %T", src, dest)
// 	}()
// 	err = Proto2Domain(DTO(src.GetSpec(), &dest.Spec))
// 	if err != nil {
// 		return dest, err
// 	}
// 	err = cdto.Proto2Domain(cdto.DTO(src.GetMetadata(), &dest.Metadata))
// 	return dest, err
// }
//
// func updHealthToDomain(src *pb.HostReq_UpdHealthStatus) (dest domain.Hosts, err error) {
// 	defer func() {
// 		err = errors.WithMessagef(err, "%T -> %T", src, dest)
// 	}()
// 	dest = misc.Tern(len(src.GetHosts()) > 0, make(domain.Hosts, len(src.GetHosts())), nil)
// 	for i, host := range src.GetHosts() {
// 		if err = Proto2Domain(DTO(host, &dest[i])); err != nil {
// 			return dest, err
// 		}
// 	}
// 	return dest, nil
// }
