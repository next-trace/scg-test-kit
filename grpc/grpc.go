// Package grpc provides an in-memory gRPC client/server for testing via bufconn.
package grpc

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/next-trace/scg-test-kit/contract"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type inMemoryGRPC struct {
	conn *grpc.ClientConn
}

func (g *inMemoryGRPC) ClientConn() *grpc.ClientConn { return g.conn }

// New creates an in-memory gRPC server and returns a client connection to it.
// Pass a register function to attach your HTTP/2 handlers onto mux (connect-go or grpc-gateway, etc.).
func New(t *testing.T, register func(mux *http.ServeMux)) (contract.GRPC, func()) {
	if t == nil {
		panic("nil *testing.T")
	}
	listener := bufconn.Listen(bufSize)

	mux := http.NewServeMux()
	if register != nil {
		register(mux)
	}
	server := &http.Server{
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() { _ = server.Serve(&bufListener{Listener: listener}) }()

	dialer := func(_ context.Context, _ string) (net.Conn, error) {
		return listener.Dial()
	}
	conn, err := grpc.DialContext( //nolint:staticcheck // grpc.DialContext is deprecated but supported throughout 1.x; acceptable for in-memory bufconn testing
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}

	impl := &inMemoryGRPC{conn: conn}
	teardown := func() {
		_ = conn.Close()
		_ = server.Close()
		_ = listener.Close()
	}
	return impl, teardown
}

// bufListener adapts a *bufconn.Listener to implement net.Listener for http.Server.
type bufListener struct{ *bufconn.Listener }

func (b *bufListener) Accept() (net.Conn, error) { return b.Listener.Accept() }
func (b *bufListener) Close() error              { return b.Listener.Close() }
func (b *bufListener) Addr() net.Addr            { return b.Listener.Addr() }
