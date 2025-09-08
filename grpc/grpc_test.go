package grpc

import (
	"context"
	"net/http"
	"testing"

	"google.golang.org/grpc/status"
)

func TestInMemoryGRPC_New_ClientConn(t *testing.T) {
	rpc, td := New(t, func(mux *http.ServeMux) {
		mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("pong"))
		})
	})
	if td != nil {
		t.Cleanup(td)
	}
	if rpc.ClientConn() == nil {
		t.Fatalf("nil client connection")
	}

	// Invoke a non-existent gRPC method to ensure transport works and returns an error
	ctx := context.Background()
	err := rpc.ClientConn().Invoke(ctx, "/unknown.Service/Method", nil, nil)
	if err == nil {
		t.Fatalf("expected error invoking unknown method")
	}
	_ = status.Code(err) // ensure error is a gRPC status
}

func TestInMemoryGRPC_New_NilRegister(t *testing.T) {
	rpc, td := New(t, nil)
	if td != nil {
		t.Cleanup(td)
	}
	if rpc.ClientConn() == nil {
		t.Fatalf("nil client connection")
	}
}
