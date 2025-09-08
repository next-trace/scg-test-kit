// Package grpc exposes an in-memory gRPC client/server for testing via bufconn.
//
// Handlers can be registered on the provided HTTP/2 mux (connect-go or others).
// The package focuses on correctness and clean teardown for tests.
package grpc
