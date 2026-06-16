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
	dto.Register[domain.AgSpec, *pb.AddressGroup_Spec](specToProto)
	dto.Register[domain.AddressGroup, *pb.AddressGroup](agToProto)
	dto.Register[domain.AddressGroups, *pb.AddressGroupResp_Upsert](agsToProto)
	dto.Register[domain.AddressGroup, *pb.AddressGroupResp_AddressGroupExt](agExtToProto)
	dto.Register[domain.AddressGroupList, *pb.AddressGroupResp_List](agListToProto)
	dto.Register[domain.AddressGroupEvent, *pb.AddressGroupResp_Watch](agEventToProto)
}

// Domain2Proto -
func Domain2Proto[v domain2protoVariants](obj v) error {
	return errors.WithMessage(obj.Convert(), "domain -> proto dto convertation")
}

type domain2protoVariants interface {
	*dto.Pair[domain.AgSpec, *pb.AddressGroup_Spec] |
		*dto.Pair[domain.AddressGroup, *pb.AddressGroup] |
		*dto.Pair[domain.AddressGroups, *pb.AddressGroupResp_Upsert] |
		*dto.Pair[domain.AddressGroup, *pb.AddressGroupResp_AddressGroupExt] |
		*dto.Pair[domain.AddressGroupList, *pb.AddressGroupResp_List] |
		*dto.Pair[domain.AddressGroupEvent, *pb.AddressGroupResp_Watch]
	Convert() error
}

func specToProto(src domain.AgSpec) (dest *pb.AddressGroup_Spec, err error) {
	dest = &pb.AddressGroup_Spec{
		DisplayName:   src.DisplayName.String(),
		Comment:       src.Comment,
		Description:   src.Description,
		DefaultAction: common.Action(src.DefaultAction),
		Logs:          src.Logs,
		Trace:         src.Trace,
	}

	return dest, nil
}

func agToProto(src domain.AddressGroup) (dest *pb.AddressGroup, err error) {
	dest = new(pb.AddressGroup)
	if err = cdto.Domain2Proto(cdto.DTO(src.Metadata, &dest.Metadata)); err != nil {
		return dest, err
	}
	err = Domain2Proto(DTO(src.Spec, &dest.Spec))
	return dest, err
}

func agsToProto(src domain.AddressGroups) (dest *pb.AddressGroupResp_Upsert, err error) {
	dest = &pb.AddressGroupResp_Upsert{
		AddressGroups: misc.Tern(len(src) > 0, make([]*pb.AddressGroup, len(src)), nil),
	}
	for i, ns := range src {
		if err = Domain2Proto(DTO(ns, &dest.AddressGroups[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func agExtToProto(src domain.AddressGroup) (dest *pb.AddressGroupResp_AddressGroupExt, err error) {
	dest = new(pb.AddressGroupResp_AddressGroupExt)
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

func agListToProto(src domain.AddressGroupList) (dest *pb.AddressGroupResp_List, err error) {
	dest = &pb.AddressGroupResp_List{
		ResourceVersion: src.ResourceVersion,
		AddressGroups:   misc.Tern(len(src.Items) > 0, make([]*pb.AddressGroupResp_AddressGroupExt, len(src.Items)), nil),
	}
	for i, ns := range src.Items {
		if err = Domain2Proto(DTO(ns, &dest.AddressGroups[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func agEventToProto(src domain.AddressGroupEvent) (dest *pb.AddressGroupResp_Watch, err error) {
	dest = &pb.AddressGroupResp_Watch{
		Type: common.WatchEventType(src.EventType),
	}
	dest.AddressGroups = make([]*pb.AddressGroupResp_AddressGroupExt, 1)
	if err = Domain2Proto(DTO(src.Object, &dest.AddressGroups[0])); err != nil {
		return dest, err
	}
	return dest, nil
}
