package harness

import (
	"errors"
	"testing"
)

type mockTB struct {
	testing.TB
	cleanupFuncs []func()
	errors       []string
}

func (m *mockTB) Cleanup(f func()) {
	m.cleanupFuncs = append(m.cleanupFuncs, f)
}

func (m *mockTB) Errorf(format string, _ ...any) {
	m.errors = append(m.errors, format)
}

func TestHarness(t *testing.T) {
	mtb := &mockTB{}
	h := New(mtb)

	cleanupCalled := false
	h.SetResource("test", "value", func() error {
		cleanupCalled = true
		return nil
	})

	val, ok := h.Resource("test")
	if !ok || val != "value" {
		t.Errorf("expected value, got %v", val)
	}

	// Trigger cleanups (manually since we mocked TB)
	for i := len(mtb.cleanupFuncs) - 1; i >= 0; i-- {
		mtb.cleanupFuncs[i]()
	}

	if !cleanupCalled {
		t.Error("cleanup was not called")
	}
}

func TestHarness_CleanupError(t *testing.T) {
	mtb := &mockTB{}
	h := New(mtb)

	h.SetResource("test", "value", func() error {
		return errors.New("boom")
	})

	// Trigger cleanups
	for i := len(mtb.cleanupFuncs) - 1; i >= 0; i-- {
		mtb.cleanupFuncs[i]()
	}

	if len(mtb.errors) == 0 {
		t.Error("expected error to be reported")
	}
}
