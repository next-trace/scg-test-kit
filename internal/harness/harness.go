// Package harness provides the internal implementation of the test harness.
package harness

import (
	"sync"
	"testing"
)

// Harness is a generic container for test resources.
type Harness struct {
	t testing.TB

	mu        sync.RWMutex
	resources map[string]any
}

// New creates a new Harness instance.
func New(t testing.TB) *Harness {
	return &Harness{
		t:         t,
		resources: make(map[string]any),
	}
}

// T returns the underlying testing.TB instance.
func (h *Harness) T() testing.TB {
	return h.t
}

// SetResource adds a named resource to the harness and registers its cleanup if provided.
func (h *Harness) SetResource(name string, value any, cleanup func() error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.resources[name] = value

	if cleanup != nil {
		h.t.Cleanup(func() {
			if err := cleanup(); err != nil {
				h.t.Errorf("cleanup %s failed: %v", name, err)
			}
		})
	}
}

// Resource retrieves a named resource from the harness.
func (h *Harness) Resource(name string) (any, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	val, ok := h.resources[name]
	return val, ok
}
