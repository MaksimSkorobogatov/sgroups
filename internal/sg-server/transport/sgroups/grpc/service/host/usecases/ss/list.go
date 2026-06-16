package ss

import (
	"context"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	agent "github.com/PRO-Robotech/sgroups/internal/sg-server/transport/agent/grpc"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"
	ucerr "github.com/PRO-Robotech/sgroups/internal/shared/usecases"

	"github.com/H-BF/corlib/pkg/parallel"
	agentv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/agent/v1"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type (
	// ListSS - use case for listing socket statistics
	ListSS struct {
		Rep                 repository.Repository
		AgentClientProvider agent.ClientProvider
	}
)

// Perform - performs use case
func (uc *ListSS) Perform(ctx context.Context, req *sgv1.HostReq_SocketStatistics_List) (resp *sgv1.HostResp_SocketStatistics_List, err error) {
	var rd repository.Reader
	if rd, err = uc.Rep.Reader(ctx); err != nil {
		return resp, err
	}
	defer func() { _ = rd.Close() }()

	var scopes []hostScope
	if scopes, err = resolveHostScopes(ctx, req, rd.ListHosts); err != nil {
		return resp, err
	}

	results := make([]*sgv1.HostResp_SocketStatistics_Host, len(scopes))
	errs := make([]error, len(scopes))
	_ = parallel.ExecAbstract(len(scopes), int32(len(scopes))-1, func(i int) error { //nolint:gosec
		results[i], errs[i] = uc.listAgentStats(ctx, scopes[i])
		return nil
	})
	if err = multierr.Combine(errs...); err != nil {
		return nil, err
	}
	return &sgv1.HostResp_SocketStatistics_List{Hosts: results}, nil
}

func (uc *ListSS) listAgentStats(ctx context.Context, sc hostScope) (ret *sgv1.HostResp_SocketStatistics_Host, err error) {
	defer func() {
		err = ucerr.AsInternal(err)
	}()
	var agentClient agent.Client
	if agentClient, err = uc.AgentClientProvider.New(ctx, sc.addr); err != nil {
		return nil, errors.WithMessagef(err, "create agent client for %s", sc)
	}
	defer func() { _ = agentClient.Close() }()

	var stats []*agentv1.SockStat
	err = transport.Retry(ctx, sc.String(), func() error {
		resp, e := agentClient.ListSocketStatistics(ctx, &agentv1.SocketStatReq_List{Selectors: sc.filters})
		if e != nil {
			return e
		}
		stats = resp.GetStats()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &sgv1.HostResp_SocketStatistics_Host{
		Name:      sc.host.Name.String(),
		Namespace: sc.host.Namespace.String(),
		Stats:     stats,
	}, nil
}
