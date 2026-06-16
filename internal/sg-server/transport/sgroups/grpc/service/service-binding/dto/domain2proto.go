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
	dto.Register[domain.ServiceBindingSpec, *pb.ServiceBinding_Spec](specToProto)
	dto.Register[domain.ServiceBinding, *pb.ServiceBinding](sbToProto)
	dto.Register[domain.ServiceBindings, *pb.ServiceBindingResp_Upsert](sbsToProto)
	dto.Register[domain.ServiceBindingList, *pb.ServiceBindingResp_List](sbListToProto)
	dto.Register[domain.ServiceBindingEvent, *pb.ServiceBindingResp_Watch](sbEventToProto)
}

// Domain2Proto -
func Domain2Proto[v domain2protoVariants](obj v) error {
	return errors.WithMessage(obj.Convert(), "domain -> proto dto convertation")
}

type domain2protoVariants interface {
	*dto.Pair[domain.ServiceBindingSpec, *pb.ServiceBinding_Spec] |
		*dto.Pair[domain.ServiceBinding, *pb.ServiceBinding] |
		*dto.Pair[domain.ServiceBindings, *pb.ServiceBindingResp_Upsert] |
		*dto.Pair[domain.ServiceBindingList, *pb.ServiceBindingResp_List] |
		*dto.Pair[domain.ServiceBindingEvent, *pb.ServiceBindingResp_Watch]
	Convert() error
}

func specToProto(src domain.ServiceBindingSpec) (dest *pb.ServiceBinding_Spec, err error) {
	dest = &pb.ServiceBinding_Spec{
		DisplayName: src.DisplayName.String(),
		Comment:     src.Comment,
		Description: src.Description,
	}
	if err = cdto.Domain2Proto(cdto.DTO(src.AddressGroup, &dest.AddressGroup)); err != nil {
		return dest, err
	}
	err = cdto.Domain2Proto(cdto.DTO(src.Service, &dest.Service))
	return dest, err
}

func sbToProto(src domain.ServiceBinding) (dest *pb.ServiceBinding, err error) {
	dest = new(pb.ServiceBinding)
	if err = cdto.Domain2Proto(cdto.DTO(src.Metadata, &dest.Metadata)); err != nil {
		return dest, err
	}
	err = Domain2Proto(DTO(src.Spec, &dest.Spec))
	return dest, err
}

func sbsToProto(src domain.ServiceBindings) (dest *pb.ServiceBindingResp_Upsert, err error) {
	dest = &pb.ServiceBindingResp_Upsert{
		ServiceBindings: misc.Tern(len(src) > 0, make([]*pb.ServiceBinding, len(src)), nil),
	}
	for i, sb := range src {
		if err = Domain2Proto(DTO(sb, &dest.ServiceBindings[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func sbListToProto(src domain.ServiceBindingList) (dest *pb.ServiceBindingResp_List, err error) {
	dest = &pb.ServiceBindingResp_List{
		ResourceVersion: src.ResourceVersion,
		ServiceBindings: misc.Tern(len(src.Items) > 0, make([]*pb.ServiceBinding, len(src.Items)), nil),
	}
	for i, sb := range src.Items {
		if err = Domain2Proto(DTO(sb, &dest.ServiceBindings[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func sbEventToProto(src domain.ServiceBindingEvent) (dest *pb.ServiceBindingResp_Watch, err error) {
	dest = &pb.ServiceBindingResp_Watch{
		Type: common.WatchEventType(src.EventType),
	}
	dest.ServiceBindings = make([]*pb.ServiceBinding, 1)
	if err = Domain2Proto(DTO(src.Object, &dest.ServiceBindings[0])); err != nil {
		return dest, err
	}
	return dest, nil
}
