package nft

import (
	"context"
	"io"
	"sync"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	agent "github.com/PRO-Robotech/sgroups/internal/sg-server/transport/agent/grpc"
	ucerr "github.com/PRO-Robotech/sgroups/internal/shared/usecases"

	"github.com/H-BF/corlib/pkg/parallel"
	agentv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/agent/v1"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
)

type (
	// WatchSS - use case for watching nft
	WatchSS struct {
		Rep                 repository.Repository
		AgentClientProvider agent.ClientProvider
		OutStream           grpc.ServerStreamingServer[sgv1.HostResp_Nft_Watch]
		sendMu              sync.Mutex
	}
)

// Perform - performs use case
func (uc *WatchSS) Perform(ctx context.Context, req *sgv1.HostReq_Nft_Watch) (err error) {
	var rd repository.Reader
	if rd, err = uc.Rep.Reader(ctx); err != nil {
		return err
	}
	defer func() { _ = rd.Close() }()

	var scopes []hostScope
	if scopes, err = resolveHostScopes(ctx, req, rd.ListHosts); err != nil {
		return err
	}
	errs := make([]error, len(scopes))
	_ = parallel.ExecAbstract(len(scopes), int32(len(scopes))-1, func(i int) error { //nolint:gosec
		errs[i] = uc.watchNft(ctx, scopes[i])
		return nil
	})

	return multierr.Combine(errs...)
}

func (uc *WatchSS) watchNft(ctx context.Context, sc hostScope) (err error) {
	defer func() {
		err = ucerr.AsInternal(err)
	}()
	var agentClient agent.Client
	if agentClient, err = uc.AgentClientProvider.New(ctx, sc.addr); err != nil {
		return errors.WithMessagef(err, "create agent client for %s", sc)
	}
	defer func() { _ = agentClient.Close() }()

	stream, e := agentClient.WatchNftables(ctx, new(agentv1.NftablesReq_Watch))
	if e != nil {
		return errors.WithMessagef(e, "open watch stream for %s", sc)
	}

	var resp *agentv1.NftablesResp_Watch
	for err == nil {
		resp, err = stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return errors.WithMessagef(err, "recv watch stream from %s", sc)
		}
		uc.sendMu.Lock()
		err = uc.OutStream.Send(&sgv1.HostResp_Nft_Watch{
			Hosts: []*sgv1.HostResp_Nft_Host{
				{
					Name:      sc.host.Name.String(),
					Namespace: sc.host.Namespace.String(),
					Nft:       resp.GetNftables(),
				},
			},
		})
		uc.sendMu.Unlock()
	}

	return err
}
