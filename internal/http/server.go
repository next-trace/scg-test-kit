// Package http provides internal implementation for HTTP test helpers.
//
// nolint:revive // package name is intentional
package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Server holds the state of a test HTTP server.
type Server struct {
	baseURL string
	client  *http.Client
}

func (s *Server) BaseURL() string      { return s.baseURL }
func (s *Server) Client() *http.Client { return s.client }
func (s *Server) Close() error         { return nil } // httptest.Server is closed by teardown

// NewServer creates a new httptest.Server and returns a Server helper and a cleanup function.
func NewServer(t testing.TB, handler http.Handler) (*Server, func() error) {
	server := httptest.NewServer(handler)

	cleanup := func() error {
		server.Close()
		return nil
	}

	return &Server{
		baseURL: server.URL,
		client:  server.Client(),
	}, cleanup
}

// EncodeJSON encodes the given value into an io.Reader.
func EncodeJSON(t testing.TB, value any) io.Reader {
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	return bytes.NewReader(data)
}

// DecodeJSON decodes the JSON from the reader into the target value.
func DecodeJSON(t testing.TB, reader io.Reader, target any) {
	if err := json.NewDecoder(reader).Decode(target); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
}

// Get performs a GET request to the given path and decodes the response into the target value.
func (s *Server) Get(t testing.TB, path string, target any) *http.Response {
	response, err := s.client.Get(s.baseURL + path)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
		return nil
	}
	defer func() {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
	}()

	if target != nil {
		DecodeJSON(t, response.Body, target)
	}
	return response
}

// Post performs a POST request with a JSON body and decodes the response into the target value.
func (s *Server) Post(t testing.TB, path string, body any, target any) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = EncodeJSON(t, body)
	}

	response, err := s.client.Post(s.baseURL+path, "application/json", bodyReader)
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
		return
	}
	defer func() {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
	}()

	if target != nil {
		DecodeJSON(t, response.Body, target)
	}
}
