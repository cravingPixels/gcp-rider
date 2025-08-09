package gcp

import (
	"context"
	"errors"
	"reflect"
	"testing"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
)

// mockInstancesAPI is a mock implementation of the instancesAPI interface.
type mockInstancesAPI struct {
	// A function we can override in each test to simulate different responses.
	AggregatedListFunc func(ctx context.Context, req *computepb.AggregatedListInstancesRequest) *compute.InstancesIterator
}

func (m *mockInstancesAPI) AggregatedList(ctx context.Context, req *computepb.AggregatedListInstancesRequest) *compute.InstancesIterator {
	return m.AggregatedListFunc(ctx, req)
}

// mockPager is a mock implementation of the iterator.Pager interface.
type mockPager struct {
	items []*computepb.InstancesScopedList
	err   error
	index int
}

func (p *mockPager) NextPage(pageInfo *iterator.PageInfo, dst interface{}) (string, error) {
	if p.err != nil {
		return "", p.err
	}
	if p.index >= len(p.items) {
		return iterator.Done, nil
	}
	// This is a bit of a hack to get the items into the iterator's internal state.
	reflect.ValueOf(dst).Elem().Set(reflect.ValueOf(p.items))
	p.index = len(p.items) // Mark as done for the next call
	return "next-page-token", nil
}

func TestFetchInstances_Success(t *testing.T) {
	vm1Name, vm1Zone := "instance-1", "us-central1-a"
	vm2Name, vm2Zone := "instance-2", "europe-west1-b"
	zoneURL1 := "https://www.googleapis.com/compute/v1/projects/proj/zones/" + vm1Zone
	zoneURL2 := "https://www.googleapis.com/compute/v1/projects/proj/zones/" + vm2Zone

	mockAPI := &mockInstancesAPI{
		AggregatedListFunc: func(ctx context.Context, req *computepb.AggregatedListInstancesRequest) *compute.InstancesIterator {
			return &compute.InstancesIterator{
				Pager: &mockPager{
					items: []*computepb.InstancesScopedList{
						{
							Instances: []*computepb.Instance{
								{Name: &vm1Name, Zone: &zoneURL1},
								{Name: &vm2Name, Zone: &zoneURL2},
							},
						},
					},
				},
			}
		},
	}

	client := &Client{api: mockAPI}
	instances, err := client.FetchInstances(context.Background(), "test-project")

	if err != nil {
		t.Fatalf("FetchInstances() returned an unexpected error: %v", err)
	}

	expected := []Instance{
		{Name: "instance-1", Zone: "us-central1-a"},
		{Name: "instance-2", Zone: "europe-west1-b"},
	}

	if !reflect.DeepEqual(instances, expected) {
		t.Errorf("expected instances %v, got %v", expected, instances)
	}
}

func TestFetchInstances_Error(t *testing.T) {
	expectedErr := errors.New("GCP API error")
	mockAPI := &mockInstancesAPI{
		AggregatedListFunc: func(ctx context.Context, req *computepb.AggregatedListInstancesRequest) *compute.InstancesIterator {
			return &compute.InstancesIterator{
				Pager: &mockPager{err: expectedErr},
			}
		},
	}

	client := &Client{api: mockAPI}
	_, err := client.FetchInstances(context.Background(), "test-project")

	if err == nil {
		t.Fatal("FetchInstances() did not return an error when one was expected")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error containing '%v', got '%v'", expectedErr, err)
	}
}
