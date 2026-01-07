// nolint:revive // package name is intentional
package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestServer_Helpers(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(body)
		} else {
			_ = json.NewEncoder(w).Encode(map[string]string{"method": r.Method})
		}
	})

	srv, cleanup := NewServer(t, handler)
	defer func() { _ = cleanup() }()

	t.Run("BaseURL", func(t *testing.T) {
		if srv.BaseURL() == "" {
			t.Error("expected BaseURL to be non-empty")
		}
	})

	t.Run("Client", func(t *testing.T) {
		if srv.Client() == nil {
			t.Error("expected Client to be non-nil")
		}
	})

	t.Run("Close", func(t *testing.T) {
		if err := srv.Close(); err != nil {
			t.Errorf("expected Close to be nil, got %v", err)
		}
	})

	t.Run("Get", func(t *testing.T) {
		var res map[string]string
		resp := srv.Get(t, "/", &res)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
		if res["method"] != "GET" {
			t.Errorf("expected GET, got %s", res["method"])
		}
	})

	t.Run("Post", func(t *testing.T) {
		var res map[string]string
		body := map[string]string{"hello": "world"}
		srv.Post(t, "/", body, &res)
		if res["hello"] != "world" {
			t.Errorf("expected world, got %s", res["hello"])
		}
	})

	t.Run("Post_NoBody", func(t *testing.T) {
		srv.Post(t, "/", nil, nil)
	})
}

func TestServer_JSONHelpers_Fail(t *testing.T) {
	mockT := &mockTB{TB: t}

	t.Run("EncodeJSON_Fail", func(t *testing.T) {
		// Channels are not marshallable to JSON
		EncodeJSON(mockT, make(chan int))
		if !mockT.failed {
			t.Error("expected EncodeJSON to fail")
		}
		mockT.failed = false
	})

	t.Run("DecodeJSON_Fail", func(t *testing.T) {
		reader := bytes.NewReader([]byte("not-json"))
		DecodeJSON(mockT, reader, nil)
		if !mockT.failed {
			t.Error("expected DecodeJSON to fail")
		}
		mockT.failed = false
	})
}

type mockTB struct {
	testing.TB
	failed bool
}

func (m *mockTB) Fatal(_ ...any) {
	m.failed = true
}

func (m *mockTB) Fatalf(_ string, _ ...any) {
	m.failed = true
}
