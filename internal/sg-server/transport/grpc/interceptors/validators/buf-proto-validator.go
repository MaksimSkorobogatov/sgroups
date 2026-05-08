package validators

import (
	"context"
	"errors"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type validatingServerStream struct {
	grpc.ServerStream

	cfg protoValidateCfg
}

// BufProtovalidateUnary returns a unary interceptor that validates incoming requests.
// If a custom validator is not provided, protovalidate.Validate (global validator) is used.
func BufProtovalidateUnary(o ...ProtoValidateOption) grpc.UnaryServerInterceptor {
	cfg := newProtoValidateCfg(o...)

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if msg, ok := req.(proto.Message); ok {
			if err := validateProtoMessage(cfg, msg); err != nil {
				return nil, err
			}
		}
		return handler(ctx, req)
	}
}

// BufProtovalidateStream returns a stream interceptor that validates each incoming message
// received via RecvMsg.
func BufProtovalidateStream(o ...ProtoValidateOption) grpc.StreamServerInterceptor {
	cfg := newProtoValidateCfg(o...)

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &validatingServerStream{
			ServerStream: ss,
			cfg:          cfg,
		})
	}
}

// RecvMsg intercepts incoming stream messages and validates them.
func (s *validatingServerStream) RecvMsg(m any) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	msg, ok := m.(proto.Message)
	if !ok {
		return nil
	}
	return validateProtoMessage(s.cfg, msg)
}

func newProtoValidateCfg(o ...ProtoValidateOption) protoValidateCfg {
	var cfg protoValidateCfg
	for _, apply := range o {
		if apply == nil {
			continue
		}
		apply(&cfg)
	}
	return cfg
}

func validateProtoMessage(cfg protoValidateCfg, msg proto.Message) error {
	var err error
	if cfg.v == nil {
		err = protovalidate.Validate(msg, cfg.opts...)
	} else {
		err = cfg.v.Validate(msg, cfg.opts...)
	}

	return protovalidateStatusError(err)
}

func protovalidateStatusError(err error) error {
	if err == nil {
		return nil
	}
	var valErr *protovalidate.ValidationError
	if !errors.As(err, &valErr) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	st := status.New(codes.InvalidArgument, "validation failed")
	if stWithDetails, dErr := st.WithDetails(valErr.ToProto()); dErr == nil {
		return stWithDetails.Err()
	}
	return status.Error(codes.InvalidArgument, valErr.Error())
}
