package dto

import (
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func init() {
	dto.Register[*common.Metadata, domain.NsMetadata](nsMetadataToDomain)
	dto.Register[*pb.Namespace_Spec, domain.NamespaceSpec](nsSpecToDomain)
	dto.Register[*pb.Namespace, domain.Namespace](namespaceToDomain)
	dto.Register[*pb.NamespaceReq_Upsert, domain.Namespaces](upsertReqToDomain)
	dto.Register[*pb.NamespaceReq_Delete_MetadataScope, domain.NsMetadata](deleteReqMetadataToDomain)
	dto.Register[*pb.NamespaceReq_Delete_Namespace, domain.Namespace](deleteReqNsToDomain)
	dto.Register[*pb.NamespaceReq_Delete, domain.Namespaces](deleteReqToDomain)
	dto.Register[*pb.NamespaceReq_Selector_FieldSelector, domain.ResFieldSelector](fieldSelectorToDomain)
	dto.Register[*pb.NamespaceReq_Selector, domain.ResSelector](resSelectorToDomain)
	dto.Register[*pb.NamespaceReq_List, domain.ResSelectorList](nsListToDomain)
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
	*dto.Pair[*common.Metadata, domain.NsMetadata] |
		*dto.Pair[*pb.Namespace_Spec, domain.NamespaceSpec] |
		*dto.Pair[*pb.Namespace, domain.Namespace] |
		*dto.Pair[*pb.NamespaceReq_Upsert, domain.Namespaces] |
		*dto.Pair[*pb.NamespaceReq_Delete_MetadataScope, domain.NsMetadata] |
		*dto.Pair[*pb.NamespaceReq_Delete_Namespace, domain.Namespace] |
		*dto.Pair[*pb.NamespaceReq_Delete, domain.Namespaces] |
		*dto.Pair[*pb.NamespaceReq_Selector_FieldSelector, domain.ResFieldSelector] |
		*dto.Pair[*pb.NamespaceReq_Selector, domain.ResSelector] |
		*dto.Pair[*pb.NamespaceReq_List, domain.ResSelectorList]
	Convert() error
}

func nsMetadataToDomain(src *common.Metadata) (dest domain.NsMetadata, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = domain.NsMetadata{
		ID: domain.ClusterScopeMetadataIdentity{
			Name: domain.ResourceName(src.GetName()),
		},
		Labels:            src.GetLabels(),
		Annotations:       src.GetAnnotations(),
		CreationTimestamp: src.GetCreationTimestamp().AsTime(),
		ResourceVersion:   src.GetResourceVersion(),
	}
	if len(src.GetUid()) > 0 {
		if dest.ID.UID, err = uuid.Parse(src.GetUid()); err != nil {
			err = errors.WithMessagef(err, "bad 'UUID' '%s'", src.GetUid())
		}
	}
	return dest, err
}

func nsSpecToDomain(src *pb.Namespace_Spec) (dest domain.NamespaceSpec, err error) {
	dest = domain.NamespaceSpec{
		DisplayName: domain.DisplayName(src.GetDisplayName()),
		Comment:     src.GetComment(),
		Description: src.GetDescription(),
	}
	return dest, nil
}

func namespaceToDomain(src *pb.Namespace) (dest domain.Namespace, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	err = Proto2Domain(DTO(src.GetMetadata(), &dest.Metadata))
	if err != nil {
		return dest, err
	}
	err = Proto2Domain(DTO(src.GetSpec(), &dest.Spec))
	return dest, err
}

func upsertReqToDomain(src *pb.NamespaceReq_Upsert) (dest domain.Namespaces, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetNamespaces()) > 0, make(domain.Namespaces, len(src.GetNamespaces())), nil)
	for i, ns := range src.GetNamespaces() {
		if err = Proto2Domain(DTO(ns, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func deleteReqMetadataToDomain(src *pb.NamespaceReq_Delete_MetadataScope) (dest domain.NsMetadata, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = domain.NsMetadata{
		ID: domain.ClusterScopeMetadataIdentity{
			Name: domain.ResourceName(src.GetName()),
		},
	}
	if len(src.GetUid()) > 0 {
		if dest.ID.UID, err = uuid.Parse(src.GetUid()); err != nil {
			err = errors.WithMessagef(err, "bad 'UUID' '%s'", src.GetUid())
		}
	}
	return dest, err
}

func deleteReqNsToDomain(src *pb.NamespaceReq_Delete_Namespace) (dest domain.Namespace, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	err = Proto2Domain(DTO(src.GetMetadata(), &dest.Metadata))
	return dest, err
}

func deleteReqToDomain(src *pb.NamespaceReq_Delete) (dest domain.Namespaces, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetNamespaces()) > 0, make(domain.Namespaces, len(src.GetNamespaces())), nil)
	for i, ns := range src.GetNamespaces() {
		if err = Proto2Domain(DTO(ns, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func fieldSelectorToDomain(src *pb.NamespaceReq_Selector_FieldSelector) (dest domain.ResFieldSelector, err error) {
	dest = domain.ResFieldSelector{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name: domain.ResourceName(src.GetName()),
		},
	}
	return dest, nil
}

func resSelectorToDomain(src *pb.NamespaceReq_Selector) (dest domain.ResSelector, err error) {
	dest = domain.ResSelector{
		LabelSelector: src.GetLabelSelector(),
	}
	err = Proto2Domain(DTO(src.GetFieldSelector(), &dest.FieldSelector))
	return dest, err
}

func nsListToDomain(src *pb.NamespaceReq_List) (dest domain.ResSelectorList, err error) {
	dest = misc.Tern(len(src.GetSelectors()) > 0, make(domain.ResSelectorList, len(src.GetSelectors())), nil)
	for i, sel := range src.GetSelectors() {
		if err = Proto2Domain(DTO(sel, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}
