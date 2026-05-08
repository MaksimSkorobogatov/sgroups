package repository

import (
	"context"
	"errors"

	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/patterns"

	"github.com/H-BF/corlib/pkg/filter"
)

type (
	// Reader -
	Reader interface {
		readerFace
		Close() error
	}

	// Writer -
	Writer interface {
		WriterFace
		Commit() error
		Abort()
	}

	// Repository -
	Repository interface {
		Subject() patterns.Subject
		Do(ctx context.Context, fn func(WriterFace) error) error
		Writer(ctx context.Context) (Writer, error)
		Reader(ctx context.Context) (Reader, error)
		Close() error
	}

	// Scope -
	Scope = filter.Scope

	// SyncOp -
	SyncOp interface {
		syncOp()
	}

	// Delete is a SyncOp of Delete
	Delete struct{}
	// Upsert is a SyncOp Update|Insert
	Upsert struct{}
	// NoSyncOp no syncop
	NoSyncOp struct{}

	// DBUpdated -
	DBUpdated struct {
		patterns.EventType
	}
)

var (
	_ SyncOp = Delete{}
	_ SyncOp = Upsert{}
	_ SyncOp = NoSyncOp{}
)

// Convinient aliases for SyncOps
var (
	UpsertOp = Upsert{}
	DeleteOp = Delete{}
)

type (
	readerFace interface {
		GetResourceVersion(ctx context.Context) (string, error)

		ListNamespaces(ctx context.Context, rsel domain.ResSelectorList) ([]domain.Namespace, error)
		WatchNamespaces(ctx context.Context, scope Scope, cb func(domain.NamespaceEvent) error) error

		ListAddressGroups(ctx context.Context, rsel domain.ResSelectorList) ([]domain.AddressGroup, error)
		WatchAddressGroups(ctx context.Context, scope Scope, cb func(domain.AddressGroupEvent) error) error

		ListNetworks(ctx context.Context, rsel domain.ResSelectorList) ([]domain.Network, error)
		WatchNetworks(ctx context.Context, scope Scope, cb func(domain.NetworkEvent) error) error

		ListHosts(ctx context.Context, rsel domain.ResSelectorList) ([]domain.Host, error)
		WatchHosts(ctx context.Context, scope Scope, cb func(domain.HostEvent) error) error

		ListHostBindings(ctx context.Context, sel domain.HostBindingSelectorList) ([]domain.HostBinding, error)
		WatchHostBindings(ctx context.Context, scope Scope, cb func(domain.HostBindingEvent) error) error

		ListNetworkBindings(ctx context.Context, sel domain.NetworkBindingSelectorList) ([]domain.NetworkBinding, error)
		WatchNetworkBindings(ctx context.Context, scope Scope, cb func(domain.NetworkBindingEvent) error) error

		ListServices(ctx context.Context, sel domain.ResSelectorList) ([]domain.Service, error)
		WatchServices(ctx context.Context, scope Scope, cb func(domain.ServiceEvent) error) error

		ListServiceBindings(ctx context.Context, sel domain.ServiceBindingSelectorList) ([]domain.ServiceBinding, error)
		WatchServiceBindings(ctx context.Context, scope Scope, cb func(domain.ServiceBindingEvent) error) error

		ListRules(ctx context.Context, sel domain.RulesSelectorList) ([]domain.Rule, error)
		WatchRules(ctx context.Context, scope Scope, cb func(domain.RuleEvent) error) error

		GetSyncStatus(ctx context.Context) (domain.SyncStatus, error)
	}

	// WriterFace -
	WriterFace interface {
		SyncNamespace(ctx context.Context, ns []domain.Namespace, op SyncOp) ([]domain.Namespace, error)
		SyncAddressGroup(ctx context.Context, ag []domain.AddressGroup, op SyncOp) ([]domain.AddressGroup, error)
		SyncNetwork(ctx context.Context, nw []domain.Network, op SyncOp) ([]domain.Network, error)
		SyncHost(ctx context.Context, scope Scope, op SyncOp) ([]domain.Host, error)
		SyncHostBinding(ctx context.Context, hb []domain.HostBinding, op SyncOp) ([]domain.HostBinding, error)
		SyncNetworkBinding(ctx context.Context, nb []domain.NetworkBinding, op SyncOp) ([]domain.NetworkBinding, error)
		SyncRules(ctx context.Context, rules []domain.Rule, op SyncOp) ([]domain.Rule, error)
		SyncService(ctx context.Context, svc []domain.Service, op SyncOp) ([]domain.Service, error)
		SyncServiceBinding(ctx context.Context, sb []domain.ServiceBinding, op SyncOp) ([]domain.ServiceBinding, error)
	}
)

func (Delete) syncOp()   {}
func (Upsert) syncOp()   {}
func (NoSyncOp) syncOp() {}

var (
	// ErrNoRepository -
	ErrNoRepository = errors.New("no repository available")

	// ErrWriterClosed -
	ErrWriterClosed = errors.New("writer is closed")

	// ErrReaderClosed -
	ErrReaderClosed = errors.New("reader is closed")

	// ErrIsNotipml -
	ErrIsNotipml = errors.New("is not implemented")
)
