package examples

import (
	"fmt"
	"testing"

	"github.com/next-trace/scg-test-kit"
)

type ClosableResource struct {
	Name string
}

func (r *ClosableResource) Cleanup() error {
	fmt.Printf("Cleaning up resource %s\n", r.Name)
	return nil
}

func TestCleanupExample(t *testing.T) {
	res := &ClosableResource{Name: "temporary-resource"}

	// Harness automatically calls the cleanup function when the test finishes
	testkit.NewIntegrationHarness(t,
		testkit.WithResource("res", res, res.Cleanup),
	)

	// No manual cleanup call needed here; Harness handles it via t.Cleanup
}
