package dto

import (
	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository/scopes"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/pkg/filter"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
)

func init() {
	dto.Register[*sgv1.RuleReq_Watch, filter.Scope](watchReqToScope)
}

// Proto2Scope -
func Proto2Scope[v proto2scopeVariants](obj v) error {
	return errors.WithMessage(obj.Convert(), "proto -> scope dto convertation")
}

type proto2scopeVariants interface {
	*dto.Pair[*sgv1.RuleReq_Watch, filter.Scope]
	Convert() error
}

func watchReqToScope(src *sgv1.RuleReq_Watch) (dest filter.Scope, err error) {
	pbSel := src.GetSelectors()
	sels := misc.Tern(len(pbSel) > 0, make([]domain.RulesSelector, len(pbSel)), nil)
	for i, sel := range pbSel {
		if err = Proto2Domain(DTO(sel, &sels[i])); err != nil {
			return dest, err
		}
	}
	dest = scopes.ScopedAnd{
		L: scopes.ScopeByResourceVersion{RV: src.GetResourceVersion()},
		R: scopes.ByRulesSelectors(sels...),
	}
	return dest, nil
}
