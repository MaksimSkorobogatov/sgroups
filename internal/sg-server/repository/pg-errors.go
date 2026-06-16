package repository

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
)

var (
	// ErrResourceVersionExpired means the request used an outdated resource version.
	ErrResourceVersionExpired = errors.New("resource version expired")
	// ErrResourceVersionNotFound means the requested resource version does not exist.
	ErrResourceVersionNotFound = errors.New("resource version not found")
	// ErrResourceNotFound means the requested resource does not exist.
	ErrResourceNotFound = errors.New("resource not found")
	// ErrImmutableFieldUpdate means an update attempted to change an immutable field.
	ErrImmutableFieldUpdate = errors.New("immutable field update")
	// ErrRequiredFieldMissing means a required field is missing.
	ErrRequiredFieldMissing = errors.New("required field missing")
	// ErrUpdateMismatch means an update did not match current row state.
	ErrUpdateMismatch = errors.New("update mismatch")
	// ErrDBInvariantViolation means the database signaled an invariant violation / misconfiguration.
	ErrDBInvariantViolation = errors.New("db invariant violation")
	// ErrTransportOverlap means service transport ranges overlap.
	ErrTransportOverlap = errors.New("transport overlap")
	// ErrEntriesOverlap means port/ICMP ranges across entries within a single rule overlap.
	ErrEntriesOverlap = errors.New("entries overlap")
	// ErrRuleTypeImmutable means an update targets a uid whose current rule belongs to a different rule type.
	ErrRuleTypeImmutable = errors.New("rule type immutable")
	// ErrDuplicateDisplayName means an insert/update would violate the per-namespace display_name uniqueness constraint.
	ErrDuplicateDisplayName = errors.New("duplicate display_name")
)

type repoDBError struct {
	kind error
	hint string
	code string
	err  error
}

// Error implements error interface.
func (e *repoDBError) Error() string {
	msg := e.hint
	if msg == "" {
		msg = "database error"
	}
	if e.code != "" {
		return fmt.Sprintf("%s (%s)", msg, e.code)
	}
	return msg
}

// Unwrap allows errors.Is and errors.As to work with repoDBError.
func (e *repoDBError) Unwrap() error { return e.err }

// Is allows errors.Is to match the underlying kind of repoDBError.
func (e *repoDBError) Is(target error) bool {
	return target != nil && target == e.kind
}

func correctPGError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	hint := pgErr.Hint
	if hint == "" {
		hint = pgErr.Message
	}

	var kind error
	switch pgErr.Detail {
	case "SG0001":
		kind = ErrResourceVersionExpired
	case "SG0002":
		kind = ErrResourceVersionNotFound
	case "SG0009":
		kind = ErrResourceNotFound
	case "SG0005":
		kind = ErrImmutableFieldUpdate
	case "SG0007":
		kind = ErrRequiredFieldMissing
	case "SG0008":
		kind = ErrUpdateMismatch
	case "SG0003", "SG0004", "SG0006":
		kind = ErrDBInvariantViolation
	case "SG0010":
		kind = ErrTransportOverlap
	case "SG0011":
		kind = ErrEntriesOverlap
	case "SG0012":
		kind = ErrRuleTypeImmutable
	case "SG0013":
		kind = ErrDuplicateDisplayName
	default:
		return err
	}

	return &repoDBError{kind: kind, hint: hint, code: pgErr.Detail, err: err}
}
