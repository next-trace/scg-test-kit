package examples

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/next-trace/scg-test-kit"
)

func TestBrowserExample(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	h := testkit.NewBrowserHarness(t, handler)

	var resp struct {
		Status string `json:"status"`
	}
	testkit.Get(t, h, "/health", &resp)

	if resp.Status != "ok" {
		t.Errorf("expected ok, got %s", resp.Status)
	}
}
