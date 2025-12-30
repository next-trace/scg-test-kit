package examples

import (
	"fmt"
	"testing"

	"github.com/next-trace/scg-test-kit"
)

type ExternalService struct {
	Endpoint string
}

func (s *ExternalService) Shutdown() error {
	fmt.Printf("Shutting down service at %s\n", s.Endpoint)
	return nil
}

func TestExternalDependencyExample(t *testing.T) {
	svc := &ExternalService{Endpoint: "http://localhost:8080"}

	h := testkit.NewIntegrationHarness(t,
		testkit.WithResource("external-svc", svc, svc.Shutdown),
	)

	val, ok := testkit.Resource[*ExternalService](h, "external-svc")
	if !ok {
		t.Fatal("external service not found")
	}

	if val.Endpoint != "http://localhost:8080" {
		t.Errorf("expected http://localhost:8080, got %s", val.Endpoint)
	}
}
