package testkit

import (
	"io"
	"net/http"
	"testing"

	"github.com/next-trace/scg-test-kit/internal/harness"
	http_internal "github.com/next-trace/scg-test-kit/internal/http"
)

// Harness is the main test harness.
type Harness = harness.Harness

// Option configures the Harness.
type Option func(*Harness)

// HTTPResourceName is the name used to store the HTTP server in harness resources.
const HTTPResourceName = "HTTPServer"

// NewHarness creates a new Harness with the given options.
func NewHarness(t testing.TB, opts ...Option) *Harness {
	h := harness.New(t)
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// NewUnitHarness creates a Harness optimized for unit tests.
func NewUnitHarness(t testing.TB, opts ...Option) *Harness {
	return NewHarness(t, opts...)
}

// NewIntegrationHarness creates a Harness for integration tests.
func NewIntegrationHarness(t testing.TB, opts ...Option) *Harness {
	return NewHarness(t, opts...)
}

// NewBrowserHarness creates a Harness for browser-like HTTP tests.
func NewBrowserHarness(t testing.TB, handler http.Handler, opts ...Option) *Harness {
	h := NewHarness(t, opts...)
	if handler != nil {
		WithHTTPServer(handler)(h)
	}
	return h
}

// WithResource adds a named resource to the harness.
func WithResource(name string, value any, cleanup func() error) Option {
	return func(h *Harness) {
		h.SetResource(name, value, cleanup)
	}
}

// Resource retrieves a named resource from the harness.
func Resource[T any](h *Harness, name string) (T, bool) {
	var zero T
	val, ok := h.Resource(name)
	if !ok {
		return zero, false
	}
	typed, ok := val.(T)
	return typed, ok
}

// WithHTTPServer plugs an HTTP server capability into the harness.
func WithHTTPServer(handler http.Handler) Option {
	return func(h *Harness) {
		server, cleanup := http_internal.NewServer(h.T(), handler)
		h.SetResource(HTTPResourceName, server, cleanup)
	}
}

// EncodeJSON encodes the given value into an io.Reader.
func EncodeJSON(t testing.TB, value any) io.Reader {
	return http_internal.EncodeJSON(t, value)
}

// DecodeJSON decodes the JSON from the reader into the target value.
func DecodeJSON(t testing.TB, reader io.Reader, target any) {
	http_internal.DecodeJSON(t, reader, target)
}

// Get performs a GET request to the given path and decodes the response into the target value.
func Get(t testing.TB, h *Harness, path string, target any) *http.Response {
	val, ok := h.Resource(HTTPResourceName)
	if !ok {
		t.Fatal("HTTPServer resource not available")
		return nil
	}
	srv, ok := val.(interface {
		BaseURL() string
		Client() *http.Client
	})
	if !ok {
		t.Fatal("HTTPServer resource does not implement required interface")
		return nil
	}

	resp, err := srv.Client().Get(srv.BaseURL() + path)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
		return nil
	}
	defer func() { _ = resp.Body.Close() }()

	if target != nil {
		DecodeJSON(t, resp.Body, target)
	}
	return resp
}

// Post performs a POST request with a JSON body and decodes the response into the target value.
func Post(t testing.TB, h *Harness, path string, body any, target any) {
	val, ok := h.Resource(HTTPResourceName)
	if !ok {
		t.Fatal("HTTPServer resource not available")
		return
	}
	srv, ok := val.(interface {
		BaseURL() string
		Client() *http.Client
	})
	if !ok {
		t.Fatal("HTTPServer resource does not implement required interface")
		return
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = EncodeJSON(t, body)
	}
	resp, err := srv.Client().Post(srv.BaseURL()+path, "application/json", bodyReader)
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if target != nil {
		DecodeJSON(t, resp.Body, target)
	}
}
