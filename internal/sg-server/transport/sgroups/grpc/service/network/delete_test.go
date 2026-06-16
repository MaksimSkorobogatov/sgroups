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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type deleteTestRepo struct {
	writer repository.Writer
}

func (r *deleteTestRepo) Subject() sharedpatterns.Subject {
	return nil
}

func (r *deleteTestRepo) Writer(context.Context) (repository.Writer, error) {
	return r.writer, nil
}

func (r *deleteTestRepo) Do(ctx context.Context, fn func(repository.WriterFace) error) error {
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

func (*deleteTestRepo) Reader(context.Context) (repository.Reader, error) {
	return nil, errors.New("unexpected Reader call")
}

func (*deleteTestRepo) Close() error {
	return nil
}

type deleteTestWriter struct {
	syncErr          error
	syncNetworkCalls int
	commitCalled     bool
	abortCalled      bool
}

func (w *deleteTestWriter) SyncNamespace(context.Context, []domain.Namespace, repository.SyncOp) ([]domain.Namespace, error) {
	return nil, errors.New("unexpected SyncNamespace call")
}

func (w *deleteTestWriter) SyncAddressGroup(context.Context, []domain.AddressGroup, repository.SyncOp) ([]domain.AddressGroup, error) {
	return nil, errors.New("unexpected SyncAddressGroup call")
}

func (w *deleteTestWriter) SyncNetwork(_ context.Context, nw []domain.Network, _ repository.SyncOp) ([]domain.Network, error) {
	w.syncNetworkCalls++
	return nw, w.syncErr
}

func (w *deleteTestWriter) SyncHost(context.Context, repository.Scope, repository.SyncOp) ([]domain.Host, error) {
	return nil, errors.New("unexpected SyncHost call")
}

func (w *deleteTestWriter) SyncHostBinding(context.Context, []domain.HostBinding, repository.SyncOp) ([]domain.HostBinding, error) {
	return nil, errors.New("unexpected SyncHostBinding call")
}

func (w *deleteTestWriter) SyncNetworkBinding(context.Context, []domain.NetworkBinding, repository.SyncOp) ([]domain.NetworkBinding, error) {
	return nil, errors.New("unexpected SyncNetworkBinding call")
}

func (deleteTestWriter) SyncRules(_ context.Context, _ []domain.Rule, _ repository.SyncOp) ([]domain.Rule, error) {
	return nil, errors.New("unexpected SyncRules call")
}

func (w *deleteTestWriter) SyncService(context.Context, []domain.Service, repository.SyncOp) ([]domain.Service, error) {
	return nil, errors.New("unexpected SyncService call")
}

func (w *deleteTestWriter) SyncServiceBinding(context.Context, []domain.ServiceBinding, repository.SyncOp) ([]domain.ServiceBinding, error) {
	return nil, errors.New("unexpected SyncServiceBinding call")
}

func (w *deleteTestWriter) Commit() error {
	w.commitCalled = true
	return nil
}

func (w *deleteTestWriter) Abort() {
	w.abortCalled = true
}

func Test_Delete_MetadataOnly_Success(t *testing.T) {
	w := &deleteTestWriter{}
	srv := &networkService{
		appCtx: context.Background(),
		rep:    &deleteTestRepo{writer: w},
	}

	req := &sgv1.NetworkReq_Delete{
		Networks: []*sgv1.NetworkReq_Delete_Network{
			{
				Metadata: &common.MetadataScope{
					Name:      "nw-1",
					Namespace: "ns-1",
				},
			},
		},
	}

	resp, err := srv.Delete(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, w.syncNetworkCalls)
	require.True(t, w.commitCalled)
	require.False(t, w.abortCalled)
}

func Test_Delete_BadMetadata_InvalidArgument(t *testing.T) {
	w := &deleteTestWriter{}
	srv := &networkService{
		appCtx: context.Background(),
		rep:    &deleteTestRepo{writer: w},
	}

	req := &sgv1.NetworkReq_Delete{
		Networks: []*sgv1.NetworkReq_Delete_Network{
			{
				Metadata: &common.MetadataScope{},
			},
		},
	}

	resp, err := srv.Delete(context.Background(), req)
	require.NotNil(t, resp)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
	require.Equal(t, 0, w.syncNetworkCalls)
	require.False(t, w.commitCalled)
	require.False(t, w.abortCalled)
}

func Test_Delete_NotFound_MappedToNotFound(t *testing.T) {
	w := &deleteTestWriter{
		syncErr: repository.ErrResourceNotFound,
	}
	srv := &networkService{
		appCtx: context.Background(),
		rep:    &deleteTestRepo{writer: w},
	}

	req := &sgv1.NetworkReq_Delete{
		Networks: []*sgv1.NetworkReq_Delete_Network{
			{
				Metadata: &common.MetadataScope{
					Name:      "nw-1",
					Namespace: "ns-1",
				},
			},
		},
	}

	resp, err := srv.Delete(context.Background(), req)
	require.NotNil(t, resp)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
	require.Equal(t, 1, w.syncNetworkCalls)
	require.False(t, w.commitCalled)
	require.True(t, w.abortCalled)
}
