package service

import (
	"fmt"
	"path"

	"github.com/H-BF/corlib/pkg/patterns"
	"github.com/H-BF/corlib/server"
	"google.golang.org/grpc"
)

// Service -
type Service interface {
	server.APIService
	server.APIGatewayProxy
}

// RegisterGRPC registers the service with the gRPC server, applying any additional options for path prefixes or service names.
func RegisterGRPC(srv Service, s *grpc.Server, opts ...Option) error {
	desc := srv.Description()
	s.RegisterService(&desc, srv)

	var op CommonOptions
	op.ApplyOptions(opts...)

	if len(op.PathPrefixes) > 0 {
		flt := map[string]struct{}{}
		var p patterns.Path
		for _, pt := range op.PathPrefixes {
			if e := p.Set(pt); e != nil {
				return fmt.Errorf("SgService: register GRPC API '%T' with additional path prefix: %v", srv, e)
			}
			if !p.IsEmpty() {
				ss := p.String()
				if _, seen := flt[ss]; !seen {
					flt[ss] = struct{}{}
					desc1 := desc
					desc1.ServiceName = path.Join(ss, desc.ServiceName)
					s.RegisterService(&desc1, srv)
				}
			}
		}
	}
	if len(op.AdditionalServiceNames) > 0 {
		for _, sn := range op.AdditionalServiceNames {
			desc1 := desc
			desc1.ServiceName = sn
			s.RegisterService(&desc1, srv)
		}
	}
	return nil
}
