package gcp

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"google.golang.org/api/option"
)

func TestFetchInstances_Success_WithMockServer(t *testing.T) {
	// The mock server will return a canned JSON response that mimics the real API.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a sample response for a VM list.
		jsonResponse := `{
			"items": {
				"zones/us-central1-a": {
					"instances": [
						{
							"name": "instance-1",
							"zone": "https://www.googleapis.com/compute/v1/projects/proj/zones/us-central1-a"
						}
					]
				},
				"zones/europe-west1-b": {
					"instances": [
						{
							"name": "instance-2",
							"zone": "https://www.googleapis.com/compute/v1/projects/proj/zones/europe-west1-b"
						}
					]
				}
			}
		}`
		fmt.Fprintln(w, jsonResponse)
	}))
	defer mockServer.Close()

	// Create a client that connects to our mock server instead of the real GCP.
	// We use `option.WithEndpoint` to point the client to our test server.
	ctx := context.Background()
	client, err := NewClient(ctx, option.WithEndpoint(mockServer.URL), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("Failed to create client for test: %v", err)
	}

	instances, err := client.FetchInstances(ctx, "test-project")
	if err != nil {
		t.Fatalf("FetchInstances() returned an unexpected error: %v", err)
	}

	expected := []Instance{
		{Name: "instance-1", Zone: "us-central1-a"},
		{Name: "instance-2", Zone: "europe-west1-b"},
	}

	// The order of items from a map is not guaranteed, so we need to sort for a stable test.
	// For this test, we'll just check the length and content in a flexible way.
	if len(instances) != len(expected) {
		t.Fatalf("expected %d instances, got %d", len(expected), len(instances))
	}

	// A simple check to see if the expected instances are present.
	foundCount := 0
	for _, exp := range expected {
		for _, got := range instances {
			if reflect.DeepEqual(exp, got) {
				foundCount++
			}
		}
	}
	if foundCount != len(expected) {
		t.Errorf("did not find all expected instances. Expected: %v, Got: %v", expected, instances)
	}
}

func TestFetchInstances_Error_WithMockServer(t *testing.T) {
	// This mock server will return an error status code.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	ctx := context.Background()
	client, err := NewClient(ctx, option.WithEndpoint(mockServer.URL), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("Failed to create client for test: %v", err)
	}

	_, err = client.FetchInstances(ctx, "test-project")
	if err == nil {
		t.Fatal("FetchInstances() did not return an error when one was expected")
	}
}
