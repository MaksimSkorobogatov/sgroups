package dto

import (
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	cdto "github.com/PRO-Robotech/sgroups/internal/shared/dto/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
)

func init() {
	dto.Register[domain.ServiceSpec, *pb.Service_Spec](specToProto)
	dto.Register[domain.Service, *pb.Service](serviceToProto)
	dto.Register[domain.Services, *pb.ServiceResp_Upsert](servicesToProto)
	dto.Register[domain.Service, *pb.ServiceResp_ServiceExt](serviceExtToProto)
	dto.Register[domain.ServiceList, *pb.ServiceResp_List](serviceListToProto)
	dto.Register[domain.ServiceEvent, *pb.ServiceResp_Watch](serviceEventToProto)
}

// Domain2Proto -
func Domain2Proto[v domain2protoVariants](obj v) error {
	return errors.WithMessage(obj.Convert(), "domain -> proto dto convertation")
}

type domain2protoVariants interface {
	*dto.Pair[domain.ServiceSpec, *pb.Service_Spec] |
		*dto.Pair[domain.Service, *pb.Service] |
		*dto.Pair[domain.Services, *pb.ServiceResp_Upsert] |
		*dto.Pair[domain.Service, *pb.ServiceResp_ServiceExt] |
		*dto.Pair[domain.ServiceList, *pb.ServiceResp_List] |
		*dto.Pair[domain.ServiceEvent, *pb.ServiceResp_Watch]
	Convert() error
}

func specToProto(src domain.ServiceSpec) (dest *pb.Service_Spec, err error) {
	dest = &pb.Service_Spec{
		DisplayName: src.DisplayName.String(),
		Comment:     src.Comment,
		Description: src.Description,
	}

	dest.Transports = misc.Tern(len(src.Transports) > 0, make([]*common.Transport, len(src.Transports)), nil)
	for i, t := range src.Transports {
		switch v := t.(type) {
		case domain.IcmpTransport:
			err = cdto.Domain2Proto(cdto.DTO(v, &dest.Transports[i]))
		case domain.L4Transport:
			err = cdto.Domain2Proto(cdto.DTO(v, &dest.Transports[i]))
		}
		if err != nil {
			return dest, err
		}
	}

	return dest, nil
}

func serviceToProto(src domain.Service) (dest *pb.Service, err error) {
	dest = new(pb.Service)
	if err = cdto.Domain2Proto(cdto.DTO(src.Metadata, &dest.Metadata)); err != nil {
		return dest, err
	}
	err = Domain2Proto(DTO(src.Spec, &dest.Spec))
	return dest, err
}

func servicesToProto(src domain.Services) (dest *pb.ServiceResp_Upsert, err error) {
	dest = &pb.ServiceResp_Upsert{
		Services: misc.Tern(len(src) > 0, make([]*pb.Service, len(src)), nil),
	}
	for i, svc := range src {
		if err = Domain2Proto(DTO(svc, &dest.Services[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func serviceExtToProto(src domain.Service) (dest *pb.ServiceResp_ServiceExt, err error) {
	dest = new(pb.ServiceResp_ServiceExt)
	if err = cdto.Domain2Proto(cdto.DTO(src.Metadata, &dest.Metadata)); err != nil {
		return dest, err
	}
	dest.Refs = misc.Tern(len(src.Refs) > 0, make([]*common.ResourceRef, len(src.Refs)), nil)
	for i, ref := range src.Refs {
		if err = cdto.Domain2Proto(cdto.DTO(ref, &dest.Refs[i])); err != nil {
			return dest, err
		}
	}
	err = Domain2Proto(DTO(src.Spec, &dest.Spec))
	return dest, err
}

func serviceListToProto(src domain.ServiceList) (dest *pb.ServiceResp_List, err error) {
	dest = &pb.ServiceResp_List{
		ResourceVersion: src.ResourceVersion,
		Services:        misc.Tern(len(src.Items) > 0, make([]*pb.ServiceResp_ServiceExt, len(src.Items)), nil),
	}
	for i, svc := range src.Items {
		if err = Domain2Proto(DTO(svc, &dest.Services[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func serviceEventToProto(src domain.ServiceEvent) (dest *pb.ServiceResp_Watch, err error) {
	dest = &pb.ServiceResp_Watch{
		Type: common.WatchEventType(src.EventType),
	}
	dest.Services = make([]*pb.ServiceResp_ServiceExt, 1)
	if err = Domain2Proto(DTO(src.Object, &dest.Services[0])); err != nil {
		return dest, err
	}
	return dest, nil
}
