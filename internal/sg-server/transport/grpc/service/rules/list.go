package rules

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/rules/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// List - returns list of rules
func (srv *rulesService) List(ctx context.Context, req *sgv1.RuleReq_List) (resp *sgv1.RuleResp_List, err error) {
	defer func() {
		err = service.CorrectError(err, service.WithAllErr[:]...)
	}()
	var rd repository.Reader
	if rd, err = srv.rep.Reader(ctx); err != nil {
		return resp, err
	}
	defer func() { _ = rd.Close() }()

	var resSelector domain.RulesSelectorList
	if err = dto.Proto2Domain(dto.DTO(req, &resSelector)); err != nil {
		return resp, err
	}
	var rulesResp domain.RuleList
	if rulesResp.Items, err = rd.ListRules(ctx, resSelector); err != nil {
		return resp, err
	}

	if rulesResp.ResourceVersion, err = rd.GetResourceVersion(ctx); err != nil {
		return resp, err
	}

	resp = new(sgv1.RuleResp_List)
	err = dto.Domain2Proto(dto.DTO(rulesResp, &resp))

	return resp, err
}
