package service

import (
	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/shared/transport"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WithAllErr is a collection of all error correctors.
var WithAllErr = [...]transport.ErrCorrector{WithDbErr, transport.WithDTOerr, transport.WithValidationErr, transport.WithUsecaseErr}

// WithDbErr converts DB-layer repository errors to gRPC status errors.
func WithDbErr(e error) (error, bool) {
	switch {
	case errors.Is(e, repository.ErrResourceVersionExpired):
		return status.Error(codes.FailedPrecondition, e.Error()), true
	case errors.Is(e, repository.ErrResourceVersionNotFound):
		return status.Error(codes.NotFound, e.Error()), true
	case errors.Is(e, repository.ErrResourceNotFound):
		return status.Error(codes.NotFound, e.Error()), true
	case errors.Is(e, repository.ErrImmutableFieldUpdate):
		return status.Error(codes.InvalidArgument, e.Error()), true
	case errors.Is(e, repository.ErrRequiredFieldMissing):
		return status.Error(codes.InvalidArgument, e.Error()), true
	case errors.Is(e, repository.ErrUpdateMismatch):
		return status.Error(codes.FailedPrecondition, e.Error()), true
	case errors.Is(e, repository.ErrDBInvariantViolation):
		return status.Error(codes.Internal, e.Error()), true
	case errors.Is(e, repository.ErrTransportOverlap):
		return status.Error(codes.AlreadyExists, e.Error()), true
	case errors.Is(e, repository.ErrEntriesOverlap):
		return status.Error(codes.InvalidArgument, e.Error()), true
	case errors.Is(e, repository.ErrRuleTypeImmutable):
		return status.Error(codes.InvalidArgument, e.Error()), true
	case errors.Is(e, repository.ErrDuplicateDisplayName):
		return status.Error(codes.AlreadyExists, e.Error()), true
	default:
		return e, false
	}
}
