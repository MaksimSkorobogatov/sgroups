package transport

import (
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// Validate runs the Validate method on each provided value and combines any errors.
func Validate[T interface{ Validate() error }](vals ...T) error {
	var combinedErr error
	for _, v := range vals {
		if err := v.Validate(); err != nil {
			combinedErr = multierr.Combine(combinedErr, err)
		}
	}
	if combinedErr != nil {
		return multierr.Combine(combinedErr, ErrValidate)
	}
	return nil
}

// ErrValidate is a sentinel error indicating that validation failed.
var ErrValidate = errors.New("validate")
