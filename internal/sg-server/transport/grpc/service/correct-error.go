package service

import (
	"context"
	"net/url"

	repository "github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	dto "github.com/PRO-Robotech/sgroups/internal/shared/dto"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrCorrector -
type ErrCorrector func(error) (error, bool)

// WithAllErr is a collection of all error correctors.
var WithAllErr = [...]ErrCorrector{WithDbErr, WithDTOerr, WithValidationErr}

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
	default:
		return e, false
	}
}

// WithDTOerr converts DTO-layer errors to gRPC status errors.
func WithDTOerr(e error) (error, bool) {
	if errors.Is(e, dto.ErrDTO) {
		return status.Error(codes.InvalidArgument, e.Error()), true
	}
	return e, false
}

// WithValidationErr converts validation errors to gRPC status errors.
func WithValidationErr(e error) (error, bool) {
	if errors.Is(e, errValidate) {
		return status.Error(codes.InvalidArgument, e.Error()), true
	}
	return e, false
}

// CorrectError applies a series of correctors to an error, returning the first corrected error.
func CorrectError(e error, correctors ...ErrCorrector) error {
	if e == nil {
		return nil
	}
	for _, c := range correctors {
		if c == nil {
			continue
		}
		if err, ok := c(e); ok {
			return err
		}
	}
	return DefCorrectError(e)
}

// DefCorrectError is the default error corrector that converts common errors to gRPC status errors.
func DefCorrectError(err error) error {
	if err != nil && status.Code(err) == codes.Unknown {
		switch errors.Cause(err) {
		case context.DeadlineExceeded:
			return status.New(codes.DeadlineExceeded, err.Error()).Err()
		case context.Canceled:
			return status.New(codes.Canceled, err.Error()).Err()
		default:
			if e := new(url.Error); errors.As(err, &e) {
				switch errors.Cause(e.Err) {
				case context.Canceled:
					return status.New(codes.Canceled, err.Error()).Err()
				case context.DeadlineExceeded:
					return status.New(codes.DeadlineExceeded, err.Error()).Err()
				default:
					if e.Timeout() {
						return status.New(codes.DeadlineExceeded, err.Error()).Err()
					}
				}
			}
			err = status.New(codes.Internal, err.Error()).Err()
		}
	}
	return err
}
