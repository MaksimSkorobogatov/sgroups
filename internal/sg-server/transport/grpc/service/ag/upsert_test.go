package ag

import (
	"context"
	"errors"
	"testing"

	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	sharedpatterns "github.com/PRO-Robotech/sgroups/internal/shared/patterns"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type testRepo struct {
	writer repository.Writer
}

func (r *testRepo) Subject() sharedpatterns.Subject {
	return nil
}

func (r *testRepo) Writer(context.Context) (repository.Writer, error) {
	return r.writer, nil
}

func (r *testRepo) Do(ctx context.Context, fn func(repository.WriterFace) error) error {
	wr, err := r.Writer(ctx)
	if err != nil {
		return err
	}
	if err = fn(wr); err != nil {
		wr.Abort()
		return err
	}
	return wr.Commit()
}

func (*testRepo) Reader(context.Context) (repository.Reader, error) {
	return nil, errors.New("unexpected Reader call")
}

func (*testRepo) Close() error {
	return nil
}

type testWriter struct {
	syncAGCalls  int
	commitCalled bool
	abortCalled  bool
}

func (w *testWriter) SyncNamespace(context.Context, []domain.Namespace, repository.SyncOp) ([]domain.Namespace, error) {
	return nil, errors.New("unexpected SyncNamespace call")
}

func (w *testWriter) SyncAddressGroup(_ context.Context, ag []domain.AddressGroup, _ repository.SyncOp) ([]domain.AddressGroup, error) {
	w.syncAGCalls++
	return ag, nil
}

func (w *testWriter) SyncNetwork(context.Context, []domain.Network, repository.SyncOp) ([]domain.Network, error) {
	return nil, errors.New("unexpected SyncNetwork call")
}

func (w *testWriter) SyncHost(context.Context, repository.Scope, repository.SyncOp) ([]domain.Host, error) {
	return nil, errors.New("unexpected SyncHost call")
}

func (w *testWriter) SyncHostBinding(context.Context, []domain.HostBinding, repository.SyncOp) ([]domain.HostBinding, error) {
	return nil, errors.New("unexpected SyncHostBinding call")
}

func (w *testWriter) SyncNetworkBinding(context.Context, []domain.NetworkBinding, repository.SyncOp) ([]domain.NetworkBinding, error) {
	return nil, errors.New("unexpected SyncNetworkBinding call")
}

func (w *testWriter) SyncRules(context.Context, []domain.Rule, repository.SyncOp) ([]domain.Rule, error) {
	return nil, errors.New("unexpected SyncRules call")
}

func (w *testWriter) SyncService(context.Context, []domain.Service, repository.SyncOp) ([]domain.Service, error) {
	return nil, errors.New("unexpected SyncService call")
}

func (w *testWriter) SyncServiceBinding(context.Context, []domain.ServiceBinding, repository.SyncOp) ([]domain.ServiceBinding, error) {
	return nil, errors.New("unexpected SyncServiceBinding call")
}

func (w *testWriter) Commit() error {
	w.commitCalled = true
	return nil
}

func (w *testWriter) Abort() {
	w.abortCalled = true
}

// Test_Upsert_InvalidDefaultAction_ShouldFail demonstrates that an AG with
// an out-of-range DefaultAction (e.g. 999) passes validation and reaches
// the database layer — which is a bug. The Upsert handler does not validate
// AgSpec, so ChainDefaultAction.Validate() is never called.
//
// EXPECTED: InvalidArgument error — validation should reject value 999.
// ACTUAL:   No error — the invalid AG reaches SyncAddressGroup.
func Test_Upsert_InvalidDefaultAction_ShouldFail(t *testing.T) {
	w := &testWriter{}
	srv := &addressGroupService{
		appCtx: context.Background(),
		rep:    &testRepo{writer: w},
	}

	req := &sgv1.AddressGroupReq_Upsert{
		AddressGroups: []*sgv1.AddressGroup{
			{
				Metadata: &common.Metadata{
					Name:      "ag-1",
					Namespace: "ns-1",
				},
				Spec: &sgv1.AddressGroup_Spec{
					DefaultAction: 999, // invalid — valid values are 0 (UNKNOWN), 1 (ALLOW), 2 (DENY)
				},
			},
		},
	}

	resp, err := srv.Upsert(context.Background(), req)

	require.Nil(t, resp)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
	// require.Equal(t, 0, w.syncAGCalls, "invalid AG should not reach the DB layer")
	// require.False(t, w.commitCalled)
	// require.True(t, w.abortCalled)

	_ = codes.InvalidArgument // keep import for when fix is applied
	_ = status.Code           // keep import for when fix is applied
}
