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
	dto.Register[domain.HostBindingSpec, *pb.HostBinding_Spec](specToProto)
	dto.Register[domain.HostBinding, *pb.HostBinding](hbToProto)
	dto.Register[domain.HostBindings, *pb.HostBindingResp_Upsert](hbsToProto)
	dto.Register[domain.HostBindingList, *pb.HostBindingResp_List](hbListToProto)
	dto.Register[domain.HostBindingEvent, *pb.HostBindingResp_Watch](hbEventToProto)
}

// Domain2Proto -
func Domain2Proto[v domain2protoVariants](obj v) error {
	return errors.WithMessage(obj.Convert(), "domain -> proto dto convertation")
}

type domain2protoVariants interface {
	*dto.Pair[domain.HostBindingSpec, *pb.HostBinding_Spec] |
		*dto.Pair[domain.HostBinding, *pb.HostBinding] |
		*dto.Pair[domain.HostBindings, *pb.HostBindingResp_Upsert] |
		*dto.Pair[domain.HostBindingList, *pb.HostBindingResp_List] |
		*dto.Pair[domain.HostBindingEvent, *pb.HostBindingResp_Watch]
	Convert() error
}

func specToProto(src domain.HostBindingSpec) (dest *pb.HostBinding_Spec, err error) {
	dest = &pb.HostBinding_Spec{
		DisplayName: src.DisplayName.String(),
		Comment:     src.Comment,
		Description: src.Description,
	}
	if err = cdto.Domain2Proto(cdto.DTO(src.AddressGroup, &dest.AddressGroup)); err != nil {
		return dest, err
	}
	err = cdto.Domain2Proto(cdto.DTO(src.Host, &dest.Host))
	return dest, err
}

func hbToProto(src domain.HostBinding) (dest *pb.HostBinding, err error) {
	dest = new(pb.HostBinding)
	if err = cdto.Domain2Proto(DTO(src.Metadata, &dest.Metadata)); err != nil {
		return dest, err
	}
	err = Domain2Proto(DTO(src.Spec, &dest.Spec))
	return dest, err
}

func hbsToProto(src domain.HostBindings) (dest *pb.HostBindingResp_Upsert, err error) {
	dest = &pb.HostBindingResp_Upsert{
		HostBindings: misc.Tern(len(src) > 0, make([]*pb.HostBinding, len(src)), nil),
	}
	for i, ns := range src {
		if err = Domain2Proto(DTO(ns, &dest.HostBindings[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func hbListToProto(src domain.HostBindingList) (dest *pb.HostBindingResp_List, err error) {
	dest = &pb.HostBindingResp_List{
		ResourceVersion: src.ResourceVersion,
		HostBindings:    misc.Tern(len(src.Items) > 0, make([]*pb.HostBinding, len(src.Items)), nil),
	}
	for i, hb := range src.Items {
		if err = Domain2Proto(DTO(hb, &dest.HostBindings[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func hbEventToProto(src domain.HostBindingEvent) (dest *pb.HostBindingResp_Watch, err error) {
	dest = &pb.HostBindingResp_Watch{
		Type: common.WatchEventType(src.EventType),
	}
	dest.HostBindings = make([]*pb.HostBinding, 1)
	if err = Domain2Proto(DTO(src.Object, &dest.HostBindings[0])); err != nil {
		return dest, err
	}
	return dest, nil
}
