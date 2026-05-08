package resources

import (
	"context"
	"net/netip"

	conf "github.com/PRO-Robotech/sgroups/internal/sg-agent/config"
	dto "github.com/PRO-Robotech/sgroups/internal/sg-agent/dto/host"
	sg "github.com/PRO-Robotech/sgroups/internal/sg-agent/transport/sgroups/grpc"
	domain "github.com/PRO-Robotech/sgroups/internal/shared/domains/sgroups"
	"github.com/PRO-Robotech/sgroups/internal/shared/misc"

	"github.com/H-BF/corlib/pkg/dict"
	"github.com/H-BF/corlib/pkg/host"
	"github.com/H-BF/corlib/pkg/parallel"
	"github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	hinfo "github.com/shirou/gopsutil/v4/host"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// LocalHost - local host information with metadata
	LocalHost struct {
		domain.Host
	}

	// HostID -
	HostID = domain.ResourceIdentifier

	// Hosts - list of hosts with metadata
	Hosts struct {
		dict.HDict[HostID, domain.Host]
	}

	hostInfo struct {
		domain.HostInfo
	}

	hostIPs struct {
		domain.DualStackIPs
	}
)

// Load loads hosts from SGroups server
func (h *LocalHost) Load(ctx context.Context, sgClients sg.Clients) error {
	client, err := sgClients.Hosts()
	if err != nil {
		return err
	}
	var host domain.ResourceIdentifier
	if host, err = conf.GetHostID(ctx); err != nil {
		return err
	}
	resp, e := client.List(ctx, &sgv1.HostReq_List{
		Selectors: makeFieldSelectorsByNames(misc.Sli(host)),
	})
	if e != nil {
		return errors.WithMessage(e, "list hosts")
	}
	for _, pbHost := range resp.GetHosts() {
		err = dto.Proto2Domain(dto.DTO(pbHost, &h.Host))
		break
	}

	return err
}

// Syncable checks if local host has enough information to be synchronized with SGroups server
func (h *LocalHost) Syncable() bool {
	return h.Metadata.ID.UID != uuid.Nil
}

// Sync synchronizes local host information with SGroups server
func (h *LocalHost) Sync(ctx context.Context, sgClients sg.Clients, ncnf host.NetConf) error {
	reqs := []func() error{
		func() error {
			var hi hostInfo
			return hi.sync(ctx, sgClients, h.Metadata.ID.UID)
		},
		func() error {
			var hIPs hostIPs
			return hIPs.sync(ctx, sgClients, h.Metadata.ID.UID, ncnf)
		},
	}
	errs := make([]error, len(reqs))
	_ = parallel.ExecAbstract(len(reqs), 0, func(i int) error {
		errs[i] = reqs[i]()
		return nil
	})
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return multierr.Combine(errs...)
}

// IsEq -
func (h *LocalHost) IsEq(other LocalHost) bool {
	return h.Host.IsEq(other.Host)
}

// IsEq -
func (h *Hosts) IsEq(other Hosts) bool {
	return h.Eq(&other.HDict, func(lAg, rAg domain.Host) bool {
		return lAg.IsEq(rAg)
	})
}

// Load -
func (h *Hosts) Load(ctx context.Context, sgClients sg.Clients, hosts []HostID) error {
	return h.load(ctx, sgClients, makeFieldSelectorsByNames(hosts))
}

// LoadFromAGs loads hosts from SGroups server by address groups
func (h *Hosts) LoadFromAGs(ctx context.Context, sgClients sg.Clients, ags []AgID) error {
	return h.load(ctx, sgClients, makeFieldSelectorsByRefs(ags, domain.AddressGroupResource))
}

func (h *Hosts) load(ctx context.Context, sgClients sg.Clients, selectors []*common.ResSelector) error {
	client, err := sgClients.Hosts()
	if err != nil {
		return err
	}
	return loader(ctx, selectors,
		func(ctx context.Context, selectors []*common.ResSelector) ([]*sgv1.HostResp_HostExt, error) {
			resp, e := client.List(ctx, &sgv1.HostReq_List{
				Selectors: selectors,
			})
			return resp.GetHosts(), errors.WithMessage(e, "list hosts")
		},
		func(pb *sgv1.HostResp_HostExt) (domain.Host, error) {
			var host domain.Host
			err := dto.Proto2Domain(dto.DTO(pb, &host))
			return host, err
		},
		h.Put,
	)
}

