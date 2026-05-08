package dto

import (
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	cdto "github.com/PRO-Robotech/sgroups/internal/shared/dto/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
)

func init() {
	dto.Register[*pb.ServiceBinding_Spec, domain.ServiceBindingSpec](specToDomain)
	dto.Register[*pb.ServiceBinding, domain.ServiceBinding](sbToDomain)
	dto.Register[*pb.ServiceBindingReq_Upsert, domain.ServiceBindings](upsertReqToDomain)
	dto.Register[*pb.ServiceBindingReq_Selectors, domain.ServiceBindingSelector](sbSelectorToDomain)
	dto.Register[*pb.ServiceBindingReq_Selectors_FieldSelector, domain.ServiceBindingFieldSelector](sbFieldSelectorToDomain)
	dto.Register[*pb.ServiceBindingReq_List, domain.ServiceBindingSelectorList](sbListToDomain)
	dto.Register[*pb.ServiceBindingReq_Delete_ServiceBinding, domain.ServiceBinding](deleteReqSbToDomain)
	dto.Register[*pb.ServiceBindingReq_Delete, domain.ServiceBindings](deleteReqToDomain)
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
	*dto.Pair[*pb.ServiceBinding_Spec, domain.ServiceBindingSpec] |
		*dto.Pair[*pb.ServiceBinding, domain.ServiceBinding] |
		*dto.Pair[*pb.ServiceBindingReq_Upsert, domain.ServiceBindings] |
		*dto.Pair[*pb.ServiceBindingReq_Selectors, domain.ServiceBindingSelector] |
		*dto.Pair[*pb.ServiceBindingReq_Selectors_FieldSelector, domain.ServiceBindingFieldSelector] |
		*dto.Pair[*pb.ServiceBindingReq_List, domain.ServiceBindingSelectorList] |
		*dto.Pair[*pb.ServiceBindingReq_Delete_ServiceBinding, domain.ServiceBinding] |
		*dto.Pair[*pb.ServiceBindingReq_Delete, domain.ServiceBindings]
	Convert() error
}

func specToDomain(src *pb.ServiceBinding_Spec) (dest domain.ServiceBindingSpec, err error) {
	dest = domain.ServiceBindingSpec{
		CommonSpec: domain.CommonSpec{
			DisplayName: domain.DisplayName(src.GetDisplayName()),
			Comment:     src.GetComment(),
			Description: src.GetDescription(),
		},
	}
	if err = cdto.Proto2Domain(cdto.DTO(src.GetAddressGroup(), &dest.AddressGroup)); err != nil {
		return dest, err
	}

	err = cdto.Proto2Domain(cdto.DTO(src.GetService(), &dest.Service))

	return dest, err
}

func sbToDomain(src *pb.ServiceBinding) (dest domain.ServiceBinding, err error) {
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

func upsertReqToDomain(src *pb.ServiceBindingReq_Upsert) (dest domain.ServiceBindings, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetServiceBindings()) > 0, make(domain.ServiceBindings, len(src.GetServiceBindings())), nil)
	for i, sb := range src.GetServiceBindings() {
		if err = Proto2Domain(DTO(sb, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func sbFieldSelectorToDomain(src *pb.ServiceBindingReq_Selectors_FieldSelector) (dst domain.ServiceBindingFieldSelector, err error) {
	dst = domain.ServiceBindingFieldSelector{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.GetName()),
			Namespace: domain.ResourceNamespace(src.GetNamespace()),
		},
	}
	if err = cdto.Proto2Domain(cdto.DTO(src.GetAddressGroup(), &dst.AddressGroup)); err != nil {
		return dst, err
	}
	err = cdto.Proto2Domain(cdto.DTO(src.GetService(), &dst.Service))
	return dst, err
}

func sbSelectorToDomain(src *pb.ServiceBindingReq_Selectors) (dst domain.ServiceBindingSelector, err error) {
	dst = domain.ServiceBindingSelector{
		LabelSelector: src.GetLabelSelector(),
	}
	err = Proto2Domain(DTO(src.GetFieldSelector(), &dst.FieldSelector))
	return dst, err
}

func sbListToDomain(src *pb.ServiceBindingReq_List) (dest domain.ServiceBindingSelectorList, err error) {
	dest = misc.Tern(len(src.GetSelectors()) > 0, make(domain.ServiceBindingSelectorList, len(src.GetSelectors())), nil)
	for i, sel := range src.GetSelectors() {
		if err = Proto2Domain(DTO(sel, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func deleteReqSbToDomain(src *pb.ServiceBindingReq_Delete_ServiceBinding) (dest domain.ServiceBinding, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	err = cdto.Proto2Domain(cdto.DTO(src.GetMetadata(), &dest.Metadata))
	return dest, err
}

func deleteReqToDomain(src *pb.ServiceBindingReq_Delete) (dest domain.ServiceBindings, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetServiceBindings()) > 0, make(domain.ServiceBindings, len(src.GetServiceBindings())), nil)
	for i, sb := range src.GetServiceBindings() {
		if err = Proto2Domain(DTO(sb, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}
