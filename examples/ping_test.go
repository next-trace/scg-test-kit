//go:build examples

package examples

import (
	"net/http"
	"testing"

	"github.com/next-trace/scg-test-kit/suite"
	"github.com/next-trace/scg-test-kit/testkit"
)

func Test_Ping_Example(t *testing.T) {
	tk := suite.TestCase(t,
		testkit.WithGRPCRegister(func(mux *http.ServeMux) {
			mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("pong"))
			})
		}),
		testkit.WithInMemoryBus(),
	)

	if tk.GRPC().ClientConn() == nil {
		t.Fatal("nil client")
	}
}
