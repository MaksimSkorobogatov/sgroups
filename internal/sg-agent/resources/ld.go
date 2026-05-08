package resources

import (
	"context"
	"slices"
	"time"

	sg "github.com/PRO-Robotech/sgroups/internal/sg-agent/transport/sgroups/grpc"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	"github.com/H-BF/corlib/logger"
	"github.com/H-BF/corlib/pkg/host"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type (
	// LocalData are used by agent to build Host Based Firewall rules
	LocalData struct {
		LocalHost LocalHost
		Hosts     Hosts
		LocalAGs  AGs
		Networks  Networks
		Services  Services
		Rules     Rules

		DefPolicyAccept bool
		ResolvedFQDN    *ResolvedFQDN
		SyncStatus      domain.SyncStatus
	}

	// LocalDataLoader -
	LocalDataLoader struct {
		SyncStatus      domain.SyncStatus
		MaxLoadDuration time.Duration
		DefPolicyAccept bool
	}
)

func (ld *LocalData) allUsedAGs() []AgID {
	return lo.Uniq(slices.Concat(ld.Rules.GetAgRefs(), ld.Services.GetAgRefs()))
}

func (ld *LocalData) nonLocalAGs() []AgID {
	return lo.Filter(ld.allUsedAGs(), func(id AgID, _ int) bool {
		_, ok := ld.LocalAGs.Get(id)
		return !ok
	})
}

// IsEq checks wether this object is equal the other one
// here we compare only rules and networks
func (ld *LocalData) IsEq(other LocalData) bool {
	eq := ld.LocalHost.IsEq(other.LocalHost)
	if eq {
		eq = ld.Hosts.IsEq(other.Hosts)
	}
	if eq {
		eq = ld.LocalAGs.IsEq(other.LocalAGs)
	}
	if eq {
		eq = ld.Networks.IsEq(other.Networks)
	}
	if eq {
		eq = ld.Services.IsEq(other.Services)
	}
	if eq {
		eq = ld.Rules.IsEq(other.Rules)
	}
	return eq
}

// Load -
func (loader *LocalDataLoader) Load(ctx context.Context, client sg.Clients, ncnf host.NetConf) (res LocalData, err error) {
	defer func() {
		err = errors.WithMessage(err, "LocalData/Load")
	}()

	res.SyncStatus = loader.SyncStatus
	res.DefPolicyAccept = loader.DefPolicyAccept

	log := logger.FromContext(ctx)
	if loader.MaxLoadDuration > 0 {
		ctx1, cancel := context.WithTimeout(ctx, loader.MaxLoadDuration)
		defer cancel()
		ctx = ctx1
	}

	log.Debug("loading local host...")
	if err = res.LocalHost.Load(ctx, client); err != nil {
		return res, err
	}
	ags := res.LocalHost.GetAddressGroupRefs()
	if len(ags) == 0 {
		log.Debug("local host has no address group refs")
		return res, nil
	}

	log.Debugf("loading local address groups from refs: (%v)...", ags)
	if err = res.LocalAGs.Load(ctx, client, ags); err != nil {
		return res, err
	}

	if res.LocalAGs.Len() == 0 {
		log.Warn("no any local AG is found")
		return res, err
	} else {
		log.Debugf("found local AG(s) %v", res.LocalAGs.IDs())
	}

	log.Debugw("loading hosts from local AG(s)...")
	if err = res.Hosts.Load(ctx, client, res.LocalAGs.GetHostRefs()); err != nil {
		return res, err
	}

	log.Debugw("loading networks from local AG(s)...")
	if err = res.Networks.Load(ctx, client, res.LocalAGs.GetNetworkRefs()); err != nil {
		return res, err
	}
	log.Debugw("loading services from local AG(s)...")
	if err = res.Services.Load(ctx, client, res.LocalAGs.GetServiceRefs()); err != nil {
		return res, err
	}
	log.Debugw("loading rules from local AG(s)...")
	if err = res.Rules.LoadFromRefs(ctx, client, res.LocalAGs.GetRuleRefs()); err != nil {
		return res, err
	}
	log.Debugw("loading rules from local Service(s)...")
	if err = res.Rules.LoadFromRefs(ctx, client, res.Services.GetRuleRefs()); err != nil {
		return res, err
	}
	nonLocalSvcRefs := lo.Uniq(lo.Filter(res.Rules.GetSvcRefs(), func(svcRef domain.ResourceIdentifier, _ int) bool {
		_, ok := res.Services.Get(svcRef)
		return !ok
	}))

	if len(nonLocalSvcRefs) > 0 {
		log.Debugw("loading non local services from rules...")
		if err = res.Services.Load(ctx, client, nonLocalSvcRefs); err != nil {
			return res, err
		}
	}
	if nonLocalAGs := res.nonLocalAGs(); len(nonLocalAGs) > 0 {
		log.Debugf("loading networks from non local AG(s) %s...", nonLocalAGs)
		if err = res.Networks.LoadFromAGs(ctx, client, nonLocalAGs); err != nil {
			return res, err
		}
		log.Debugf("loading services from non local AG(s) %s...", nonLocalAGs)
		if err = res.Services.LoadFromAGs(ctx, client, nonLocalAGs); err != nil {
			return res, err
		}
		log.Debugf("loading hosts from non local AG(s) %s...", nonLocalAGs)
		if err = res.Hosts.LoadFromAGs(ctx, client, nonLocalAGs); err != nil {
			return res, err
		}
	}

	return res, err
}
