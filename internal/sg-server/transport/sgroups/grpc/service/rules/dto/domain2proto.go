package dto

import (
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	cdto "github.com/PRO-Robotech/sgroups/internal/shared/dto/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
)

func init() {
	dto.Register[domain.RuleSpec, *sgv1.Rule_Spec](ruleSpec2proto)
	dto.Register[domain.Rule, *sgv1.Rule](rule2proto)
	dto.Register[domain.Rules, *sgv1.RuleResp_Upsert](rules2proto)
	dto.Register[domain.RuleList, *sgv1.RuleResp_List](rlListToProto)
	dto.Register[domain.RuleEvent, *sgv1.RuleResp_Watch](rlEventToProto)
}

// Domain2Proto -
func Domain2Proto[T domain2protoVariants](a T) (err error) {
	return errors.WithMessage(a.Convert(), "common domain -> proto dto convertation")
}

// DTO -
func DTO[tFrom any, tTo any](a tFrom, b *tTo) *dto.Pair[tFrom, tTo] {
	return dto.MakePair(a, b)
}

type domain2protoVariants interface {
	*dto.Pair[domain.RuleSpec, *sgv1.Rule_Spec] |
		*dto.Pair[domain.Rule, *sgv1.Rule] |
		*dto.Pair[domain.Rules, *sgv1.RuleResp_Upsert] |
		*dto.Pair[domain.RuleList, *sgv1.RuleResp_List] |
		*dto.Pair[domain.RuleEvent, *sgv1.RuleResp_Watch]

	Convert() error
}

func ruleSpec2proto(src domain.RuleSpec) (dest *sgv1.Rule_Spec, err error) {
	session := new(common.Session)
	if err = cdto.Domain2Proto(cdto.DTO(src.Traffic, &session.Traffic)); err != nil {
		return dest, err
	}
	dest = &sgv1.Rule_Spec{
		DisplayName: src.DisplayName.String(),
		Comment:     src.Comment,
		Description: src.Description,
		Action:      common.Action(src.Action), //nolint:gosec
		Session:     session,
		Endpoints:   new(common.Endpoints),
	}
	if local, ok := src.Local.(domain.EpLocal); ok {
		err = cdto.Domain2Proto(cdto.DTO(local, &dest.Endpoints.Local))
		if err != nil {
			return dest, err
		}
	}
	err = cdto.Domain2Proto(cdto.DTO(src.Remote, &dest.Endpoints.Remote))
	if err != nil {
		return dest, err
	}

	err = cdto.Domain2Proto(cdto.DTO(src.Transport, &dest.Transport))

	return dest, err
}

func rule2proto(src domain.Rule) (dest *sgv1.Rule, err error) {
	dest = new(sgv1.Rule)
	err = cdto.Domain2Proto(cdto.DTO(src.Metadata, &dest.Metadata))
	if err != nil {
		return dest, err
	}
	err = Domain2Proto(DTO(src.Spec, &dest.Spec))
	return dest, err
}

func rules2proto(src domain.Rules) (dest *sgv1.RuleResp_Upsert, err error) {
	dest = &sgv1.RuleResp_Upsert{
		Rules: misc.Tern(len(src) > 0, make([]*sgv1.Rule, len(src)), nil),
	}
	for i, r := range src {
		err = Domain2Proto(DTO(r, &dest.Rules[i]))
		if err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func rlListToProto(src domain.RuleList) (dest *sgv1.RuleResp_List, err error) {
	dest = &sgv1.RuleResp_List{
		ResourceVersion: src.ResourceVersion,
		Rules:           misc.Tern(len(src.Items) > 0, make([]*sgv1.Rule, len(src.Items)), nil),
	}
	for i, rl := range src.Items {
		if err = Domain2Proto(DTO(rl, &dest.Rules[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func rlEventToProto(src domain.RuleEvent) (dest *sgv1.RuleResp_Watch, err error) {
	dest = &sgv1.RuleResp_Watch{
		Type: common.WatchEventType(src.EventType),
	}
	dest.Rules = make([]*sgv1.Rule, 1)
	if err = Domain2Proto(DTO(src.Object, &dest.Rules[0])); err != nil {
		return dest, err
	}
	return dest, nil
}
