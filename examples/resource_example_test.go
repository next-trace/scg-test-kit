package examples

import (
	"fmt"
	"testing"

	"github.com/next-trace/scg-test-kit"
)

// GenericClient is a fake client for example purposes.
type GenericClient struct {
	ID string
}

func (c *GenericClient) Close() error {
	fmt.Printf("Closing client %s\n", c.ID)
	return nil
}

func TestResourceExample(t *testing.T) {
	// 1. Service provisions its own dependency
	client := &GenericClient{ID: "resource-123"}
	cleanup := client.Close

	// 2. Inject into Harness
	h := testkit.NewIntegrationHarness(t,
		testkit.WithResource("my-resource", client, cleanup),
	)

	// 3. Retrieve and use
	res, ok := testkit.Resource[*GenericClient](h, "my-resource")
	if !ok {
		t.Fatal("resource not found")
	}

	if res.ID != "resource-123" {
		t.Errorf("unexpected ID: %s", res.ID)
	}
}
