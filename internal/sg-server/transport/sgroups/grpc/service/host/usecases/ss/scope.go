package ss

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/PRO-Robotech/sgroups/internal/sg-server/transport/sgroups/grpc/service/host/dto"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"

	agentv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/agent/v1"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
)

type (
	hostScope struct {
		host    domain.ResourceIdentifier
		addr    string
		filters []*agentv1.SockStat_Selectors
	}

	hostLister func(ctx context.Context, sel domain.ResSelectorList) ([]domain.Host, error)
)

// String - returns string representation of host scopes
func (h hostScope) String() string {
	return fmt.Sprintf("host %s at %s", h.host, h.addr)
}

func resolveHostScopes[T interface {
	*sgv1.HostReq_SocketStatistics_List |
		*sgv1.HostReq_SocketStatistics_Watch
	GetSelectors() []*sgv1.HostReq_SocketStatistics_FieldSelector
}](
	ctx context.Context,
	req T,
	listHosts hostLister,
) (scopes []hostScope, err error) {
	var (
		sel   domain.ResSelectorList
		hosts []domain.Host
	)
	switch r := any(req).(type) {
	case *sgv1.HostReq_SocketStatistics_List:
		err = dto.Proto2Domain(dto.DTO(r, &sel))
	case *sgv1.HostReq_SocketStatistics_Watch:
		err = dto.Proto2Domain(dto.DTO(r, &sel))
	}
	if err != nil {
		return nil, err
	}
	if hosts, err = listHosts(ctx, sel); err != nil {
		return nil, err
	}

	for _, host := range hosts {
		ep := host.Spec.Endpoints
		if ep == nil {
			continue
		}
		if !ep.Address.IsValid() {
			return nil, errors.New("invalid host endpoint address")
		}

		filters := filtersForHost(host.ResourceID(), req.GetSelectors())

		addr := ep.Address.String()
		for _, p := range ep.Ports {
			if p.Name != domain.AgentApiEndpointName {
				continue
			}
			scopes = append(scopes, hostScope{
				host:    host.ResourceID(),
				addr:    net.JoinHostPort(addr, strconv.Itoa(int(p.Port))),
				filters: filters,
			})
		}
	}

	return scopes, nil
}

func filtersForHost(
	hostID domain.ResourceIdentifier,
	selectors []*sgv1.HostReq_SocketStatistics_FieldSelector,
) []*agentv1.SockStat_Selectors {
	var filters []*agentv1.SockStat_Selectors
	for _, sel := range selectors {
		if name := sel.GetName(); len(name) > 0 && domain.ResourceName(name) != hostID.Name {
			continue
		}
		if ns := sel.GetNamespace(); len(ns) > 0 && domain.ResourceNamespace(ns) != hostID.Namespace {
			continue
		}
		filters = append(filters, sel.GetFilters()...)
	}
	return filters
}
