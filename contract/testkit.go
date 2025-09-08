// Package contract defines the public interfaces exposed by scg-test-kit.
package contract

import (
	"testing"

	db_contract "github.com/next-trace/scg-database/contract"
	bus_contract "github.com/next-trace/scg-service-bus/contract/bus"
	"google.golang.org/grpc"
)

// TestKit is the main entry point interface for the test kit. It exposes
// handles to sub-components and the underlying *testing.T instance.
// This contract must not leak any concrete implementation details.
//
// Components:
// - DB: database lifecycle and connection access
// - GRPC: in-memory gRPC client connection for API testing
//
// Keep this package 100% free of concrete tech details.
// Only depend on other libraries' contracts (e.g., scg-database/contract).
// This aligns with SOLID, KISS, and DRY principles.
//
// Inspired by diabuddy_testkit in terms of developer experience.
// The library provides consistent building blocks for unit, integration,
// and end-to-end tests across services.
//
// Methods:
// - T: returns the testing.T instance for advanced control (Cleanup, TempDir, etc.)
// - DB: returns the database component interface
// - GRPC: returns the gRPC component interface
//
// Implementations are provided in internal packages of this module.
// Users should interact with the contract via the root New() constructor.
// See root package for orchestration.
//
//go:generate echo "contract package contains interfaces only"
type TestKit interface {
	T() *testing.T
	DB() DB
	GRPC() GRPC
	Bus() bus_contract.Bus
}

// DB is the database component contract.
// It returns an abstract database interface from scg-database.
// No ORM or concrete driver details are exposed here.
type DB interface {
	Conn() db_contract.Connection
}

// GRPC is the gRPC component contract.
// It exposes a client connection to the in-memory test server.
type GRPC interface {
	ClientConn() *grpc.ClientConn
}
