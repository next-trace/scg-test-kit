package testkit

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHarness_Presets(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		h := New(t)
		if h == nil {
			t.Fatal("expected harness to be created")
		}
		if h.T() != t {
			t.Errorf("expected T() to return testing.TB")
		}
	})

	t.Run("NewHarness", func(t *testing.T) {
		h := NewHarness(t)
		if h == nil {
			t.Fatal("expected harness to be created")
		}
	})

	t.Run("NewUnitHarness", func(t *testing.T) {
		h := NewUnitHarness(t)
		if h == nil {
			t.Fatal("expected harness to be created")
		}
	})

	t.Run("NewIntegrationHarness", func(t *testing.T) {
		h := NewIntegrationHarness(t)
		if h == nil {
			t.Fatal("expected harness to be created")
		}
	})

	t.Run("NewBrowserHarness", func(t *testing.T) {
		h := NewBrowserHarness(t, http.NotFoundHandler())
		if h == nil {
			t.Fatal("expected harness to be created")
		}
		_, ok := h.Resource(HTTPResourceName)
		if !ok {
			t.Error("expected HTTPServer resource to be present")
		}
	})

	t.Run("NewBrowserHarness_NoHandler", func(t *testing.T) {
		h := NewBrowserHarness(t, nil)
		if h == nil {
			t.Fatal("expected harness to be created")
		}
		_, ok := h.Resource(HTTPResourceName)
		if ok {
			t.Error("expected HTTPServer resource to be absent when handler is nil")
		}
	})
}

func TestHarness_ResourceAPI(t *testing.T) {
	h := New(t)

	t.Run("WithResource and Resource", func(t *testing.T) {
		WithResource("test", "value", nil)(h)
		val, ok := Resource[string](h, "test")
		if !ok || val != "value" {
			t.Errorf("expected 'value', got '%v' (ok=%v)", val, ok)
		}
	})

	t.Run("Resource_NotFound", func(t *testing.T) {
		_, ok := Resource[string](h, "not_found")
		if ok {
			t.Error("expected ok=false for non-existent resource")
		}
	})

	t.Run("Resource_WrongType", func(t *testing.T) {
		WithResource("int", 123, nil)(h)
		_, ok := Resource[string](h, "int")
		if ok {
			t.Error("expected ok=false for wrong type resource")
		}
	})
}

func TestHarness_Cleanup(t *testing.T) {
	t.Run("Cleanup is called", func(t *testing.T) {
		cleaned := false
		{
			h := New(t)
			h.RegisterCleanup(func() {
				cleaned = true
			})
			h.Cleanup()
		}
		if !cleaned {
			t.Error("expected cleanup to be called")
		}
	})

	t.Run("Cleanup is idempotent", func(t *testing.T) {
		count := 0
		{
			h := New(t)
			h.RegisterCleanup(func() {
				count++
			})
			h.Cleanup()
			h.Cleanup()
		}
		if count != 1 {
			t.Errorf("expected cleanup to run once, got %d", count)
		}
	})

	t.Run("Cleanup order is LIFO", func(t *testing.T) {
		var order []int
		{
			h := New(t)
			h.RegisterCleanup(func() { order = append(order, 1) })
			h.RegisterCleanup(func() { order = append(order, 2) })
			h.Cleanup()
		}
		if len(order) != 2 || order[0] != 2 || order[1] != 1 {
			t.Errorf("expected [2 1], got %v", order)
		}
	})
}

func TestHarness_JSONHelpers(t *testing.T) {
	t.Run("EncodeJSON", func(t *testing.T) {
		val := map[string]string{"foo": "bar"}
		reader := EncodeJSON(t, val)
		var decoded map[string]string
		if err := json.NewDecoder(reader).Decode(&decoded); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if decoded["foo"] != "bar" {
			t.Errorf("expected bar, got %s", decoded["foo"])
		}
	})

	t.Run("DecodeJSON", func(t *testing.T) {
		reader := EncodeJSON(t, map[string]string{"foo": "bar"})
		var decoded map[string]string
		DecodeJSON(t, reader, &decoded)
		if decoded["foo"] != "bar" {
			t.Errorf("expected bar, got %s", decoded["foo"])
		}
	})
}

func TestHarness_HTTPHelpers(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = json.NewEncoder(w).Encode(body)
		} else {
			_ = json.NewEncoder(w).Encode(map[string]string{"method": r.Method})
		}
	})

	h := NewBrowserHarness(t, handler)

	t.Run("Get", func(t *testing.T) {
		var res map[string]string
		resp := Get(t, h, "/", &res)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
		if res["method"] != "GET" {
			t.Errorf("expected GET, got %s", res["method"])
		}

		// Test with nil target
		Get(t, h, "/", nil)
	})

	t.Run("Post", func(t *testing.T) {
		var res map[string]string
		body := map[string]string{"hello": "world"}
		Post(t, h, "/", body, &res)
		if res["hello"] != "world" {
			t.Errorf("expected world, got %s", res["hello"])
		}
	})

	t.Run("Post_NoBody", func(t *testing.T) {
		Post(t, h, "/", nil, nil)
	})
}

func TestHarness_HTTPHelpers_NoServer(t *testing.T) {
	h := New(t)

	// We use a mock testing.TB to capture the Fatal call
	mockT := &mockTB{TB: t}

	Get(mockT, h, "/", nil)
	if !mockT.failed {
		t.Error("expected Get to fail when HTTPServer is missing")
	}

	mockT.failed = false
	Post(mockT, h, "/", nil, nil)
	if !mockT.failed {
		t.Error("expected Post to fail when HTTPServer is missing")
	}

	mockT.failed = false
	WithResource(HTTPResourceName, "not a server", nil)(h)
	Get(mockT, h, "/", nil)
	if !mockT.failed {
		t.Error("expected Get to fail when HTTPServer is wrong type")
	}

	mockT.failed = false
	Post(mockT, h, "/", nil, nil)
	if !mockT.failed {
		t.Error("expected Post to fail when HTTPServer is wrong type")
	}
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
