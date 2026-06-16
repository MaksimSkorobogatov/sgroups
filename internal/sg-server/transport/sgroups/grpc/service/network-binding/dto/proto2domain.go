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
	dto.Register[*pb.NetworkBinding_Spec, domain.NetworkBindingSpec](specToDomain)
	dto.Register[*pb.NetworkBinding, domain.NetworkBinding](nbToDomain)
	dto.Register[*pb.NetworkBindingReq_Upsert, domain.NetworkBindings](upsertReqToDomain)
	dto.Register[*pb.NetworkBindingReq_Selectors, domain.NetworkBindingSelector](nbSelectorToDomain)
	dto.Register[*pb.NetworkBindingReq_Selectors_FieldSelector, domain.NetworkBindingFieldSelector](nbFieldSelectorToDomain)
	dto.Register[*pb.NetworkBindingReq_List, domain.NetworkBindingSelectorList](nbListToDomain)
	dto.Register[*pb.NetworkBindingReq_Delete_NetworkBinding, domain.NetworkBinding](deleteReqNbToDomain)
	dto.Register[*pb.NetworkBindingReq_Delete, domain.NetworkBindings](deleteReqToDomain)
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
	*dto.Pair[*pb.NetworkBinding_Spec, domain.NetworkBindingSpec] |
		*dto.Pair[*pb.NetworkBinding, domain.NetworkBinding] |
		*dto.Pair[*pb.NetworkBindingReq_Upsert, domain.NetworkBindings] |
		*dto.Pair[*pb.NetworkBindingReq_Selectors, domain.NetworkBindingSelector] |
		*dto.Pair[*pb.NetworkBindingReq_Selectors_FieldSelector, domain.NetworkBindingFieldSelector] |
		*dto.Pair[*pb.NetworkBindingReq_List, domain.NetworkBindingSelectorList] |
		*dto.Pair[*pb.NetworkBindingReq_Delete_NetworkBinding, domain.NetworkBinding] |
		*dto.Pair[*pb.NetworkBindingReq_Delete, domain.NetworkBindings]
	Convert() error
}

func specToDomain(src *pb.NetworkBinding_Spec) (dest domain.NetworkBindingSpec, err error) {
	dest = domain.NetworkBindingSpec{
		CommonSpec: domain.CommonSpec{
			DisplayName: domain.DisplayName(src.GetDisplayName()),
			Comment:     src.GetComment(),
			Description: src.GetDescription(),
		},
	}
	if err = cdto.Proto2Domain(cdto.DTO(src.GetAddressGroup(), &dest.AddressGroup)); err != nil {
		return dest, err
	}

	err = cdto.Proto2Domain(cdto.DTO(src.GetNetwork(), &dest.Network))

	return dest, err
}

func nbToDomain(src *pb.NetworkBinding) (dest domain.NetworkBinding, err error) {
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

func upsertReqToDomain(src *pb.NetworkBindingReq_Upsert) (dest domain.NetworkBindings, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetNetworkBindings()) > 0, make(domain.NetworkBindings, len(src.GetNetworkBindings())), nil)
	for i, nb := range src.GetNetworkBindings() {
		if err = Proto2Domain(DTO(nb, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func nbFieldSelectorToDomain(src *pb.NetworkBindingReq_Selectors_FieldSelector) (dst domain.NetworkBindingFieldSelector, err error) {
	dst = domain.NetworkBindingFieldSelector{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.GetName()),
			Namespace: domain.ResourceNamespace(src.GetNamespace()),
		},
	}
	if err = cdto.Proto2Domain(cdto.DTO(src.GetAddressGroup(), &dst.AddressGroup)); err != nil {
		return dst, err
	}
	err = cdto.Proto2Domain(cdto.DTO(src.GetNetwork(), &dst.Network))
	return dst, err
}

func nbSelectorToDomain(src *pb.NetworkBindingReq_Selectors) (dst domain.NetworkBindingSelector, err error) {
	dst = domain.NetworkBindingSelector{
		LabelSelector: src.GetLabelSelector(),
	}
	err = Proto2Domain(DTO(src.GetFieldSelector(), &dst.FieldSelector))
	return dst, err
}

func nbListToDomain(src *pb.NetworkBindingReq_List) (dest domain.NetworkBindingSelectorList, err error) {
	dest = misc.Tern(len(src.GetSelectors()) > 0, make(domain.NetworkBindingSelectorList, len(src.GetSelectors())), nil)
	for i, sel := range src.GetSelectors() {
		if err = Proto2Domain(DTO(sel, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func deleteReqNbToDomain(src *pb.NetworkBindingReq_Delete_NetworkBinding) (dest domain.NetworkBinding, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	err = cdto.Proto2Domain(cdto.DTO(src.GetMetadata(), &dest.Metadata))
	return dest, err
}

func deleteReqToDomain(src *pb.NetworkBindingReq_Delete) (dest domain.NetworkBindings, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetNetworkBindings()) > 0, make(domain.NetworkBindings, len(src.GetNetworkBindings())), nil)
	for i, nb := range src.GetNetworkBindings() {
		if err = Proto2Domain(DTO(nb, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}
