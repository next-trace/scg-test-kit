package testkit_test

import (
	"testing"

	"github.com/next-trace/scg-test-kit/suite"
	"github.com/next-trace/scg-test-kit/testkit"
)

func TestWithInMemoryBus_And_GRPC(t *testing.T) {
	tk := suite.TestCase(t, testkit.WithInMemoryBus())
	if tk.Bus() == nil {
		t.Fatalf("expected non-nil bus")
	}
	if tk.GRPC().ClientConn() == nil {
		t.Fatalf("expected non-nil grpc client conn")
	}
}
