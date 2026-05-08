package repository

import (
	"context"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
)

type (
	notImplReaderFace struct{}
	notImplWriterFace struct{}
)

var (
	_ readerFace = (*notImplReaderFace)(nil)
	_ WriterFace = (*notImplWriterFace)(nil)
)

func (notImplReaderFace) GetResourceVersion(_ context.Context) (string, error) {
	return "", ErrIsNotipml
}

// ListNamespaces returns ErrIsNotipml error
func (notImplReaderFace) ListNamespaces(_ context.Context, _ domain.ResSelectorList) ([]domain.Namespace, error) {
	return nil, ErrIsNotipml
}

// WatchNamespaces returns ErrIsNotipml error
func (notImplReaderFace) WatchNamespaces(_ context.Context, _ Scope, _ func(domain.NamespaceEvent) error) error {
	return ErrIsNotipml
}

// SyncNamespace returns ErrIsNotipml error
func (notImplWriterFace) SyncNamespace(_ context.Context, _ []domain.Namespace, _ SyncOp) ([]domain.Namespace, error) {
	return nil, ErrIsNotipml
}

// ListAddressGroups returns ErrIsNotipml error
func (notImplReaderFace) ListAddressGroups(_ context.Context, _ domain.ResSelectorList) ([]domain.AddressGroup, error) {
	return nil, ErrIsNotipml
}

// WatchAddressGroups returns ErrIsNotipml error
func (notImplReaderFace) WatchAddressGroups(_ context.Context, _ Scope, _ func(domain.AddressGroupEvent) error) error {
	return ErrIsNotipml
}

// SyncAddressGroup returns ErrIsNotipml error
func (notImplWriterFace) SyncAddressGroup(_ context.Context, _ []domain.AddressGroup, _ SyncOp) ([]domain.AddressGroup, error) {
	return nil, ErrIsNotipml
}

// ListNetworks returns ErrIsNotipml error
func (notImplReaderFace) ListNetworks(_ context.Context, _ domain.ResSelectorList) ([]domain.Network, error) {
	return nil, ErrIsNotipml
}

// WatchNetworks returns ErrIsNotipml error
func (notImplReaderFace) WatchNetworks(_ context.Context, _ Scope, _ func(domain.NetworkEvent) error) error {
	return ErrIsNotipml
}

// SyncNetwork returns ErrIsNotipml error
func (notImplWriterFace) SyncNetwork(_ context.Context, _ []domain.Network, _ SyncOp) ([]domain.Network, error) {
	return nil, ErrIsNotipml
}

// ListHosts returns ErrIsNotipml error
func (notImplReaderFace) ListHosts(_ context.Context, _ domain.ResSelectorList) ([]domain.Host, error) {
	return nil, ErrIsNotipml
}

// WatchHosts returns ErrIsNotipml error
func (notImplReaderFace) WatchHosts(_ context.Context, _ Scope, _ func(domain.HostEvent) error) error {
	return ErrIsNotipml
}

// SyncHost returns ErrIsNotipml error
func (notImplWriterFace) SyncHost(_ context.Context, _ Scope, _ SyncOp) ([]domain.Host, error) {
	return nil, ErrIsNotipml
}

// ListHostBindings returns ErrIsNotipml error
func (notImplReaderFace) ListHostBindings(ctx context.Context, sel domain.HostBindingSelectorList) ([]domain.HostBinding, error) {
	return nil, ErrIsNotipml
}

// WatchHostBindings returns ErrIsNotipml error
func (notImplReaderFace) WatchHostBindings(_ context.Context, _ Scope, _ func(domain.HostBindingEvent) error) error {
	return ErrIsNotipml
}

// SyncHostBinding returns ErrIsNotipml error
func (notImplWriterFace) SyncHostBinding(_ context.Context, _ []domain.HostBinding, _ SyncOp) ([]domain.HostBinding, error) {
	return nil, ErrIsNotipml
}

// ListNetworkBindings returns ErrIsNotipml error
func (notImplReaderFace) ListNetworkBindings(_ context.Context, _ domain.NetworkBindingSelectorList) ([]domain.NetworkBinding, error) {
	return nil, ErrIsNotipml
}

// WatchNetworkBindings returns ErrIsNotipml error
func (notImplReaderFace) WatchNetworkBindings(_ context.Context, _ Scope, _ func(domain.NetworkBindingEvent) error) error {
	return ErrIsNotipml
}

// SyncNetworkBinding returns ErrIsNotipml error
func (notImplWriterFace) SyncNetworkBinding(_ context.Context, _ []domain.NetworkBinding, _ SyncOp) ([]domain.NetworkBinding, error) {
	return nil, ErrIsNotipml
}

// ListRules returns ErrIsNotipml error
func (notImplReaderFace) ListRules(_ context.Context, _ domain.RulesSelectorList) ([]domain.Rule, error) {
	return nil, ErrIsNotipml
}

// WatchRules returns ErrIsNotipml error
func (notImplReaderFace) WatchRules(_ context.Context, _ Scope, _ func(domain.RuleEvent) error) error {
	return ErrIsNotipml
}

// SyncRules returns ErrIsNotipml error
func (notImplWriterFace) SyncRules(_ context.Context, _ []domain.Rule, _ SyncOp) ([]domain.Rule, error) {
	return nil, ErrIsNotipml
}

// ListServices returns ErrIsNotipml error
func (notImplReaderFace) ListServices(_ context.Context, _ domain.ResSelectorList) ([]domain.Service, error) {
	return nil, ErrIsNotipml
}

// WatchServices returns ErrIsNotipml error
func (notImplReaderFace) WatchServices(_ context.Context, _ Scope, _ func(domain.ServiceEvent) error) error {
	return ErrIsNotipml
}

// SyncService returns ErrIsNotipml error
func (notImplWriterFace) SyncService(_ context.Context, _ []domain.Service, _ SyncOp) ([]domain.Service, error) {
	return nil, ErrIsNotipml
}

// ListServiceBindings returns ErrIsNotipml error
func (notImplReaderFace) ListServiceBindings(_ context.Context, _ domain.ServiceBindingSelectorList) ([]domain.ServiceBinding, error) {
	return nil, ErrIsNotipml
}

// WatchServiceBindings returns ErrIsNotipml error
func (notImplReaderFace) WatchServiceBindings(_ context.Context, _ Scope, _ func(domain.ServiceBindingEvent) error) error {
	return ErrIsNotipml
}

// SyncServiceBinding returns ErrIsNotipml error
func (notImplWriterFace) SyncServiceBinding(_ context.Context, _ []domain.ServiceBinding, _ SyncOp) ([]domain.ServiceBinding, error) {
	return nil, ErrIsNotipml
}

// GetSyncStatus returns ErrIsNotipml error
func (notImplReaderFace) GetSyncStatus(_ context.Context) (ret domain.SyncStatus, err error) {
	return ret, ErrIsNotipml
}
