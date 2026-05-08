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
	dto.Register[domain.NetworkBindingSpec, *pb.NetworkBinding_Spec](specToProto)
	dto.Register[domain.NetworkBinding, *pb.NetworkBinding](nbToProto)
	dto.Register[domain.NetworkBindings, *pb.NetworkBindingResp_Upsert](nbsToProto)
	dto.Register[domain.NetworkBindingList, *pb.NetworkBindingResp_List](nbListToProto)
	dto.Register[domain.NetworkBindingEvent, *pb.NetworkBindingResp_Watch](nbEventToProto)
}

// Domain2Proto -
func Domain2Proto[v domain2protoVariants](obj v) error {
	return errors.WithMessage(obj.Convert(), "domain -> proto dto convertation")
}

type domain2protoVariants interface {
	*dto.Pair[domain.NetworkBindingSpec, *pb.NetworkBinding_Spec] |
		*dto.Pair[domain.NetworkBinding, *pb.NetworkBinding] |
		*dto.Pair[domain.NetworkBindings, *pb.NetworkBindingResp_Upsert] |
		*dto.Pair[domain.NetworkBindingList, *pb.NetworkBindingResp_List] |
		*dto.Pair[domain.NetworkBindingEvent, *pb.NetworkBindingResp_Watch]
	Convert() error
}

func specToProto(src domain.NetworkBindingSpec) (dest *pb.NetworkBinding_Spec, err error) {
	dest = &pb.NetworkBinding_Spec{
		DisplayName: src.DisplayName.String(),
		Comment:     src.Comment,
		Description: src.Description,
	}
	if err = cdto.Domain2Proto(cdto.DTO(src.AddressGroup, &dest.AddressGroup)); err != nil {
		return dest, err
	}
	err = cdto.Domain2Proto(cdto.DTO(src.Network, &dest.Network))
	return dest, err
}

func nbToProto(src domain.NetworkBinding) (dest *pb.NetworkBinding, err error) {
	dest = new(pb.NetworkBinding)
	if err = cdto.Domain2Proto(DTO(src.Metadata, &dest.Metadata)); err != nil {
		return dest, err
	}
	err = Domain2Proto(DTO(src.Spec, &dest.Spec))
	return dest, err
}

func nbsToProto(src domain.NetworkBindings) (dest *pb.NetworkBindingResp_Upsert, err error) {
	dest = &pb.NetworkBindingResp_Upsert{
		NetworkBindings: misc.Tern(len(src) > 0, make([]*pb.NetworkBinding, len(src)), nil),
	}
	for i, ns := range src {
		if err = Domain2Proto(DTO(ns, &dest.NetworkBindings[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func nbListToProto(src domain.NetworkBindingList) (dest *pb.NetworkBindingResp_List, err error) {
	dest = &pb.NetworkBindingResp_List{
		ResourceVersion: src.ResourceVersion,
		NetworkBindings: misc.Tern(len(src.Items) > 0, make([]*pb.NetworkBinding, len(src.Items)), nil),
	}
	for i, nb := range src.Items {
		if err = Domain2Proto(DTO(nb, &dest.NetworkBindings[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func nbEventToProto(src domain.NetworkBindingEvent) (dest *pb.NetworkBindingResp_Watch, err error) {
	dest = &pb.NetworkBindingResp_Watch{
		Type: common.WatchEventType(src.EventType),
	}
	dest.NetworkBindings = make([]*pb.NetworkBinding, 1)
	if err = Domain2Proto(DTO(src.Object, &dest.NetworkBindings[0])); err != nil {
		return dest, err
	}
	return dest, nil
}
