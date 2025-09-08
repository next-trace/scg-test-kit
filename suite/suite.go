// Package suite provides a simple DX facade around the testkit constructors.
package suite

import (
	"testing"

	"github.com/next-trace/scg-test-kit/contract"
	"github.com/next-trace/scg-test-kit/testkit"
)

// TestCase returns a minimal kit. Compose behavior via options.
func TestCase(t *testing.T, opts ...testkit.Option) contract.TestKit {
	return testkit.NewTestCase(t, opts...)
}

// IntegrationTest returns a kit for integration tests (same builder, different intent).
func IntegrationTest(t *testing.T, opts ...testkit.Option) contract.TestKit {
	return testkit.NewIntegrationTest(t, opts...)
}
