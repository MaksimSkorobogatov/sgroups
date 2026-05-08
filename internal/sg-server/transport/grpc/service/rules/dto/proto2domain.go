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
	dto.Register[*sgv1.Rule_Spec, domain.RuleSpec](ruleSpec2domain)
	dto.Register[*sgv1.Rule, domain.Rule](rule2domain)
	dto.Register[*sgv1.RuleReq_Upsert, domain.Rules](rules2domain)
	dto.Register[*sgv1.RuleReq_Delete_Rule, domain.Rule](ruleDeleteReq2domain)
	dto.Register[*sgv1.RuleReq_Delete, domain.Rules](rulesDeleteReq2domain)
	dto.Register[*sgv1.RuleReq_Selectors_FieldSelector, domain.RuleFieldSelector](rlFieldSelectorToDomain)
	dto.Register[*sgv1.RuleReq_Selectors, domain.RulesSelector](rlSelectorToDomain)
	dto.Register[*sgv1.RuleReq_List, domain.RulesSelectorList](rlListToDomain)
}

// Proto2Domain -
func Proto2Domain[T proto2domainVariants](a T) (err error) {
	return errors.WithMessage(a.Convert(), "common proto -> domain dto convertation")
}

type proto2domainVariants interface {
	*dto.Pair[*sgv1.Rule_Spec, domain.RuleSpec] |
		*dto.Pair[*sgv1.Rule, domain.Rule] |
		*dto.Pair[*sgv1.RuleReq_Upsert, domain.Rules] |
		*dto.Pair[*sgv1.RuleReq_Delete_Rule, domain.Rule] |
		*dto.Pair[*sgv1.RuleReq_Delete, domain.Rules] |
		*dto.Pair[*sgv1.RuleReq_Selectors_FieldSelector, domain.RuleFieldSelector] |
		*dto.Pair[*sgv1.RuleReq_Selectors, domain.RulesSelector] |
		*dto.Pair[*sgv1.RuleReq_List, domain.RulesSelectorList]
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
	if err = cdto.Proto2Domain(cdto.DTO(src.GetEndpoints().GetLocal(), &local)); err != nil {
		return dest, err
	}
	dest.Local = local

	if err = cdto.Proto2Domain(cdto.DTO(src.GetEndpoints().GetRemote(), &dest.Remote)); err != nil {
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

func rules2domain(src *sgv1.RuleReq_Upsert) (dest domain.Rules, err error) {
	dest = misc.Tern(len(src.GetRules()) > 0, make(domain.Rules, len(src.GetRules())), nil)
	for i, r := range src.GetRules() {
		if err = Proto2Domain(cdto.DTO(r, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func ruleDeleteReq2domain(src *sgv1.RuleReq_Delete_Rule) (dest domain.Rule, err error) {
	err = cdto.Proto2Domain(cdto.DTO(src.GetMetadata(), &dest.Metadata))
	return dest, err
}

func rulesDeleteReq2domain(src *sgv1.RuleReq_Delete) (dest domain.Rules, err error) {
	dest = misc.Tern(len(src.GetRules()) > 0, make(domain.Rules, len(src.GetRules())), nil)
	for i, r := range src.GetRules() {
		if err = Proto2Domain(cdto.DTO(r, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}

func protoToPtr(p common.Transport_Protocol) (*domain.IPproto, error) {
	if p == common.Transport_PROTOCOL_UNDEF {
		return nil, nil
	}
	var v domain.IPproto
	if err := cdto.Proto2Domain(cdto.DTO(p, &v)); err != nil {
		return nil, err
	}
	return &v, nil
}

func trafficToPtr(t common.Session_Traffic) (ret *domain.Traffic, err error) {
	switch t {
	case common.Session_TRAFFIC_UNDEF:
	default:
		ret = new(domain.Traffic)
		if err = cdto.Proto2Domain(cdto.DTO(t, ret)); err != nil {
			return nil, err
		}
	}
	return ret, err
}

func rlFieldSelectorToDomain(src *sgv1.RuleReq_Selectors_FieldSelector) (dst domain.RuleFieldSelector, err error) {
	var (
		proto   *domain.IPproto
		traffic *domain.Traffic
	)
	proto, err = protoToPtr(src.GetProtocol())
	if err != nil {
		return dst, err
	}
	traffic, err = trafficToPtr(src.GetTraffic())
	if err != nil {
		return dst, err
	}
	dst = domain.RuleFieldSelector{
		ResourceIdentifier: domain.ResourceIdentifier{
			Name:      domain.ResourceName(src.GetName()),
			Namespace: domain.ResourceNamespace(src.GetNamespace()),
		},
		Traffic: traffic,
		Proto:   proto,
	}

	if src.GetEndpoints().GetLocal() != nil {
		var local domain.EpLocal
		err = cdto.Proto2Domain(cdto.DTO(src.GetEndpoints().GetLocal(), &local))
		if err != nil {
			return dst, err
		}
		dst.Local = local
	}

	if src.GetEndpoints().GetRemote() != nil {
		var remote domain.EndpointSpec
		err = cdto.Proto2Domain(cdto.DTO(src.GetEndpoints().GetRemote(), &remote))
		if err != nil {
			return dst, err
		}
		dst.Remote = remote
	}
	return dst, err
}

func rlSelectorToDomain(src *sgv1.RuleReq_Selectors) (dst domain.RulesSelector, err error) {
	dst = domain.RulesSelector{
		LabelSelector: src.GetLabelSelector(),
	}
	err = Proto2Domain(DTO(src.GetFieldSelector(), &dst.FieldSelector))
	return dst, err
}

func rlListToDomain(src *sgv1.RuleReq_List) (dest domain.RulesSelectorList, err error) {
	dest = misc.Tern(len(src.GetSelectors()) > 0, make(domain.RulesSelectorList, len(src.GetSelectors())), nil)
	for i, sel := range src.GetSelectors() {
		if err = Proto2Domain(DTO(sel, &dest[i])); err != nil {
			return dest, err
		}
	}
	return dest, nil
}
