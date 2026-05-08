package dto

import (
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	cdto "github.com/PRO-Robotech/sgroups/internal/shared/dto/sgroups"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
)

func init() {
	dto.Register[*sgv1.Rule_Spec, domain.RuleSpec](ruleSpec2domain)
	dto.Register[*sgv1.Rule, domain.Rule](rule2domain)
}

// Proto2Domain -
func Proto2Domain[T proto2domainVariants](a T) (err error) {
	return errors.WithMessage(a.Convert(), "common proto -> domain dto convertation")
}

// DTO -
func DTO[tFrom any, tTo any](a tFrom, b *tTo) *dto.Pair[tFrom, tTo] {
	return dto.MakePair(a, b)
}

type proto2domainVariants interface {
	*dto.Pair[*sgv1.Rule_Spec, domain.RuleSpec] |
		*dto.Pair[*sgv1.Rule, domain.Rule]
	Convert() error
}

func ruleSpec2domain(src *sgv1.Rule_Spec) (dest domain.RuleSpec, err error) {
	defer func() {
		err = errors.WithMessagef(err, "%T -> %T", src, dest)
	}()
	dest = domain.RuleSpec{
		CommonSpec: domain.CommonSpec{
			DisplayName: domain.DisplayName(src.GetDisplayName()),
			Comment:     src.GetComment(),
			Description: src.GetDescription(),
		},
		Action: domain.PolicyAction(src.GetAction()), //nolint:gosec
	}
	if err = cdto.Proto2Domain(cdto.DTO(src.GetSession().GetTraffic(), &dest.Traffic)); err != nil {
		return dest, err
	}
	var local domain.EpLocal
	err = cdto.Proto2Domain(cdto.DTO(src.GetEndpoints().GetLocal(), &local))
	if err != nil {
		return dest, err
	}
	dest.Local = local

	err = cdto.Proto2Domain(cdto.DTO(src.GetEndpoints().GetRemote(), &dest.Remote))
	if err != nil {
		return dest, err
	}
	err = cdto.Proto2Domain(cdto.DTO(src.GetTransport(), &dest.Transport))

	return dest, err
}

func rule2domain(src *sgv1.Rule) (dest domain.Rule, err error) {
	err = cdto.Proto2Domain(cdto.DTO(src.GetMetadata(), &dest.Metadata))
	if err != nil {
		return dest, err
	}
	err = Proto2Domain(cdto.DTO(src.GetSpec(), &dest.Spec))
	return dest, err
}
