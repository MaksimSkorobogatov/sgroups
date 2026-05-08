package dto

import (
	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository/scopes"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/dto"
	cdto "github.com/PRO-Robotech/sgroups/internal/shared/dto/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/pkg/filter"
	pb "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
)

func init() {
	dto.Register[*pb.NetworkReq_Watch, filter.Scope](watchReqToScope)
}

// Proto2Scope -
func Proto2Scope[v proto2scopeVariants](obj v) error {
	return errors.WithMessage(obj.Convert(), "proto -> scope dto convertation")
}

type proto2scopeVariants interface {
	*dto.Pair[*pb.NetworkReq_Watch, filter.Scope]
	Convert() error
}

func watchReqToScope(src *pb.NetworkReq_Watch) (dest filter.Scope, err error) {
	pbSel := src.GetSelectors()
	sels := misc.Tern(len(pbSel) > 0, make([]domain.ResSelector, len(pbSel)), nil)
	for i, sel := range pbSel {
		if err = cdto.Proto2Domain(cdto.DTO(sel, &sels[i])); err != nil {
			return dest, err
		}
	}
	dest = scopes.ScopedAnd{
		L: scopes.ScopeByResourceVersion{RV: src.GetResourceVersion()},
		R: scopes.ByResSelectors(sels...),
	}
	return dest, nil
}
