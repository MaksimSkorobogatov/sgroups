package rules

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/rules/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// Upsert - create or update rules
func (srv *rulesService) Upsert(ctx context.Context, req *sgv1.RuleReq_Upsert) (resp *sgv1.RuleResp_Upsert, err error) {
	defer func() {
		err = transport.CorrectError(err, service.WithAllErr[:]...)
	}()

	var rules domain.Rules
	if err = dto.Proto2Domain(dto.DTO(req, &rules)); err != nil {
		return resp, err
	}
	if err = transport.Validate(rules...); err != nil {
		return resp, err
	}
	var rlResp []domain.Rule
	err = srv.rep.Do(ctx, func(wr repository.WriterFace) (e error) {
		rlResp, e = wr.SyncRules(ctx, rules, repository.UpsertOp)
		return e
	})
	if err != nil {
		return resp, err
	}

	resp = new(sgv1.RuleResp_Upsert)
	err = dto.Domain2Proto(dto.DTO(domain.Rules(rlResp), &resp))

	return resp, err
}
