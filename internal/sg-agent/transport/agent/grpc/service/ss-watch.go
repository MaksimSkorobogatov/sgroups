package service

import (
	uc "github.com/PRO-Robotech/sgroups/internal/sg-agent/usecases/ss"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	agentv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/agent/v1"
	"google.golang.org/grpc"
)

// WatchSocketStatistics -
func (srv *agentAPI) WatchSocketStatistics(req *agentv1.SocketStatReq_Watch, stream grpc.ServerStreamingServer[agentv1.SocketStatResp_Watch]) (err error) {
	defer func() {
		err = transport.CorrectError(err, transport.WithValidationErr, transport.WithUsecaseErr)
	}()
	if err = ValidateSelectors(req.GetSelectors()); err != nil {
		return err
	}
	u := uc.NewWatchUseCase(stream, srv.ssCacheTTL)
	return u.Watch(srv.appCtx, req)
}
