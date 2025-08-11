// Package gcp handles interactions with the Google Cloud Platform API.
package gcp

import (
	"context"
	"fmt"
	"path"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Instance holds the essential information for a GCP VM instance.
type Instance struct {
	Name string
	Zone string
}

// Client is an interface for a GCP client, allowing for mock implementations.
type Client interface {
	FetchInstances(ctx context.Context, projectID string) ([]Instance, error)
	Close() error
}

// realClient is the concrete implementation of the Client interface.
type realClient struct {
	computeClient *compute.InstancesClient
}

// NewClient creates a new real GCP client that conforms to the Client interface.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	c, err := compute.NewInstancesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create instances client: %w", err)
	}
	return &realClient{computeClient: c}, nil
}

// FetchInstances retrieves a list of VM instances from a given project.
func (c *realClient) FetchInstances(ctx context.Context, projectID string) ([]Instance, error) {
	req := &computepb.AggregatedListInstancesRequest{
		Project: projectID,
	}
	it := c.computeClient.AggregatedList(ctx, req)
	var vms []Instance
	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over instances: %w", err)
		}
		if pair.Value != nil && len(pair.Value.Instances) > 0 {
			for _, instance := range pair.Value.Instances {
				zone := path.Base(*instance.Zone)
				vms = append(vms, Instance{Name: *instance.Name, Zone: zone})
			}
		}
	}
	return vms, nil
}

// Close closes the underlying client connection.
func (c *realClient) Close() error {
	return c.computeClient.Close()
}