func (h *hostInfo) sync(ctx context.Context, sgClients sg.Clients, uid domain.UUID) error {
	client, err := sgClients.Hosts()
	if err != nil {
		return err
	}
	var host domain.ResourceIdentifier
	if host, err = conf.GetHostID(ctx); err != nil {
		return err
	}

	if err = h.load(ctx); err != nil {
		return err
	}

	_, err = client.UpdMetaInfo(ctx, &sgv1.HostReq_UpdMetaInfo{
		Hosts: []*sgv1.HostReq_UpdMetaInfo_HostInfo{
			{
				Metadata: &common.MetadataScope{
					Uid:       uid.String(),
					Name:      host.Name.String(),
					Namespace: host.Namespace.String(),
				},
				Spec: &sgv1.HostReq_UpdMetaInfo_HostInfo_Spec{
					MetaInfo: &sgv1.Host_Spec_MetaInfo{
						HostName:        h.HostName,
						Os:              h.OS,
						Platform:        h.Platform,
						PlatformFamily:  h.PlatformFamily,
						PlatformVersion: h.PlatformVersion,
						KernelVersion:   h.KernelVersion,
					},
				},
			},
		},
	})
	if misc.IsIn(status.Code(err), codes.NotFound, codes.InvalidArgument) {
		err = ErrHostNotFound
	}
	return err
}

func (h *hostInfo) load(ctx context.Context) error {
	info, err := hinfo.InfoWithContext(ctx)
	if err != nil {
		return errors.WithMessage(err, "failed to get host info")
	}

	h.HostName = info.Hostname
	h.OS = info.OS
	h.Platform = info.Platform
	h.PlatformFamily = info.PlatformFamily
	h.PlatformVersion = info.PlatformVersion
	h.KernelVersion = info.KernelVersion

	return nil
}

func (h *hostIPs) sync(ctx context.Context, sgClients sg.Clients, uid domain.UUID, ncnf host.NetConf) error {
	client, err := sgClients.Hosts()
	if err != nil {
		return err
	}
	var host domain.ResourceIdentifier
	if host, err = conf.GetHostID(ctx); err != nil {
		return err
	}

	if err = h.load(ncnf); err != nil {
		return err
	}
	_, err = client.UpdIPs(ctx, &sgv1.HostReq_UpdIPs{
		Hosts: []*sgv1.HostReq_UpdIPs_Host{
			{
				Metadata: &common.MetadataScope{
					Uid:       uid.String(),
					Name:      host.Name.String(),
					Namespace: host.Namespace.String(),
				},
				Spec: &sgv1.HostReq_UpdIPs_Host_Spec{
					Ips: &common.IPs{
						Ipv4: misc.SliceToStringFunc(h.IPv4.Values(), func(ip netip.Addr) string {
							return ip.String()
						}),
						Ipv6: misc.SliceToStringFunc(h.IPv6.Values(), func(ip netip.Addr) string {
							return ip.String()
						}),
					},
				},
			},
		},
	})
	if misc.IsIn(status.Code(err), codes.NotFound, codes.InvalidArgument) {
		err = ErrHostNotFound
	}
	return err
}

func (h *hostIPs) load(ncnf host.NetConf) (err error) {
	v4, v6 := ncnf.LocalIPs()
	for _, ip := range append(host.IPvSet2List(v4).NetIPs(), host.IPvSet2List(v6).NetIPs()...) {
		if addr, ok := netip.AddrFromSlice(ip); ok {
			if addr.Is4() {
				h.IPv4.Put(addr)
			} else {
				h.IPv6.Put(addr)
			}
		} else {
			return errors.Errorf("bad local IP '%s'", ip.String())
		}
	}
	if h.IPv4.Len() == 0 && h.IPv6.Len() == 0 {
		err = errors.New("no any host IP is provided")
	}
	return err
}

// ErrHostNotFound -
var ErrHostNotFound = errors.New("host not found")
