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
	dto.Register[*pb.HostBinding_Spec, domain.HostBindingSpec](specToDomain)
	dto.Register[*pb.HostBinding, domain.HostBinding](hbToDomain)
	dto.Register[*pb.HostBindingReq_Upsert, domain.HostBindings](upsertReqToDomain)
	dto.Register[*pb.HostBindingReq_Selectors, domain.HostBindingSelector](hbSelectorToDomain)
	dto.Register[*pb.HostBindingReq_Selectors_FieldSelector, domain.HostBindingFieldSelector](hbFildSelectorTODomain)
	dto.Register[*pb.HostBindingReq_List, domain.HostBindingSelectorList](hbListToDomain)
	dto.Register[*pb.HostBindingReq_Delete_HostBinding, domain.HostBinding](deleteReqHbToDomain)
	dto.Register[*pb.HostBindingReq_Delete, domain.HostBindings](deleteReqToDomain)
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
	*dto.Pair[*pb.HostBinding_Spec, domain.HostBindingSpec] |
		*dto.Pair[*pb.HostBinding, domain.HostBinding] |
		*dto.Pair[*pb.HostBindingReq_Upsert, domain.HostBindings] |
		*dto.Pair[*pb.HostBindingReq_Selectors, domain.HostBindingSelector] |
		*dto.Pair[*pb.HostBindingReq_Selectors_FieldSelector, domain.HostBindingFieldSelector] |
		*dto.Pair[*pb.HostBindingReq_List, domain.HostBindingSelectorList] |
		*dto.Pair[*pb.HostBindingReq_Delete_HostBinding, domain.HostBinding] |
		*dto.Pair[*pb.HostBindingReq_Delete, domain.HostBindings]
	Convert() error
}

func specToDomain(src *pb.HostBinding_Spec) (dest domain.HostBindingSpec, err error) {
	dest = domain.HostBindingSpec{
		CommonSpec: domain.CommonSpec{
			DisplayName: domain.DisplayName(src.GetDisplayName()),
			Comment:     src.GetComment(),
			Description: src.GetDescription(),
		},
	}
	if err = cdto.Proto2Domain(cdto.DTO(src.GetAddressGroup(), &dest.AddressGroup)); err != nil {
		return dest, err
	}
	if err = cdto.Proto2Domain(cdto.DTO(src.GetHost(), &dest.Host)); err != nil {
		return dest, err
	}

	return dest, err
}

func hbToDomain(src *pb.HostBinding) (dest domain.HostBinding, err error) {
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

func upsertReqToDomain(src *pb.HostBindingReq_Upsert) (dest domain.HostBindings, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetHostBindings()) > 0, make(domain.HostBindings, len(src.GetHostBindings())), nil)
	for i, hb := range src.GetHostBindings() {
		if err = Proto2Domain(DTO(hb, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func hbFildSelectorTODomain(src *pb.HostBindingReq_Selectors_FieldSelector) (dst domain.HostBindingFieldSelector, err error) {
	dst = domain.HostBindingFieldSelector{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.GetName()),
			Namespace: domain.ResourceNamespace(src.GetNamespace()),
		},
	}
	if err = cdto.Proto2Domain(cdto.DTO(src.GetAddressGroup(), &dst.AddressGroup)); err != nil {
		return dst, err
	}
	err = cdto.Proto2Domain(cdto.DTO(src.GetHost(), &dst.Host))
	return dst, err
}

func hbSelectorToDomain(src *pb.HostBindingReq_Selectors) (dst domain.HostBindingSelector, err error) {
	dst = domain.HostBindingSelector{
		LabelSelector: src.GetLabelSelector(),
	}
	err = Proto2Domain(DTO(src.GetFieldSelector(), &dst.FieldSelector))
	return dst, err
}

func hbListToDomain(src *pb.HostBindingReq_List) (dest domain.HostBindingSelectorList, err error) {
	dest = misc.Tern(len(src.GetSelectors()) > 0, make(domain.HostBindingSelectorList, len(src.GetSelectors())), nil)
	for i, sel := range src.GetSelectors() {
		if err = Proto2Domain(DTO(sel, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func deleteReqHbToDomain(src *pb.HostBindingReq_Delete_HostBinding) (dest domain.HostBinding, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	err = cdto.Proto2Domain(cdto.DTO(src.GetMetadata(), &dest.Metadata))
	return dest, err
}

func deleteReqToDomain(src *pb.HostBindingReq_Delete) (dest domain.HostBindings, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetHostBindings()) > 0, make(domain.HostBindings, len(src.GetHostBindings())), nil)
	for i, hb := range src.GetHostBindings() {
		if err = Proto2Domain(DTO(hb, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}
