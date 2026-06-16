package dto

import (
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	dto.Register[domain.NsMetadata, *common.Metadata](metadataToProto)
	dto.Register[domain.NamespaceSpec, *pb.Namespace_Spec](specToProto)
	dto.Register[domain.Namespace, *pb.Namespace](namespaceToProto)
	dto.Register[domain.Namespaces, *pb.NamespaceResp_Upsert](namespacesToProto)
	dto.Register[domain.NamespaceList, *pb.NamespaceResp_List](nsListToProto)
	dto.Register[domain.NamespaceEvent, *pb.NamespaceResp_Watch](nsEventToProto)
}

// Domain2Proto -
func Domain2Proto[v domain2protoVariants](obj v) error {
	return errors.WithMessage(obj.Convert(), "domain -> proto dto convertation")
}

type domain2protoVariants interface {
	*dto.Pair[domain.NsMetadata, *common.Metadata] |
		*dto.Pair[domain.NamespaceSpec, *pb.Namespace_Spec] |
		*dto.Pair[domain.Namespace, *pb.Namespace] |
		*dto.Pair[domain.Namespaces, *pb.NamespaceResp_Upsert] |
		*dto.Pair[domain.NamespaceList, *pb.NamespaceResp_List] |
		*dto.Pair[domain.NamespaceEvent, *pb.NamespaceResp_Watch]
	Convert() error
}

func metadataToProto(src domain.NsMetadata) (dest *common.Metadata, err error) {
	dest = &common.Metadata{
		Uid:               src.ID.UID.String(),
		Name:              src.ID.Name.String(),
		Labels:            src.Labels,
		Annotations:       src.Annotations,
		CreationTimestamp: timestamppb.New(src.CreationTimestamp),
		ResourceVersion:   src.ResourceVersion,
	}
	return dest, nil
}

func specToProto(src domain.NamespaceSpec) (dest *pb.Namespace_Spec, err error) {
	dest = &pb.Namespace_Spec{
		DisplayName: src.DisplayName.String(),
		Comment:     src.Comment,
		Description: src.Description,
	}
	return dest, nil
}

func namespaceToProto(src domain.Namespace) (dest *pb.Namespace, err error) {
	dest = new(pb.Namespace)
	if err = Domain2Proto(DTO(src.Metadata, &dest.Metadata)); err != nil {
		return dest, err
	}
	err = Domain2Proto(DTO(src.Spec, &dest.Spec))
	return dest, err
}

func namespacesToProto(src domain.Namespaces) (dest *pb.NamespaceResp_Upsert, err error) {
	dest = &pb.NamespaceResp_Upsert{
		Namespaces: misc.Tern(len(src) > 0, make([]*pb.Namespace, len(src)), nil),
	}
	for i, ns := range src {
		if err = Domain2Proto(DTO(ns, &dest.Namespaces[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func nsListToProto(src domain.NamespaceList) (dest *pb.NamespaceResp_List, err error) {
	dest = &pb.NamespaceResp_List{
		ResourceVersion: src.ResourceVersion,
		Namespaces:      misc.Tern(len(src.Items) > 0, make([]*pb.Namespace, len(src.Items)), nil),
	}
	for i, ns := range src.Items {
		if err = Domain2Proto(DTO(ns, &dest.Namespaces[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func nsEventToProto(src domain.NamespaceEvent) (dest *pb.NamespaceResp_Watch, err error) {
	dest = &pb.NamespaceResp_Watch{
		Type: common.WatchEventType(src.EventType),
	}
	dest.Namespaces = make([]*pb.Namespace, 1)
	if err = Domain2Proto(DTO(src.Object, &dest.Namespaces[0])); err != nil {
		return dest, err
	}
	return dest, nil
}
