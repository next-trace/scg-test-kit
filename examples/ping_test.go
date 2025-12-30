//go:build examples

package examples

import (
	"net/http"
	"testing"

	"github.com/next-trace/scg-test-kit"
)

func Test_Ping_Example(t *testing.T) {
	h := testkit.NewBrowserHarness(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ping" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	}))

	resp := testkit.Get(t, h, "/ping", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
