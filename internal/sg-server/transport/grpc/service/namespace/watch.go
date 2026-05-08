package namespace

import (
	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/grpc/service/namespace/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"google.golang.org/grpc"
)

// Watch -
func (srv *namespaceService) Watch(req *sgv1.NamespaceReq_Watch, stream grpc.ServerStreamingServer[sgv1.NamespaceResp_Watch]) (err error) {
	defer func() {
		err = service.CorrectError(err, service.WithAllErr[:]...)
	}()
	ctx, cancel := misc.GroupContext(srv.appCtx, stream.Context())
	defer cancel()

	var rd repository.Reader
	if rd, err = srv.rep.Reader(ctx); err != nil {
		return err
	}
	defer func() { _ = rd.Close() }()

	var scope repository.Scope
	if err = dto.Proto2Scope(dto.DTO(req, &scope)); err != nil {
		return err
	}

	if err = rd.WatchNamespaces(ctx, scope, func(ne domain.NamespaceEvent) error {
		var msg *sgv1.NamespaceResp_Watch
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err = dto.Domain2Proto(dto.DTO(ne, &msg)); err != nil {
			return err
		}
		return stream.Send(msg)
	}); err != nil {
		return err
	}

	return err
}
