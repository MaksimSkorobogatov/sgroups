package repository

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func Test_correctPGError_MapsDetailToRepositoryErrors(t *testing.T) {
	cases := []struct {
		name   string
		detail string
		hint   string
		want   error
	}{
		{name: "rv-expired", detail: "SG0001", hint: "rv expired", want: ErrResourceVersionExpired},
		{name: "rv-not-found", detail: "SG0002", hint: "rv not found", want: ErrResourceVersionNotFound},
		{name: "immutable", detail: "SG0005", hint: "immutable", want: ErrImmutableFieldUpdate},
		{name: "required", detail: "SG0007", hint: "required", want: ErrRequiredFieldMissing},
		{name: "mismatch", detail: "SG0008", hint: "mismatch", want: ErrUpdateMismatch},
		{name: "invariant", detail: "SG0003", hint: "bad trigger", want: ErrDBInvariantViolation},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			orig := &pgconn.PgError{Detail: tc.detail, Hint: tc.hint, Message: "msg"}
			got := correctPGError(orig)
			require.Error(t, got)
			require.True(t, errors.Is(got, tc.want))
			require.Contains(t, got.Error(), tc.hint)
			require.Same(t, orig, errors.Unwrap(got))
		})
	}
}

func Test_correctPGError_PassThroughUnknownDetail(t *testing.T) {
	orig := &pgconn.PgError{Detail: "SOMETHING", Hint: "x"}
	got := correctPGError(orig)
	require.Same(t, orig, got)
}
