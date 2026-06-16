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
	dto.Register[domain.NetworkSpec, *pb.Network_Spec](specToProto)
	dto.Register[domain.Network, *pb.Network](networkToProto)
	dto.Register[domain.Networks, *pb.NetworkResp_Upsert](networksToProto)
	dto.Register[domain.Network, *pb.NetworkResp_NetworkExt](networkExtToProto)
	dto.Register[domain.NetworkList, *pb.NetworkResp_List](networkListToProto)
	dto.Register[domain.NetworkEvent, *pb.NetworkResp_Watch](networkEventToProto)
}

// Domain2Proto -
func Domain2Proto[v domain2protoVariants](obj v) error {
	return errors.WithMessage(obj.Convert(), "domain -> proto dto convertation")
}

type domain2protoVariants interface {
	*dto.Pair[domain.NetworkSpec, *pb.Network_Spec] |
		*dto.Pair[domain.Network, *pb.Network] |
		*dto.Pair[domain.Networks, *pb.NetworkResp_Upsert] |
		*dto.Pair[domain.Network, *pb.NetworkResp_NetworkExt] |
		*dto.Pair[domain.NetworkList, *pb.NetworkResp_List] |
		*dto.Pair[domain.NetworkEvent, *pb.NetworkResp_Watch]
	Convert() error
}

func specToProto(src domain.NetworkSpec) (dest *pb.Network_Spec, err error) {
	dest = &pb.Network_Spec{
		DisplayName: src.DisplayName.String(),
		Comment:     src.Comment,
		Description: src.Description,
		Cidr:        misc.Tern(len(src.CIDR.IP) > 0, src.CIDR.String(), ""),
	}

	return dest, nil
}

func networkToProto(src domain.Network) (dest *pb.Network, err error) {
	dest = new(pb.Network)
	if err = cdto.Domain2Proto(DTO(src.Metadata, &dest.Metadata)); err != nil {
		return dest, err
	}
	err = Domain2Proto(DTO(src.Spec, &dest.Spec))
	return dest, err
}

func networksToProto(src domain.Networks) (dest *pb.NetworkResp_Upsert, err error) {
	dest = &pb.NetworkResp_Upsert{
		Networks: misc.Tern(len(src) > 0, make([]*pb.Network, len(src)), nil),
	}
	for i, nw := range src {
		if err = Domain2Proto(DTO(nw, &dest.Networks[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func networkExtToProto(src domain.Network) (dest *pb.NetworkResp_NetworkExt, err error) {
	dest = new(pb.NetworkResp_NetworkExt)
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

func networkListToProto(src domain.NetworkList) (dest *pb.NetworkResp_List, err error) {
	dest = &pb.NetworkResp_List{
		ResourceVersion: src.ResourceVersion,
		Networks:        misc.Tern(len(src.Items) > 0, make([]*pb.NetworkResp_NetworkExt, len(src.Items)), nil),
	}
	for i, nw := range src.Items {
		if err = Domain2Proto(DTO(nw, &dest.Networks[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func networkEventToProto(src domain.NetworkEvent) (dest *pb.NetworkResp_Watch, err error) {
	dest = &pb.NetworkResp_Watch{
		Type: common.WatchEventType(src.EventType),
	}
	dest.Networks = make([]*pb.NetworkResp_NetworkExt, 1)
	if err = Domain2Proto(DTO(src.Object, &dest.Networks[0])); err != nil {
		return dest, err
	}
	return dest, nil
}
