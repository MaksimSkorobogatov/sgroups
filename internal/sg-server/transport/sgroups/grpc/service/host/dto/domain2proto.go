package dto

import (
	"slices"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	cdto "github.com/PRO-Robotech/sgroups/internal/shared/dto/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
)

func init() {
	dto.Register[domain.NamedPort, *pb.Host_Spec_Endpoints_Port](hostPortToProto)
	dto.Register[*domain.HostEndpoints, *pb.Host_Spec_Endpoints](hostEndpointsToProto)
	dto.Register[domain.HostInfo, *pb.Host_Spec_MetaInfo](hostInfoToProto)
	dto.Register[domain.HostSpec, *pb.Host_Spec](specToProto)
	dto.Register[domain.Host, *pb.Host](hostToProto)
	dto.Register[domain.Hosts, *pb.HostResp_Upsert](hostsToProto)
	dto.Register[domain.Host, *pb.HostResp_HostExt](hostExtToProto)
	dto.Register[domain.HostList, *pb.HostResp_List](hostListToProto)
	dto.Register[domain.HostEvent, *pb.HostResp_Watch](hostEventToProto)
	dto.Register[domain.Hosts, *pb.HostResp_UpdIPs](updIpToProto)
	dto.Register[domain.Hosts, *pb.HostResp_UpdMetaInfo](updMetaInfoToProto)

	// TODO: Раскоментировать после добавления proto:
	// dto.Register[domain.Hosts, *pb.HostResp_UpdHealthStatus](updHealthToProto)
}

// Domain2Proto -
func Domain2Proto[v domain2protoVariants](obj v) error {
	return errors.WithMessage(obj.Convert(), "domain -> proto dto convertation")
}

type domain2protoVariants interface {
	*dto.Pair[domain.NamedPort, *pb.Host_Spec_Endpoints_Port] |
		*dto.Pair[*domain.HostEndpoints, *pb.Host_Spec_Endpoints] |
		*dto.Pair[domain.HostInfo, *pb.Host_Spec_MetaInfo] |
		*dto.Pair[domain.HostSpec, *pb.Host_Spec] |
		*dto.Pair[domain.Host, *pb.Host] |
		*dto.Pair[domain.Hosts, *pb.HostResp_Upsert] |
		*dto.Pair[domain.Host, *pb.HostResp_HostExt] |
		*dto.Pair[domain.HostList, *pb.HostResp_List] |
		*dto.Pair[domain.HostEvent, *pb.HostResp_Watch] |
		*dto.Pair[domain.Hosts, *pb.HostResp_UpdIPs] |
		*dto.Pair[domain.Hosts, *pb.HostResp_UpdMetaInfo]
	Convert() error
}

func hostPortToProto(src domain.NamedPort) (dest *pb.Host_Spec_Endpoints_Port, err error) {
	return &pb.Host_Spec_Endpoints_Port{
		Name: src.Name,
		Port: uint32(src.Port),
	}, nil
}

