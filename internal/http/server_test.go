package http

import (
	"encoding/json"
	"net/http"
	"strings"
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
	defer func() {
		_ = cleanup()
	}()

	t.Run("BaseURL", func(t *testing.T) {
		if srv.BaseURL() == "" {
			t.Error("expected non-empty base URL")
		}
	})

	t.Run("Client", func(t *testing.T) {
		if srv.Client() == nil {
			t.Error("expected non-nil client")
		}
	})

	t.Run("Close", func(t *testing.T) {
		if err := srv.Close(); err != nil {
			t.Errorf("expected nil error on Close, got %v", err)
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

		// Test with nil target
		srv.Get(t, "/", nil)
	})

	t.Run("Post", func(t *testing.T) {
		var res map[string]string
		body := map[string]string{"hello": "world"}
		resp := srv.Post(t, "/", body, &res)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
		if res["hello"] != "world" {
			t.Errorf("expected world, got %s", res["hello"])
		}
	})

	t.Run("Post_NoBody", func(t *testing.T) {
		resp := srv.Post(t, "/", nil, nil)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestServer_JSONHelpers_Fail(t *testing.T) {
	mockT := &mockTB{TB: t}

	t.Run("EncodeJSON_Fail", func(t *testing.T) {
		EncodeJSON(mockT, make(chan int))
		if !mockT.failed {
			t.Error("expected EncodeJSON to fail for unmarshallable type")
		}
	})

	t.Run("DecodeJSON_Fail", func(t *testing.T) {
		mockT.failed = false
		DecodeJSON(mockT, strings.NewReader("invalid json"), nil)
		if !mockT.failed {
			t.Error("expected DecodeJSON to fail for invalid json")
		}
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
