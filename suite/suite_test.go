package suite

import (
	"net/http"
	"testing"

	"github.com/next-trace/scg-test-kit/testkit"
)

func TestSuite_TestCaseAndIntegration(t *testing.T) {
	// TestCase should return a non-nil kit whose T() equals the provided t
	kit1 := TestCase(t)
	if kit1 == nil {
		t.Fatalf("TestCase returned nil kit")
	}
	if kit1.T() != t {
		t.Fatalf("TestCase kit.T() mismatch: got %p want %p", kit1.T(), t)
	}

	// IntegrationTest should behave equivalently (facade over same builder)
	kit2 := IntegrationTest(t)
	if kit2 == nil {
		t.Fatalf("IntegrationTest returned nil kit")
	}
	if kit2.T() != t {
		t.Fatalf("IntegrationTest kit.T() mismatch: got %p want %p", kit2.T(), t)
	}

	// Additional integration-style facade test
	reg := func(mux *http.ServeMux) {
		mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	}
	kit3 := IntegrationTest(t, testkit.WithGRPCRegister(reg), testkit.WithInMemoryBus())
	if kit3.GRPC().ClientConn() == nil {
		t.Fatalf("expected non-nil grpc client conn")
	}
	if kit3.Bus() == nil {
		t.Fatalf("expected non-nil bus")
	}
}
