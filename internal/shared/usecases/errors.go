package usecases

type (
	// InternalError represents an internal error.
	InternalError struct {
		Err error
	}
	// InvalidArgument represents an error due to invalid arguments.
	InvalidArgument struct {
		Err error
	}
	// UnimplementedError represents an error for unimplemented features.
	UnimplementedError struct {
		Err error
	}
)

// Error implements the error interface for InternalError.
func (e InternalError) Error() string {
	return e.Err.Error()
}

// Error implements the error interface for InvalidArgument.
func (e InvalidArgument) Error() string {
	return e.Err.Error()
}

// Error implements the error interface for UnimplementedError.
func (e UnimplementedError) Error() string {
	return e.Err.Error()
}

// AsInternal converts an error to an InternalError. If the input error is nil, it returns nil.
func AsInternal(err error) error {
	if err == nil {
		return nil
	}
	return InternalError{Err: err}
}
