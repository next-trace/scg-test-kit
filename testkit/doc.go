// Package testkit contains the root orchestrator and functional options used
// to compose a TestKit instance for unit and integration tests.
//
// It wires together contract-only components (DB, GRPC, Bus) and registers
// teardowns via testing.T.
package testkit
