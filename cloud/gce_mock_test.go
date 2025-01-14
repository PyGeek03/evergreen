package cloud

import (
	"context"
	"errors"

	"github.com/evergreen-ci/evergreen/model/host"
	"golang.org/x/oauth2/jwt"
	compute "google.golang.org/api/compute/v1"
)

type gceClientMock struct {
	// API call options
	failInit   bool
	failCreate bool
	failGet    bool
	failDelete bool

	// Other options
	isActive        bool
	hasAccessConfig bool
}

func (c *gceClientMock) Init(context.Context, *jwt.Config) error {
	if c.failInit {
		return errors.New("failed to initialize client")
	}

	return nil
}

// CreateInstance returns a unique identifier for the mock instance.
func (c *gceClientMock) CreateInstance(h *host.Host, _ *GCESettings) (string, error) {
	if c.failCreate {
		return "", errors.New("failed to create instance")
	}

	return h.Id, nil
}

func (c *gceClientMock) GetInstance(_ *host.Host) (*compute.Instance, error) {
	if c.failGet {
		return nil, errors.New("failed to get instance")
	}

	instance := &compute.Instance{
		Status:      "RUNNING",
		Zone:        "us-east1-c",
		MachineType: "zones/us-east1-c/machineTypes/n1-standard-8",
	}

	if !c.isActive {
		instance.Status = "STOPPING"
	}

	if c.hasAccessConfig {
		instance.NetworkInterfaces = []*compute.NetworkInterface{&compute.NetworkInterface{
			AccessConfigs: []*compute.AccessConfig{
				&compute.AccessConfig{NatIP: "0.0.0.0"},
			},
		}}
	}

	return instance, nil
}

func (c *gceClientMock) DeleteInstance(_ *host.Host) error {
	if c.failDelete {
		return errors.New("failed to delete instance")
	}

	return nil
}
