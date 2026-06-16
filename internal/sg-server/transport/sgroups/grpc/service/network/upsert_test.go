package network

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
	syncNetworkCalls int
	commitCalled     bool
	abortCalled      bool
}

func (w *testWriter) SyncNamespace(context.Context, []domain.Namespace, repository.SyncOp) ([]domain.Namespace, error) {
	return nil, errors.New("unexpected SyncNamespace call")
}

func (w *testWriter) SyncAddressGroup(context.Context, []domain.AddressGroup, repository.SyncOp) ([]domain.AddressGroup, error) {
	return nil, errors.New("unexpected SyncAddressGroup call")
}

func (w *testWriter) SyncNetwork(_ context.Context, nw []domain.Network, _ repository.SyncOp) ([]domain.Network, error) {
	w.syncNetworkCalls++
	return nw, nil
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

func (testWriter) SyncRules(_ context.Context, _ []domain.Rule, _ repository.SyncOp) ([]domain.Rule, error) {
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

func Test_Upsert_HostBitsCIDR_ReturnsError(t *testing.T) {
	w := &testWriter{}
	srv := &networkService{
		appCtx: context.Background(),
		rep:    &testRepo{writer: w},
	}

	req := &sgv1.NetworkReq_Upsert{
		Networks: []*sgv1.Network{
			{
				Metadata: &common.Metadata{
					Name:      "nw-1",
					Namespace: "ns-1",
				},
				Spec: &sgv1.Network_Spec{
					Cidr: "10.0.0.1/24",
				},
			},
		},
	}

	_, err := srv.Upsert(context.Background(), req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not in canonical form")
	require.Equal(t, 0, w.syncNetworkCalls)
	require.False(t, w.commitCalled)
	require.False(t, w.abortCalled)
}

func Test_Upsert_OverlappingNetworksAllowed_Success(t *testing.T) {
	w := &testWriter{}
	srv := &networkService{
		appCtx: context.Background(),
		rep:    &testRepo{writer: w},
	}

	req := &sgv1.NetworkReq_Upsert{
		Networks: []*sgv1.Network{
			{
				Metadata: &common.Metadata{
					Name:      "nw-1",
					Namespace: "ns-1",
				},
				Spec: &sgv1.Network_Spec{
					Cidr: "10.0.0.0/24",
				},
			},
			{
				Metadata: &common.Metadata{
					Name:      "nw-2",
					Namespace: "ns-1",
				},
				Spec: &sgv1.Network_Spec{
					Cidr: "10.0.0.128/25",
				},
			},
		},
	}

	resp, err := srv.Upsert(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.GetNetworks(), 2)
	require.Equal(t, "10.0.0.0/24", resp.GetNetworks()[0].GetSpec().GetCidr())
	require.Equal(t, "10.0.0.128/25", resp.GetNetworks()[1].GetSpec().GetCidr())
	require.Equal(t, 1, w.syncNetworkCalls)
	require.True(t, w.commitCalled)
	require.False(t, w.abortCalled)
}
