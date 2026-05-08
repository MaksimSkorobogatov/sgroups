package grpc

import (
	"fmt"
	"reflect"

	grpcClient "github.com/H-BF/corlib/client/grpc"
	client "github.com/PRO-Robotech/sgroups-proto/pkg"
	sgv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
)

// JsonCodecName -
const JsonCodecName = "json"

type (
	closableClient interface {
		Init(conn grpc.ClientConnInterface) error
		Close() error
	}
	// Clients SecGrpups server clients
	Clients struct {
		apiClients map[reflect.Type]closableClient
	}
)

// NewClients constructs 'sgroups' API clients
func NewClients(c grpc.ClientConnInterface) (*Clients, error) {
	cl := Clients{
		apiClients: map[reflect.Type]closableClient{
			reflect.TypeFor[sgv1.SGroupsNamespaceAPIClient]():      &client.ClosableClient[sgv1.SGroupsNamespaceAPIClient]{},
			reflect.TypeFor[sgv1.SGroupsAddressGroupsAPIClient]():  &client.ClosableClient[sgv1.SGroupsAddressGroupsAPIClient]{},
			reflect.TypeFor[sgv1.SGroupsNetworksAPIClient]():       &client.ClosableClient[sgv1.SGroupsNetworksAPIClient]{},
			reflect.TypeFor[sgv1.SGroupsHostsAPIClient]():          &client.ClosableClient[sgv1.SGroupsHostsAPIClient]{},
			reflect.TypeFor[sgv1.SGroupsServicesAPIClient]():       &client.ClosableClient[sgv1.SGroupsServicesAPIClient]{},
			reflect.TypeFor[sgv1.SGroupsRulesAPIClient]():          &client.ClosableClient[sgv1.SGroupsRulesAPIClient]{},
			reflect.TypeFor[sgv1.SGroupsNetworkBindingAPIClient](): &client.ClosableClient[sgv1.SGroupsNetworkBindingAPIClient]{},
			reflect.TypeFor[sgv1.SGroupsHostBindingAPIClient]():    &client.ClosableClient[sgv1.SGroupsHostBindingAPIClient]{},
			reflect.TypeFor[sgv1.SGroupsServiceBindingAPIClient](): &client.ClosableClient[sgv1.SGroupsServiceBindingAPIClient]{},
			reflect.TypeFor[sgv1.SGroupsStatusAPIClient]():         &client.ClosableClient[sgv1.SGroupsStatusAPIClient]{},
		},
	}
	closable := grpcClient.MakeCloseable(
		grpcClient.WithErrorWrapper(c, "sgroups"),
	)
	for _, v := range cl.apiClients {
		if err := v.Init(closable); err != nil {
			return nil, errors.WithMessage(err, "init client")
		}
	}
	return &cl, nil
}

// Namespace returns 'sgroups' Namespace API client
func (c *Clients) Namespace() (sgv1.SGroupsNamespaceAPIClient, error) {
	return getAPIclient[sgv1.SGroupsNamespaceAPIClient](c)
}

// AddressGroups returns 'sgroups' AddressGroups API client
func (c *Clients) AddressGroups() (sgv1.SGroupsAddressGroupsAPIClient, error) {
	return getAPIclient[sgv1.SGroupsAddressGroupsAPIClient](c)
}

// Networks returns 'sgroups' Networks API client
func (c *Clients) Networks() (sgv1.SGroupsNetworksAPIClient, error) {
	return getAPIclient[sgv1.SGroupsNetworksAPIClient](c)
}

// Hosts returns 'sgroups' Hosts API client
func (c *Clients) Hosts() (sgv1.SGroupsHostsAPIClient, error) {
	return getAPIclient[sgv1.SGroupsHostsAPIClient](c)
}

// Services returns 'sgroups' Services API client
func (c *Clients) Services() (sgv1.SGroupsServicesAPIClient, error) {
	return getAPIclient[sgv1.SGroupsServicesAPIClient](c)
}

// Rules returns 'sgroups' Rules API client
func (c *Clients) Rules() (sgv1.SGroupsRulesAPIClient, error) {
	return getAPIclient[sgv1.SGroupsRulesAPIClient](c)
}

// NetworkBinding returns 'sgroups' NetworkBinding API client
func (c *Clients) NetworkBinding() (sgv1.SGroupsNetworkBindingAPIClient, error) {
	return getAPIclient[sgv1.SGroupsNetworkBindingAPIClient](c)
}

// HostBinding returns 'sgroups' HostBinding API client
func (c *Clients) HostBinding() (sgv1.SGroupsHostBindingAPIClient, error) {
	return getAPIclient[sgv1.SGroupsHostBindingAPIClient](c)
}

// ServiceBinding returns 'sgroups' ServiceBinding API client
func (c *Clients) ServiceBinding() (sgv1.SGroupsServiceBindingAPIClient, error) {
	return getAPIclient[sgv1.SGroupsServiceBindingAPIClient](c)
}

// Status returns 'sgroups' Status API client
func (c *Clients) Status() (sgv1.SGroupsStatusAPIClient, error) {
	return getAPIclient[sgv1.SGroupsStatusAPIClient](c)
}

// Close closes all clients
func (c *Clients) Close() error {
	var err error
	for _, v := range c.apiClients {
		err = multierr.Append(err, v.Close())
	}
	return err
}

func getAPIclient[T any](c *Clients) (ret T, err error) {
	v, ok := c.apiClients[reflect.TypeFor[T]()]
	if !ok {
		return ret, fmt.Errorf("sgroups: client %s not registered", reflect.TypeOf(ret))
	}

	return v.(*client.ClosableClient[T]).C, nil
}