func hostEndpointsToProto(src *domain.HostEndpoints) (dest *pb.Host_Spec_Endpoints, err error) {
	dest = new(pb.Host_Spec_Endpoints)
	if src == nil {
		return dest, err
	}
	dest.Address = src.Address.String()
	dest.Ports = misc.Tern(len(src.Ports) > 0, make([]*pb.Host_Spec_Endpoints_Port, len(src.Ports)), nil)
	for i, port := range src.Ports {
		if err = Domain2Proto(DTO(port, &dest.Ports[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func hostInfoToProto(src domain.HostInfo) (dest *pb.Host_Spec_MetaInfo, err error) {
	dest = &pb.Host_Spec_MetaInfo{
		HostName:        src.HostName,
		Os:              src.OS,
		Platform:        src.Platform,
		PlatformFamily:  src.PlatformFamily,
		PlatformVersion: src.PlatformVersion,
		KernelVersion:   src.KernelVersion,
	}
	return dest, nil
}

func specToProto(src domain.HostSpec) (dest *pb.Host_Spec, err error) {
	dest = &pb.Host_Spec{
		DisplayName: src.DisplayName.String(),
		Comment:     src.Comment,
		Description: src.Description,
		Ips:         new(common.IPs),
	}

	for _, ip := range slices.Concat(src.IPs.IPv4.Values(), src.IPs.IPv6.Values()) {
		if ip.Is4() {
			dest.Ips.Ipv4 = append(dest.Ips.Ipv4, ip.String())
		} else if ip.Is6() {
			dest.Ips.Ipv6 = append(dest.Ips.Ipv6, ip.String())
		}
	}

	if err = Domain2Proto(DTO(src.MetaInfo, &dest.MetaInfo)); err != nil {
		return dest, err
	}

	err = Domain2Proto(DTO(src.Endpoints, &dest.Endpoints))

	return dest, err
}

func hostToProto(src domain.Host) (dest *pb.Host, err error) {
	dest = new(pb.Host)
	if err = cdto.Domain2Proto(DTO(src.Metadata, &dest.Metadata)); err != nil {
		return dest, err
	}
	err = Domain2Proto(DTO(src.Spec, &dest.Spec))
	return dest, err
}

func hostsToProto(src domain.Hosts) (dest *pb.HostResp_Upsert, err error) {
	dest = &pb.HostResp_Upsert{
		Hosts: misc.Tern(len(src) > 0, make([]*pb.Host, len(src)), nil),
	}
	for i, ns := range src {
		if err = Domain2Proto(DTO(ns, &dest.Hosts[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func hostExtToProto(src domain.Host) (dest *pb.HostResp_HostExt, err error) {
	dest = new(pb.HostResp_HostExt)
	if err = cdto.Domain2Proto(DTO(src.Metadata, &dest.Metadata)); err != nil {
		return dest, err
	}
	dest.Refs = misc.Tern(len(src.Refs) > 0, make([]*common.ResourceRef, len(src.Refs)), nil)
	for i, ref := range src.Refs {
		if err = cdto.Domain2Proto(DTO(ref, &dest.Refs[i])); err != nil {
			return dest, err
		}
	}
	err = Domain2Proto(DTO(src.Spec, &dest.Spec))
	return dest, err
}

func hostListToProto(src domain.HostList) (dest *pb.HostResp_List, err error) {
	dest = &pb.HostResp_List{
		ResourceVersion: src.ResourceVersion,
		Hosts:           misc.Tern(len(src.Items) > 0, make([]*pb.HostResp_HostExt, len(src.Items)), nil),
	}
	for i, ns := range src.Items {
		if err = Domain2Proto(DTO(ns, &dest.Hosts[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func hostEventToProto(src domain.HostEvent) (dest *pb.HostResp_Watch, err error) {
	dest = &pb.HostResp_Watch{
		Type: common.WatchEventType(src.EventType),
	}
	dest.Hosts = make([]*pb.HostResp_HostExt, 1)
	if err = Domain2Proto(DTO(src.Object, &dest.Hosts[0])); err != nil {
		return dest, err
	}
	return dest, nil
}

func updIpToProto(src domain.Hosts) (dest *pb.HostResp_UpdIPs, err error) {
	dest = &pb.HostResp_UpdIPs{
		Hosts: misc.Tern(len(src) > 0, make([]*pb.Host, len(src)), nil),
	}
	for i, h := range src {
		if err = Domain2Proto(DTO(h, &dest.Hosts[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func updMetaInfoToProto(src domain.Hosts) (dest *pb.HostResp_UpdMetaInfo, err error) {
	dest = &pb.HostResp_UpdMetaInfo{
		Hosts: misc.Tern(len(src) > 0, make([]*pb.Host, len(src)), nil),
	}
	for i, h := range src {
		if err = Domain2Proto(DTO(h, &dest.Hosts[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

// TODO: Раскоментировать после обновления proto.
//func updHealthToProto(src domain.Hosts) (dest *pb.HostResp_UpdHealthStatus, err error) {
//	dest = &pb.HostResp_UpdHealthStatus{
//		Hosts: misc.Tern(len(src) > 0, make([]*pb.Host, len(src)), nil),
//	}
//	for i, h := range src {
//		if err = Domain2Proto(DTO(h, &dest.Hosts[i])); err != nil {
//			return dest, err
//		}
//	}
//	return dest, nil
//}
