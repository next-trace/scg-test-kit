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
	cleanups  []func()
	cleanOnce sync.Once
}

// New creates a new Harness instance.
func New(t testing.TB) *Harness {
	h := &Harness{
		t:         t,
		resources: make(map[string]any),
		cleanups:  make([]func(), 0),
	}
	// Automatically register Cleanup to run at the end of the test
	t.Cleanup(h.Cleanup)
	return h
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
		// Wrap cleanup to handle error logging
		wrapped := func() {
			if err := cleanup(); err != nil {
				h.t.Errorf("cleanup %s failed: %v", name, err)
			}
		}
		h.cleanups = append(h.cleanups, wrapped)
	}
}

// Resource retrieves a named resource from the harness.
func (h *Harness) Resource(name string) (any, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	val, ok := h.resources[name]
	return val, ok
}

// RegisterCleanup registers a function to be run during cleanup.
// Cleanups are run in LIFO order (Last In, First Out).
func (h *Harness) RegisterCleanup(fn func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cleanups = append(h.cleanups, fn)
}

// Cleanup runs all registered cleanups in LIFO order.
// It is idempotent and safe to call multiple times.
func (h *Harness) Cleanup() {
	h.cleanOnce.Do(func() {
		h.mu.Lock()
		// Copy cleanups to allow releasing the lock while running them
		ops := make([]func(), len(h.cleanups))
		copy(ops, h.cleanups)
		h.mu.Unlock()

		// Run cleanups in reverse order
		for i := len(ops) - 1; i >= 0; i-- {
			ops[i]()
		}
	})
}

// Close is an alias for Cleanup.
func (h *Harness) Close() {
	h.Cleanup()
}
