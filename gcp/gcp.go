// Package gcp handles interactions with the Google Cloud Platform API.
package gcp

import (
	"context"
	"fmt"
	"path"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
)

// Instance holds the essential information for a GCP VM instance.
type Instance struct {
	Name string
	Zone string
}

// instancesAPI is an interface that abstracts the GCP compute client.
// This allows us to mock the client in tests.
type instancesAPI interface {
	AggregatedList(context.Context, *computepb.AggregatedListInstancesRequest) *compute.InstancesIterator
}

// Client is a wrapper around the GCP compute client.
type Client struct {
	api instancesAPI
}

// NewClient creates a new real GCP client.
func NewClient(ctx context.Context) (*Client, error) {
	c, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create instances client: %w", err)
	}
	return &Client{api: c}, nil
}

// FetchInstances retrieves a list of VM instances from a given project.
func (c *Client) FetchInstances(ctx context.Context, projectID string) ([]Instance, error) {
	req := &computepb.AggregatedListInstancesRequest{
		Project: projectID,
	}
	it := c.api.AggregatedList(ctx, req)
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
