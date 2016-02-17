package discovery

import (
	"github.com/kyani-inc/ecs-discovery/discover"
)

const (
	ERR_NO_CLIENT = "no discovery client set, please call NewClient()"
)

type (
	Discoverer interface {
		Discover() error

		NewClient(cluster, region, defaultdomain string)
	}

	discovery struct {
		*discover.Client
	}
)

// NewClient creates a new discovery client, it is part of the interface
// should be invoked after the provider has already been initialized.
// Example:
// kv, _ := ConsulKV(nil)
// kv.NewClient(nil)
// kv.Discover()
func (discovery *discovery) NewClient(cluster, region, defaultdomain string) {
	discovery.Client = discover.NewClient(cluster, region, defaultdomain)
}
