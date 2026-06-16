package transport

import (
	"context"
	"net/url"

	dto "github.com/PRO-Robotech/sgroups/internal/shared/dto"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"
	"github.com/PRO-Robotech/sgroups/internal/shared/usecases"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrCorrector -
type ErrCorrector func(error) (error, bool)

// WithDTOerr converts DTO-layer errors to gRPC status errors.
func WithDTOerr(e error) (error, bool) {
	if errors.Is(e, dto.ErrDTO) {
		return status.Error(codes.InvalidArgument, e.Error()), true
	}
	return e, false
}

// WithUsecaseErr converts usecase errors to gRPC status errors.
func WithUsecaseErr(e error) (error, bool) {
	if v, ok := misc.ErrTo[usecases.InvalidArgument](e); ok {
		return status.Error(codes.InvalidArgument, v.Err.Error()), true
	} else if v, ok := misc.ErrTo[usecases.InternalError](e); ok {
		return status.Error(codes.Internal, v.Err.Error()), true
	}
	return e, false
}

// WithValidationErr converts validation errors to gRPC status errors.
func WithValidationErr(e error) (error, bool) {
	if errors.Is(e, ErrValidate) {
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
