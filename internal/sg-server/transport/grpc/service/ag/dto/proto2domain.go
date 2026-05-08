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
	dto.Register[*pb.AddressGroup_Spec, domain.AgSpec](specToDomain)
	dto.Register[*pb.AddressGroup, domain.AddressGroup](agToDomain)
	dto.Register[*pb.AddressGroupReq_Upsert, domain.AddressGroups](upsertReqToDomain)
	dto.Register[*pb.AddressGroupReq_List, domain.ResSelectorList](agListToDomain)
	dto.Register[*pb.AddressGroupReq_Delete_AddressGroup, domain.AddressGroup](deleteReqAgToDomain)
	dto.Register[*pb.AddressGroupReq_Delete, domain.AddressGroups](deleteReqToDomain)
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
	*dto.Pair[*pb.AddressGroup_Spec, domain.AgSpec] |
		*dto.Pair[*pb.AddressGroup, domain.AddressGroup] |
		*dto.Pair[*pb.AddressGroupReq_Upsert, domain.AddressGroups] |
		*dto.Pair[*pb.AddressGroupReq_List, domain.ResSelectorList] |
		*dto.Pair[*pb.AddressGroupReq_Delete_AddressGroup, domain.AddressGroup] |
		*dto.Pair[*pb.AddressGroupReq_Delete, domain.AddressGroups]
	Convert() error
}

func specToDomain(src *pb.AddressGroup_Spec) (dest domain.AgSpec, err error) {
	dest = domain.AgSpec{
		CommonSpec: domain.CommonSpec{
			DisplayName: domain.DisplayName(src.GetDisplayName()),
			Comment:     src.GetComment(),
			Description: src.GetDescription(),
		},
		DefaultAction: domain.PolicyAction(src.GetDefaultAction()), //nolint:gosec
		Logs:          src.GetLogs(),
		Trace:         src.GetTrace(),
	}

	return dest, nil
}

func agToDomain(src *pb.AddressGroup) (dest domain.AddressGroup, err error) {
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

func upsertReqToDomain(src *pb.AddressGroupReq_Upsert) (dest domain.AddressGroups, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetAddressGroups()) > 0, make(domain.AddressGroups, len(src.GetAddressGroups())), nil)
	for i, ag := range src.GetAddressGroups() {
		if err = Proto2Domain(DTO(ag, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func agListToDomain(src *pb.AddressGroupReq_List) (dest domain.ResSelectorList, err error) {
	dest = misc.Tern(len(src.GetSelectors()) > 0, make(domain.ResSelectorList, len(src.GetSelectors())), nil)
	for i, sel := range src.GetSelectors() {
		if err = cdto.Proto2Domain(cdto.DTO(sel, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func deleteReqAgToDomain(src *pb.AddressGroupReq_Delete_AddressGroup) (dest domain.AddressGroup, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	err = cdto.Proto2Domain(cdto.DTO(src.GetMetadata(), &dest.Metadata))
	return dest, err
}

func deleteReqToDomain(src *pb.AddressGroupReq_Delete) (dest domain.AddressGroups, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = misc.Tern(len(src.GetAddressGroups()) > 0, make(domain.AddressGroups, len(src.GetAddressGroups())), nil)
	for i, ag := range src.GetAddressGroups() {
		if err = Proto2Domain(DTO(ag, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}
