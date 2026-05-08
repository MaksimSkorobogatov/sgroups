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
	dto.Register[*pb.AddressGroupResp_AddressGroupExt, domain.AddressGroup](agToDomain)
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
		*dto.Pair[*pb.AddressGroupResp_AddressGroupExt, domain.AddressGroup]
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

func agToDomain(src *pb.AddressGroupResp_AddressGroupExt) (dest domain.AddressGroup, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	err = cdto.Proto2Domain(cdto.DTO(src.GetMetadata(), &dest.Metadata))
	if err != nil {
		return dest, err
	}
	if err = Proto2Domain(DTO(src.GetSpec(), &dest.Spec)); err != nil {
		return dest, err
	}
	refs := src.GetRefs()
	dest.Refs = misc.Tern(len(refs) > 0, make([]domain.ResourceRef, len(refs)), nil)
	for i, ref := range refs {
		if err = cdto.Proto2Domain(cdto.DTO(ref, &dest.Refs[i])); err != nil {
			break
		}
	}
	return dest, err
}
