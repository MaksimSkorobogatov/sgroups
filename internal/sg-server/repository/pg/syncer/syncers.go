package syncer

import (
	pg "github.com/PRO-Robotech/sgroups/internal/sg-server/repository/pg/domain"
	"github.com/PRO-Robotech/sgroups/internal/sg-server/repository/pg/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
)

// NamespaceSyncer -
var NamespaceSyncer = syncerObj[domain.Namespace, pg.Namespace]{
	syncArgs: func(ns pg.Namespace) []any {
		return []any{
			ns.UID,
			ns.Name,
			ns.Labels,
			ns.Annotations,
			ns.Comment,
			ns.Description,
			ns.DisplayName,
		}
	},
	sqlBuilder: rowBuilder{
		argCount:    7,
		sqlFuncName: "sgroups.sync_namespaces",
	},
	toPgConv: func(src domain.Namespace) (dst pg.Namespace, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	fromPgConv: func(src pg.Namespace) (dst domain.Namespace, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
}

// AddressGroupSyncer -
var AddressGroupSyncer = syncerObj[domain.AddressGroup, pg.AddressGroup]{
	syncArgs: func(ag pg.AddressGroup) []any {
		return []any{
			ag.UID,
			ag.Name,
			ag.Namespace,
			ag.Labels,
			ag.Annotations,
			ag.Comment,
			ag.Description,
			ag.DisplayName,
			ag.DefaultAction,
			ag.Logs,
			ag.Trace,
		}
	},
	delArgs: func(ag pg.AddressGroup) []any {
		return []any{
			ag.UID,
			ag.Name,
			ag.Namespace,
			nil, nil, nil, nil, nil, nil, nil, nil,
		}
	},
	sqlBuilder: rowBuilder{
		argCount:    11,
		sqlFuncName: "sgroups.sync_address_groups",
	},
	toPgConv: func(src domain.AddressGroup) (dst pg.AddressGroup, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	fromPgConv: func(src pg.AddressGroup) (dst domain.AddressGroup, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
}

// NetworkSyncer -
var NetworkSyncer = syncerObj[domain.Network, pg.Network]{
	syncArgs: func(nw pg.Network) []any {
		return []any{
			nw.UID,
			nw.Name,
			nw.Namespace,
			nw.Labels,
			nw.Annotations,
			nw.Comment,
			nw.Description,
			nw.DisplayName,
			nw.Network.IPNet,
		}
	},
	delArgs: func(nw pg.Network) []any {
		return []any{
			nw.UID,
			nw.Name,
			nw.Namespace,
			nil, nil, nil, nil, nil, nil,
		}
	},
	sqlBuilder: rowBuilder{
		argCount:    9,
		sqlFuncName: "sgroups.sync_networks",
	},
	toPgConv: func(src domain.Network) (dst pg.Network, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	fromPgConv: func(src pg.Network) (dst domain.Network, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
}

// HostIPsSyncer -
var HostIPsSyncer = makeHostSyncer(
	"sgroups.sync_host_ipset",
	func(h pg.Host) []any {
		return []any{
			h.UID,
			h.Name,
			h.Namespace,
			h.IPs,
			h.Endpoints,
		}
	},
)

// HostInfoSyncer -
var HostInfoSyncer = makeHostSyncer(
	"sgroups.sync_host_info",
	func(h pg.Host) []any {
		return []any{
			h.UID,
			h.Name,
			h.Namespace,
			h.MetaInfo,
		}
	},
)

// HostHealthSyncer -
var HostHealthSyncer = makeHostSyncer(
	"sgroups.sync_host_health_status",
	func(h pg.Host) []any {
		return []any{
			h.UID,
			h.Name,
			h.Namespace,
			h.HealthStatus,
		}
	},
)

// HostSyncer -
var HostSyncer = makeHostSyncer(
	"sgroups.sync_hosts",
	func(h pg.Host) []any {
		return []any{
			h.UID,
			h.Name,
			h.Namespace,
			h.Labels,
			h.Annotations,
			h.Comment,
			h.Description,
			h.DisplayName,
		}
	},
)

// HostBindingSyncer -
var HostBindingSyncer = syncerObj[domain.HostBinding, pg.HostBinding]{ //nolint:dupl
	syncArgs: func(hb pg.HostBinding) []any {
		return []any{
			hb.UID,
			hb.Name,
			hb.Namespace,
			hb.Labels,
			hb.Annotations,
			hb.Comment,
			hb.Description,
			hb.DisplayName,
			hb.AddressGroup,
			hb.Host,
		}
	},
	sqlBuilder: rowBuilder{
		argCount:    10,
		sqlFuncName: "sgroups.sync_host_bindings",
	},
	toPgConv: func(src domain.HostBinding) (dst pg.HostBinding, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	fromPgConv: func(src pg.HostBinding) (dst domain.HostBinding, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
}

// NetworkBindingSyncer -
var NetworkBindingSyncer = syncerObj[domain.NetworkBinding, pg.NetworkBinding]{ //nolint:dupl
	syncArgs: func(nb pg.NetworkBinding) []any {
		return []any{
			nb.UID,
			nb.Name,
			nb.Namespace,
			nb.Labels,
			nb.Annotations,
			nb.Comment,
			nb.Description,
			nb.DisplayName,
			nb.AddressGroup,
			nb.Network,
		}
	},
	delArgs: func(nb pg.NetworkBinding) []any {
		return []any{
			nb.UID,
			nb.Name,
			nb.Namespace,
			nil, nil, nil, nil, nil, nil, nil,
		}
	},
	sqlBuilder: rowBuilder{
		argCount:    10,
		sqlFuncName: "sgroups.sync_network_bindings",
	},
	toPgConv: func(src domain.NetworkBinding) (dst pg.NetworkBinding, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	fromPgConv: func(src pg.NetworkBinding) (dst domain.NetworkBinding, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
}

// RuleSyncer - delete-only syncer.
var RuleSyncer = syncerObj[domain.Rule, pg.Rule]{
	delArgs: func(r pg.Rule) []any {
		return []any{r.UID, r.Name, r.Namespace}
	},
	sqlBuilder: rowBuilder{
		argCount:    3,
		sqlFuncName: "sgroups.sync_rule_via_registry",
	},
	toPgConv: func(src domain.Rule) (dst pg.Rule, err error) {
		dst.UID = src.Metadata.ID.UID
		dst.Name = string(src.Metadata.ID.Name)
		dst.Namespace = string(src.Metadata.ID.Namespace)
		return dst, nil
	},
	fromPgConv: func(src pg.Rule) (dst domain.Rule, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
	disabledOps: []SyncOp{Upsert},
}

// Ag2AgSyncer -
var Ag2AgSyncer = makeSyncer(
	"sgroups.sync_ag2ag_rule",
	func(r pg.Res2ResRule) []any {
		return []any{
			r.UID,
			r.Name,
			r.Namespace,
		}
	},
	func(src domain.Rule) (dst pg.Res2ResRule, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	func(src pg.Res2ResRule) (dst domain.Rule, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
)

// Ag2AgIcmpSyncer -
var Ag2AgIcmpSyncer = makeSyncer(
	"sgroups.sync_ag2ag_icmp_rule",
	func(r pg.Res2ResIcmpRule) []any {
		return []any{
			r.UID,
			r.Name,
			r.Namespace,
		}
	},
	func(src domain.Rule) (dst pg.Res2ResIcmpRule, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	func(src pg.Res2ResIcmpRule) (dst domain.Rule, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
)

// Ag2IcmpSyncer -
var Ag2IcmpSyncer = makeSyncer(
	"sgroups.sync_ag2icmp_rule",
	func(r pg.Res2IcmpRule) []any {
		return []any{
			r.UID,
			r.Name,
			r.Namespace,
		}
	},
	func(src domain.Rule) (dst pg.Res2IcmpRule, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	func(src pg.Res2IcmpRule) (dst domain.Rule, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
)

// Ag2CidrSyncer -
var Ag2CidrSyncer = makeSyncer(
	"sgroups.sync_ag2cidr_rule",
	func(r pg.Res2CidrRule) []any {
		return []any{
			r.UID,
			r.Name,
			r.Namespace,
		}
	},
	func(src domain.Rule) (dst pg.Res2CidrRule, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	func(src pg.Res2CidrRule) (dst domain.Rule, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
)

// Ag2CidrIcmpSyncer -
var Ag2CidrIcmpSyncer = makeSyncer(
	"sgroups.sync_ag2cidr_icmp_rule",
	func(r pg.Res2CidrIcmpRule) []any {
		return []any{
			r.UID,
			r.Name,
			r.Namespace,
		}
	},
	func(src domain.Rule) (dst pg.Res2CidrIcmpRule, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	func(src pg.Res2CidrIcmpRule) (dst domain.Rule, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
)

// Ag2FqdnSyncer -
var Ag2FqdnSyncer = makeSyncer(
	"sgroups.sync_ag2fqdn_rule",
	func(r pg.Res2FqdnRule) []any {
		return []any{
			r.UID,
			r.Name,
			r.Namespace,
		}
	},
	func(src domain.Rule) (dst pg.Res2FqdnRule, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	func(src pg.Res2FqdnRule) (dst domain.Rule, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
)

// Svc2SvcSyncer -
var Svc2SvcSyncer = makeSyncer(
	"sgroups.sync_svc2svc_rule",
	func(r pg.Res2ResRule) []any {
		return []any{
			r.UID,
			r.Name,
			r.Namespace,
		}
	},
	func(src domain.Rule) (dst pg.Res2ResRule, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	func(src pg.Res2ResRule) (dst domain.Rule, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
)

// Svc2FqdnSyncer -
var Svc2FqdnSyncer = makeSyncer(
	"sgroups.sync_svc2fqdn_rule",
	func(r pg.Res2FqdnRule) []any {
		return []any{
			r.UID,
			r.Name,
			r.Namespace,
		}
	},
	func(src domain.Rule) (dst pg.Res2FqdnRule, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	func(src pg.Res2FqdnRule) (dst domain.Rule, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
)

// Svc2CidrSyncer -
var Svc2CidrSyncer = makeSyncer(
	"sgroups.sync_svc2cidr_rule",
	func(r pg.Res2CidrRule) []any {
		return []any{
			r.UID,
			r.Name,
			r.Namespace,
		}
	},
	func(src domain.Rule) (dst pg.Res2CidrRule, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	func(src pg.Res2CidrRule) (dst domain.Rule, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
)

// Svc2CidrIcmpSyncer -
var Svc2CidrIcmpSyncer = makeSyncer(
	"sgroups.sync_svc2cidr_icmp_rule",
	func(r pg.Res2CidrIcmpRule) []any {
		return []any{
			r.UID,
			r.Name,
			r.Namespace,
		}
	},
	func(src domain.Rule) (dst pg.Res2CidrIcmpRule, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	func(src pg.Res2CidrIcmpRule) (dst domain.Rule, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
)

// Ag2SvcSyncer -
var Ag2SvcSyncer = makeSyncer(
	"sgroups.sync_ag2svc_rule",
	func(r pg.Res2ResRule) []any {
		return []any{
			r.UID,
			r.Name,
			r.Namespace,
		}
	},
	func(src domain.Rule) (dst pg.Res2ResRule, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	func(src pg.Res2ResRule) (dst domain.Rule, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
)

// Svc2AgSyncer -
var Svc2AgSyncer = makeSyncer(
	"sgroups.sync_svc2ag_rule",
	func(r pg.Res2ResRule) []any {
		return []any{
			r.UID,
			r.Name,
			r.Namespace,
		}
	},
	func(src domain.Rule) (dst pg.Res2ResRule, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	func(src pg.Res2ResRule) (dst domain.Rule, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
)

// Svc2AgIcmpSyncer -
var Svc2AgIcmpSyncer = makeSyncer(
	"sgroups.sync_svc2ag_icmp_rule",
	func(r pg.Res2ResIcmpRule) []any {
		return []any{
			r.UID,
			r.Name,
			r.Namespace,
		}
	},
	func(src domain.Rule) (dst pg.Res2ResIcmpRule, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	func(src pg.Res2ResIcmpRule) (dst domain.Rule, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
)

// Ag2SvcIcmpSyncer -
var Ag2SvcIcmpSyncer = makeSyncer(
	"sgroups.sync_ag2svc_icmp_rule",
	func(r pg.Res2ResIcmpRule) []any {
		return []any{
			r.UID,
			r.Name,
			r.Namespace,
		}
	},
	func(src domain.Rule) (dst pg.Res2ResIcmpRule, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	func(src pg.Res2ResIcmpRule) (dst domain.Rule, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
)

// ServiceSyncer -
var ServiceSyncer = makeSyncer(
	"sgroups.sync_services",
	func(s pg.Service) []any {
		return []any{
			s.UID,
			s.Name,
			s.Namespace,
		}
	},
	func(src domain.Service) (dst pg.Service, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	func(src pg.Service) (dst domain.Service, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
)

// ServiceBindingSyncer -
var ServiceBindingSyncer = makeSyncer(
	"sgroups.sync_service_bindings",
	func(sb pg.ServiceBinding) []any {
		return []any{
			sb.UID,
			sb.Name,
			sb.Namespace,
		}
	},
	func(src domain.ServiceBinding) (dst pg.ServiceBinding, err error) {
		err = dto.Domain2Pg(dto.DTO(src, &dst))
		return dst, err
	},
	func(src pg.ServiceBinding) (dst domain.ServiceBinding, err error) {
		err = dto.Pg2Domain(dto.DTO(src, &dst))
		return dst, err
	},
)

//nolint:dupl
func makeHostSyncer(mutator string, syncArgs func(h pg.Host) []any) syncerObj[domain.Host, pg.Host] {
	return syncerObj[domain.Host, pg.Host]{
		syncArgs: syncArgs,
		sqlBuilder: rowBuilder{
			argCount:    len(syncArgs(pg.Host{})),
			sqlFuncName: mutator,
		},
		toPgConv: func(src domain.Host) (dst pg.Host, err error) {
			err = dto.Domain2Pg(dto.DTO(src, &dst))
			return dst, err
		},
		fromPgConv: func(src pg.Host) (dst domain.Host, err error) {
			err = dto.Pg2Domain(dto.DTO(src, &dst))
			return dst, err
		},
	}
}

func makeSyncer[domainT any, pgT interface{ SyncArgs() []any }](
	mutator string,
	delArgs func(pgT) []any,
	toPgConv func(domainT) (pgT, error),
	fromPgConv func(pgT) (domainT, error),
) syncerObj[domainT, pgT] {
	var (
		pgZero pgT
		del    func(pgT) []any
	)
	argCnt := len(pgZero.SyncArgs())
	if delArgs != nil {
		del = func(pt pgT) []any {
			da := delArgs(pt)
			if len(da) < argCnt {
				da = append(da, make([]any, argCnt-len(da))...)
			}
			return da
		}
	}
	return syncerObj[domainT, pgT]{
		syncArgs: func(pt pgT) []any {
			return pt.SyncArgs()
		},
		delArgs: del,
		sqlBuilder: rowBuilder{
			argCount:    argCnt,
			sqlFuncName: mutator,
		},
		toPgConv:   toPgConv,
		fromPgConv: fromPgConv,
	}
}
