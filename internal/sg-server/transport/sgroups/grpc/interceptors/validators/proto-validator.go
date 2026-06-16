package validators

import (
	"context"

	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProtoCrossValidateUnary returns a unary interceptor that runs a proto validator
// against the incoming proto request.
func ProtoCrossValidateUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if err := dispatchProtoValidator(req); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// ProtoCrossValidateStream returns a stream interceptor that validates each
// message received from the client via RecvMsg.
func ProtoCrossValidateStream() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &structValidatingServerStream{ServerStream: ss})
	}
}

type structValidatingServerStream struct {
	grpc.ServerStream
}

// RecvMsg -
func (s *structValidatingServerStream) RecvMsg(m any) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	return dispatchProtoValidator(m)
}

func dispatchProtoValidator(req any) error {
	var err error
	switch r := req.(type) {
	case *sgv1.RuleReq_Upsert:
		err = validateRuleReqUpsert(r)
	case *sgv1.ServiceReq_Upsert:
		err = validateServiceReqUpsert(r)
	}
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}
