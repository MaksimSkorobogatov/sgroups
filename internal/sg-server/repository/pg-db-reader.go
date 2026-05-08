package repository

import (
	"context"
	"slices"
	"strconv"

	pg "github.com/PRO-Robotech/sgroups/internal/sg-server/repository/pg/domain"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository/pg/dto"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository/pg/lister"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository/scopes"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/pkg/parallel"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

var _ Reader = (*pgDbReader)(nil)

type pgDbReader struct {
	notImplReaderFace //nolint:unused
	doIt              func(context.Context, func(pgConn) error) error
	close             func()
}

// Close impl Reader interface
func (rd *pgDbReader) Close() error {
	if rd.close != nil {
		rd.close()
	}
	return nil
}

// GetSyncStatus impl Reader interface
func (rd *pgDbReader) GetSyncStatus(ctx context.Context) (ret domain.SyncStatus, err error) {
	const (
		api = "GetSyncStatus"
		qry = "select updated_at, sync_count from sgroups.tbl_sync_status where id = (select max(id) from sgroups.tbl_sync_status)"
	)
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	var statusLister lister.ListerWatcher[domain.SyncStatus, pg.SyncStatus]
	if err = statusLister.Init(scopes.NoScope, qry); err != nil {
		return ret, err
	}
	var ss domain.SyncStatus
	err = rd.doIt(ctx, func(c pgConn) error {
		var e error
		ss, e = statusLister.Get(ctx, c.Query, func(s pg.SyncStatus) (domain.SyncStatus, error) {
			return domain.SyncStatus{
				UpdatedAt: s.Updated,
			}, nil
		})
		return e
	})
	return ss, err
}

// GetResourceVersion -
func (rd *pgDbReader) GetResourceVersion(ctx context.Context) (ret string, err error) {
	const (
		api = "GetResourceVersion"
		qry = "select * from sgroups.get_resource_version()"
	)
	defer func() {
		err = errors.WithMessage(err, api)
	}()
	var rvLister lister.ListerWatcher[string, string]
	if err = rvLister.Init(scopes.NoScope, qry); err != nil {
		return "", err
	}
	var rv string
	err = rd.doIt(ctx, func(c pgConn) error {
		var e error
		rv, e = rvLister.Get(ctx, c.Query, func(s string) (string, error) {
			return s, nil
		})
		return e
	})
	return rv, err
}

// ListNamespaces -
func (rd *pgDbReader) ListNamespaces(ctx context.Context, rsel domain.ResSelectorList) (ret []domain.Namespace, err error) {
	const (
		api  = "ListNamespaces"
		from = "sgroups.list_namespaces($1)"
	)
	return list(ctx, api, from, scopes.ByResSelectors(rsel...), rd.doIt,
		func(pgObj pg.Namespace) (domain.Namespace, error) {
			var dom domain.Namespace
			err := dto.Pg2Domain(dto.DTO(pgObj, &dom))
			return dom, err
		},
	)
}

// WatchNamespaces -
func (rd *pgDbReader) WatchNamespaces(ctx context.Context, scope Scope, cb func(domain.NamespaceEvent) error) (err error) {
	const (
		api = "WatchNamespaces"
		ch  = "resource_Namespace"
	)
	return watch(ctx, api, ch, scope, domain.NamespaceResource, rd.doIt,
		func(pgObj pg.ResourceEvent) (domain.NamespaceEvent, error) {
			var dom domain.NamespaceEvent
			err := dto.Pg2Domain(dto.DTO(pgObj, &dom))
			return dom, err
		},
		cb,
	)
}

// ListAddressGroups -
func (rd *pgDbReader) ListAddressGroups(ctx context.Context, rsel domain.ResSelectorList) (ret []domain.AddressGroup, err error) {
	const (
		api  = "ListAddressGroups"
		from = "sgroups.list_ag($1)"
	)
	return list(ctx, api, from, scopes.ByResSelectors(rsel...), rd.doIt,
		func(pgObj pg.AddressGroup) (domain.AddressGroup, error) {
			var dom domain.AddressGroup
			err := dto.Pg2Domain(dto.DTO(pgObj, &dom))
			return dom, err
		},
	)
}

// WatchAddressGroups -
func (rd *pgDbReader) WatchAddressGroups(ctx context.Context, scope Scope, cb func(domain.AddressGroupEvent) error) (err error) {
	const (
		api = "WatchAddressGroups"
		ch  = "resource_AddressGroup"
	)
	return watch(ctx, api, ch, scope, domain.AddressGroupResource, rd.doIt,
		func(pgObj pg.ResourceEvent) (domain.AddressGroupEvent, error) {
			var dom domain.AddressGroupEvent
			err := dto.Pg2Domain(dto.DTO(pgObj, &dom))
			return dom, err
		},
		cb,
	)
}

// ListNetworks -
func (rd *pgDbReader) ListNetworks(ctx context.Context, rsel domain.ResSelectorList) (ret []domain.Network, err error) {
	const (
		api  = "ListNetworks"
		from = "sgroups.list_networks($1)"
	)
	return list(ctx, api, from, scopes.ByResSelectors(rsel...), rd.doIt,
		func(pgObj pg.Network) (domain.Network, error) {
			var dom domain.Network
			err := dto.Pg2Domain(dto.DTO(pgObj, &dom))
			return dom, err
		},
	)
}

// WatchNetworks -
func (rd *pgDbReader) WatchNetworks(ctx context.Context, scope Scope, cb func(domain.NetworkEvent) error) (err error) {
	const (
		api = "WatchNetworks"
		ch  = "resource_Network"
	)
	return watch(ctx, api, ch, scope, domain.NetworkResource, rd.doIt,
		func(pgObj pg.ResourceEvent) (domain.NetworkEvent, error) {
			var dom domain.NetworkEvent
			err := dto.Pg2Domain(dto.DTO(pgObj, &dom))
			return dom, err
		},
		cb,
	)
}

// ListHosts -
func (rd *pgDbReader) ListHosts(ctx context.Context, rsel domain.ResSelectorList) (ret []domain.Host, err error) {
	const (
		api  = "ListHosts"
		from = "sgroups.list_hosts($1)"
	)
	return list(ctx, api, from, scopes.ByResSelectors(rsel...), rd.doIt,
		func(pgObj pg.Host) (domain.Host, error) {
			var dom domain.Host
			err := dto.Pg2Domain(dto.DTO(pgObj, &dom))
			return dom, err
		},
	)
}

// WatchHosts -
func (rd *pgDbReader) WatchHosts(ctx context.Context, scope Scope, cb func(domain.HostEvent) error) (err error) {
	const (
		api = "WatchHosts"
		ch  = "resource_Host"
	)
	return watch(ctx, api, ch, scope, domain.HostResource, rd.doIt,
		func(pgObj pg.ResourceEvent) (domain.HostEvent, error) {
			var dom domain.HostEvent
			err := dto.Pg2Domain(dto.DTO(pgObj, &dom))
			return dom, err
		},
		cb,
	)
}

// ListHostBindings -
func (rd *pgDbReader) ListHostBindings(ctx context.Context, sel domain.HostBindingSelectorList) (ret []domain.HostBinding, err error) {
	const (
		api  = "ListHostBindings"
		from = "sgroups.list_host_bindings()"
	)
	return list(ctx, api, from, scopes.ByHostBindingSelectors(sel...), rd.doIt,
		func(pgObj pg.HostBinding) (domain.HostBinding, error) {
			var dom domain.HostBinding
			err := dto.Pg2Domain(dto.DTO(pgObj, &dom))
			return dom, err
		},
	)
}

// WatchHostBindings -
func (rd *pgDbReader) WatchHostBindings(ctx context.Context, scope Scope, cb func(domain.HostBindingEvent) error) (err error) {
	const (
		api = "WatchHostBindings"
		ch  = "resource_HostBinding"
	)
	return watch(ctx, api, ch, scope, domain.HostBindingResource, rd.doIt,
		func(pgObj pg.ResourceEvent) (domain.HostBindingEvent, error) {
			var dom domain.HostBindingEvent
			err := dto.Pg2Domain(dto.DTO(pgObj, &dom))
			return dom, err
		},
		cb,
	)
}

// ListNetworkBindings -
func (rd *pgDbReader) ListNetworkBindings(ctx context.Context, sel domain.NetworkBindingSelectorList) (ret []domain.NetworkBinding, err error) {
	const (
		api  = "ListNetworkBindings"
		from = "sgroups.list_network_bindings()"
	)
	return list(ctx, api, from, scopes.ByNetworkBindingSelectors(sel...), rd.doIt,
		func(pgObj pg.NetworkBinding) (domain.NetworkBinding, error) {
			var dom domain.NetworkBinding
			err := dto.Pg2Domain(dto.DTO(pgObj, &dom))
			return dom, err
		},
	)
}

// WatchNetworkBindings -
func (rd *pgDbReader) WatchNetworkBindings(ctx context.Context, scope Scope, cb func(domain.NetworkBindingEvent) error) (err error) {
	const (
		api = "WatchNetworkBindings"
		ch  = "resource_NetworkBinding"
	)
	return watch(ctx, api, ch, scope, domain.NetworkBindingResource, rd.doIt,
		func(pgObj pg.ResourceEvent) (domain.NetworkBindingEvent, error) {
			var dom domain.NetworkBindingEvent
			err := dto.Pg2Domain(dto.DTO(pgObj, &dom))
			return dom, err
		},
		cb,
	)
}

// ListServices -
func (rd *pgDbReader) ListServices(ctx context.Context, rsel domain.ResSelectorList) (ret []domain.Service, err error) {
	const (
		api  = "ListServices"
		from = "sgroups.list_services($1)"
	)
	return list(ctx, api, from, scopes.ByResSelectors(rsel...), rd.doIt,
		func(pgObj pg.Service) (domain.Service, error) {
			var dom domain.Service
			err := dto.Pg2Domain(dto.DTO(pgObj, &dom))
			return dom, err
		},
	)
}

// WatchServices -
func (rd *pgDbReader) WatchServices(ctx context.Context, scope Scope, cb func(domain.ServiceEvent) error) (err error) {
	const (
		api = "WatchServices"
		ch  = "resource_Service"
	)
	return watch(ctx, api, ch, scope, domain.ServiceResource, rd.doIt,
		func(pgObj pg.ResourceEvent) (domain.ServiceEvent, error) {
			var dom domain.ServiceEvent
			err := dto.Pg2Domain(dto.DTO(pgObj, &dom))
			return dom, err
		},
		cb,
	)
}

// ListServiceBindings -
func (rd *pgDbReader) ListServiceBindings(ctx context.Context, sel domain.ServiceBindingSelectorList) (ret []domain.ServiceBinding, err error) {
	const (
		api  = "ListServiceBindings"
		from = "sgroups.list_service_bindings()"
	)
	return list(ctx, api, from, scopes.ByServiceBindingSelectors(sel...), rd.doIt,
		func(pgObj pg.ServiceBinding) (domain.ServiceBinding, error) {
			var dom domain.ServiceBinding
			err := dto.Pg2Domain(dto.DTO(pgObj, &dom))
			return dom, err
		},
	)
}

// WatchServiceBindings -
func (rd *pgDbReader) WatchServiceBindings(ctx context.Context, scope Scope, cb func(domain.ServiceBindingEvent) error) (err error) {
	const (
		api = "WatchServiceBindings"
		ch  = "resource_ServiceBinding"
	)
	return watch(ctx, api, ch, scope, domain.ServiceBindingResource, rd.doIt,
		func(pgObj pg.ResourceEvent) (domain.ServiceBindingEvent, error) {
			var dom domain.ServiceBindingEvent
			err := dto.Pg2Domain(dto.DTO(pgObj, &dom))
			return dom, err
		},
		cb,
	)
}

// ListRules -
func (rd *pgDbReader) ListRules(ctx context.Context, sel domain.RulesSelectorList) ([]domain.Rule, error) {
	ctx1, cancel := context.WithCancel(ctx)
	defer cancel()
	rsel := scopes.ByRulesSelectors(sel...)
	listers := [...]func() ([]domain.Rule, error){
		func() ([]domain.Rule, error) {
			return list(ctx1, "ListAg2AgRule", "sgroups.list_ag2ag_rule()", rsel, rd.doIt,
				func(pgObj pg.Res2ResRule) (domain.Rule, error) {
					var dom domain.Rule
					e := dto.Pg2Domain(dto.DTO(pgObj, &dom))
					return dom, e
				},
			)
		},
		func() ([]domain.Rule, error) {
			return list(ctx1, "ListAg2AgIcmpRule", "sgroups.list_ag2ag_icmp_rule()", rsel, rd.doIt,
				func(pgObj pg.Res2ResIcmpRule) (domain.Rule, error) {
					var dom domain.Rule
					e := dto.Pg2Domain(dto.DTO(pgObj, &dom))
					return dom, e
				},
			)
		},
		func() ([]domain.Rule, error) {
			return list(ctx1, "ListAg2IcmpRule", "sgroups.list_ag2icmp_rule()", rsel, rd.doIt,
				func(pgObj pg.Res2IcmpRule) (domain.Rule, error) {
					var dom domain.Rule
					e := dto.Pg2Domain(dto.DTO(pgObj, &dom))
					return dom, e
				},
			)
		},
		func() ([]domain.Rule, error) {
			return list(ctx1, "ListAg2CidrRule", "sgroups.list_ag2cidr_rule()", rsel, rd.doIt,
				func(pgObj pg.Res2CidrRule) (domain.Rule, error) {
					var dom domain.Rule
					e := dto.Pg2Domain(dto.DTO(pgObj, &dom))
					return dom, e
				},
			)
		},
		func() ([]domain.Rule, error) {
			return list(ctx1, "ListAg2CidrIcmpRule", "sgroups.list_ag2cidr_icmp_rule()", rsel, rd.doIt,
				func(pgObj pg.Res2CidrIcmpRule) (domain.Rule, error) {
					var dom domain.Rule
					e := dto.Pg2Domain(dto.DTO(pgObj, &dom))
					return dom, e
				},
			)
		},
		func() ([]domain.Rule, error) {
			return list(ctx1, "ListAg2FqdnRule", "sgroups.list_ag2fqdn_rule()", rsel, rd.doIt,
				func(pgObj pg.Res2FqdnRule) (domain.Rule, error) {
					var dom domain.Rule
					e := dto.Pg2Domain(dto.DTO(pgObj, &dom))
					return dom, e
				},
			)
		},
		func() ([]domain.Rule, error) {
			return list(ctx1, "ListSvc2SvcRule", "sgroups.list_svc2svc_rule()", rsel, rd.doIt,
				func(pgObj pg.Res2ResRule) (domain.Rule, error) {
					var dom domain.Rule
					e := dto.Pg2Domain(dto.DTO(pgObj, &dom))
					return dom, e
				},
			)
		},
		func() ([]domain.Rule, error) {
			return list(ctx1, "ListSvc2FqdnRule", "sgroups.list_svc2fqdn_rule()", rsel, rd.doIt,
				func(pgObj pg.Res2FqdnRule) (domain.Rule, error) {
					var dom domain.Rule
					e := dto.Pg2Domain(dto.DTO(pgObj, &dom))
					return dom, e
				},
			)
		},
		func() ([]domain.Rule, error) {
			return list(ctx1, "ListSvc2CidrRule", "sgroups.list_svc2cidr_rule()", rsel, rd.doIt,
				func(pgObj pg.Res2CidrRule) (domain.Rule, error) {
					var dom domain.Rule
					e := dto.Pg2Domain(dto.DTO(pgObj, &dom))
					return dom, e
				},
			)
		},
		func() ([]domain.Rule, error) {
			return list(ctx1, "ListSvc2CidrIcmpRule", "sgroups.list_svc2cidr_icmp_rule()", rsel, rd.doIt,
				func(pgObj pg.Res2CidrIcmpRule) (domain.Rule, error) {
					var dom domain.Rule
					e := dto.Pg2Domain(dto.DTO(pgObj, &dom))
					return dom, e
				},
			)
		},
		func() ([]domain.Rule, error) {
			return list(ctx1, "ListAg2SvcRule", "sgroups.list_ag2svc_rule()", rsel, rd.doIt,
				func(pgObj pg.Res2ResRule) (domain.Rule, error) {
					var dom domain.Rule
					e := dto.Pg2Domain(dto.DTO(pgObj, &dom))
					return dom, e
				},
			)
		},
		func() ([]domain.Rule, error) {
			return list(ctx1, "ListSvc2AgRule", "sgroups.list_svc2ag_rule()", rsel, rd.doIt,
				func(pgObj pg.Res2ResRule) (domain.Rule, error) {
					var dom domain.Rule
					e := dto.Pg2Domain(dto.DTO(pgObj, &dom))
					return dom, e
				},
			)
		},
		func() ([]domain.Rule, error) {
			return list(ctx1, "ListSvc2AgIcmpRule", "sgroups.list_svc2ag_icmp_rule()", rsel, rd.doIt,
				func(pgObj pg.Res2ResIcmpRule) (domain.Rule, error) {
					var dom domain.Rule
					e := dto.Pg2Domain(dto.DTO(pgObj, &dom))
					return dom, e
				},
			)
		},
		func() ([]domain.Rule, error) {
			return list(ctx1, "ListAg2SvcIcmpRule", "sgroups.list_ag2svc_icmp_rule()", rsel, rd.doIt,
				func(pgObj pg.Res2ResIcmpRule) (domain.Rule, error) {
					var dom domain.Rule
					e := dto.Pg2Domain(dto.DTO(pgObj, &dom))
					return dom, e
				},
			)
		},
	}
	r := make([][]domain.Rule, len(listers))
	errs := make([]error, len(listers))
	_ = parallel.ExecAbstract(len(listers), int32(len(listers))-1, func(i int) error {
		r[i], errs[i] = listers[i]()
		return nil
	})
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return slices.Concat(r...), multierr.Combine(errs...)
}

// WatchRules -
func (rd *pgDbReader) WatchRules(ctx context.Context, scope Scope, cb func(domain.RuleEvent) error) error {
	ctx1, cancel := context.WithCancel(ctx)
	defer cancel()

	conv := func(pgObj pg.ResourceEvent) (domain.RuleEvent, error) {
		var dom domain.RuleEvent
		e := dto.Pg2Domain(dto.DTO(pgObj, &dom))
		return dom, e
	}

	watchers := [...]func() error{
		func() error {
			return watch(ctx1, "WatchAg2AgRule", "resource_Ag2AgRule", scope, domain.Ag2AgRule, rd.doIt, conv, cb)
		},
		func() error {
			return watch(ctx1, "WatchAg2AgIcmpRule", "resource_Ag2AgIcmpRule", scope, domain.Ag2AgIcmpRule, rd.doIt, conv, cb)
		},
		func() error {
			return watch(ctx1, "WatchAg2IcmpRule", "resource_Ag2IcmpRule", scope, domain.Ag2IcmpRule, rd.doIt, conv, cb)
		},
		func() error {
			return watch(ctx1, "WatchAg2CidrRule", "resource_Ag2CidrRule", scope, domain.Ag2CidrRule, rd.doIt, conv, cb)
		},
		func() error {
			return watch(ctx1, "WatchAg2CidrIcmpRule", "resource_Ag2CidrIcmpRule", scope, domain.Ag2CidrIcmpRule, rd.doIt, conv, cb)
		},
		func() error {
			return watch(ctx1, "WatchAg2FqdnRule", "resource_Ag2FqdnRule", scope, domain.Ag2FqdnRule, rd.doIt, conv, cb)
		},
		func() error {
			return watch(ctx1, "WatchSvc2SvcRule", "resource_Svc2SvcRule", scope, domain.Svc2SvcRule, rd.doIt, conv, cb)
		},
		func() error {
			return watch(ctx1, "WatchSvc2FqdnRule", "resource_Svc2FqdnRule", scope, domain.Svc2FqdnRule, rd.doIt, conv, cb)
		},
		func() error {
			return watch(ctx1, "WatchSvc2CidrRule", "resource_Svc2CidrRule", scope, domain.Svc2CidrRule, rd.doIt, conv, cb)
		},
		func() error {
			return watch(ctx1, "WatchSvc2CidrIcmpRule", "resource_Svc2CidrIcmpRule", scope, domain.Svc2CidrIcmpRule, rd.doIt, conv, cb)
		},
		func() error {
			return watch(ctx1, "WatchAg2SvcRule", "resource_Ag2SvcRule", scope, domain.Ag2SvcRule, rd.doIt, conv, cb)
		},
		func() error {
			return watch(ctx1, "WatchSvc2AgRule", "resource_Svc2AgRule", scope, domain.Svc2AgRule, rd.doIt, conv, cb)
		},
		func() error {
			return watch(ctx1, "WatchSvc2AgIcmpRule", "resource_Svc2AgIcmpRule", scope, domain.Svc2AgIcmpRule, rd.doIt, conv, cb)
		},
		func() error {
			return watch(ctx1, "WatchAg2SvcIcmpRule", "resource_Ag2SvcIcmpRule", scope, domain.Ag2SvcIcmpRule, rd.doIt, conv, cb)
		},
	}

	errs := make([]error, len(watchers))
	_ = parallel.ExecAbstract(len(watchers), int32(len(watchers))-1, func(i int) error {
		errs[i] = watchers[i]()
		if errs[i] != nil {
			cancel()
		}
		return nil
	})

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return multierr.Combine(errs...)
}

func list[domainT any, pgT interface{ Columns() []string }](
	ctx context.Context,
	api string,
	from string,
	scope Scope,
	doIt func(context.Context, func(pgConn) error) error,
	conv func(pgT) (domainT, error),
) (ret []domainT, err error) {
	defer func() {
		err = errors.WithMessage(err, api)
	}()

	var (
		args   []any
		flt    Scope
		qry    string
		pgZero pgT
	)

	qry, _, err = sq.Select(pgZero.Columns()...).
		From(from).
		ToSql()
	if err != nil {
		return nil, err
	}

	switch sc := scope.(type) {
	case scopes.ScopeByResSelectors:
		var pgRsel pg.ResSelectorList
		if err = dto.Domain2Pg(dto.DTO(sc.Selectors, &pgRsel)); err != nil {
			return nil, err
		}
		args = []any{pgx.QueryExecModeDescribeExec, pgRsel}
		flt = scopes.NoScope
	case scopes.ScopeByHostBindingSelectors:
		flt = sc
	case scopes.ScopeByNetworkBindingSelectors:
		flt = sc
	case scopes.ScopeByServiceBindingSelectors:
		flt = sc
	case scopes.ScopeByRulesSelectors:
		flt = sc
	}

	var lw lister.ListerWatcher[domainT, pgT]
	if err = lw.Init(flt, qry, args...); err != nil {
		return nil, err
	}
	err = doIt(ctx, func(c pgConn) error {
		var e error
		ret, e = lw.List(ctx, c.Query, conv)
		return e
	})

	return ret, err
}

func watch[domainT any, pgT any](
	ctx context.Context,
	api string,
	channel string,
	scope Scope,
	res domain.ResourceType,
	doIt func(context.Context, func(pgConn) error) error,
	conv func(pgT) (domainT, error),
	cb func(domainT) (err error)) (err error) {
	const (
		sql = "select * from sgroups.list_outbox_resource_events($1, $2, $3)"
	)
	var (
		argResType  = res.String()
		argMaxEvent = 1000000
		flt         Scope
	)

	defer func() {
		err = errors.WithMessage(err, api)
	}()

	args := []any{pgx.QueryExecModeDescribeExec, argResType, nil, argMaxEvent}

	switch sc := scope.(type) {
	case scopes.ScopedAnd:
		for _, x := range []any{sc.L, sc.R} {
			switch a := x.(type) {
			case scopes.ScopeByResourceVersion:
				rv, e := strconv.Atoi(a.RV)
				if e != nil {
					return errors.WithMessagef(e, "invalid resource version '%s'", a.RV)
				}
				args[2] = rv
			case scopes.ScopeByResSelectors:
				flt = a
			case scopes.ScopeByHostBindingSelectors:
				flt = a
			case scopes.ScopeByNetworkBindingSelectors:
				flt = a
			case scopes.ScopeByServiceBindingSelectors:
				flt = a
			case scopes.ScopeByRulesSelectors:
				flt = a
			}
		}
	}

	var lw lister.ListerWatcher[domainT, pgT]
	if err = lw.Init(flt, sql, args...); err != nil {
		return err
	}
	var evts []domainT
	err = doIt(ctx, func(c pgConn) error {
		if evts, err = lw.List(ctx, c.Query, conv); err != nil {
			return err
		}
		for _, evt := range evts {
			if err = cb(evt); err != nil {
				return err
			}
		}

		return lw.Watch(ctx, c, channel, conv, cb)
	})
	return err
}
