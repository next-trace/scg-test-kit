package testkit

import (
	"net/http"
	"testing"

	db_contract "github.com/next-trace/scg-database/contract"
	bus_contract "github.com/next-trace/scg-service-bus/contract/bus"
	"github.com/next-trace/scg-service-bus/memory"
)

// Option configures the test kit builder during construction.
type Option func(*builder)

type builder struct {
	register func(*http.ServeMux)

	// DB
	dbConn    db_contract.Connection
	dbTd      func()
	dbFactory func(t *testing.T) (db_contract.Connection, func(), error)

	// Bus
	bus    bus_contract.Bus
	busTds []func()
}

// WithGRPCRegister registers HTTP/2 handlers onto the in-memory server mux.
func WithGRPCRegister(fn func(*http.ServeMux)) Option {
	return func(b *builder) { b.register = fn }
}

// WithDBInstance injects an existing db_contract.Connection with an optional teardown.
func WithDBInstance(conn db_contract.Connection, teardown func()) Option {
	return func(b *builder) {
		b.dbConn = conn
		b.dbTd = teardown
	}
}

// WithEphemeralDB sets a factory that returns a fresh db connection and its teardown.
// The factory is executed during build; scg-test-kit never owns lifecycle logic.
func WithEphemeralDB(factory func(t *testing.T) (db_contract.Connection, func(), error)) Option {
	return func(b *builder) { b.dbFactory = factory }
}

// WithBusInstance injects an app-provided bus (Kafka/RabbitMQ/etc.) plus optional teardown.
func WithBusInstance(bus bus_contract.Bus, teardown func()) Option {
	return func(b *builder) {
		b.bus = bus
		if teardown != nil {
			b.busTds = append(b.busTds, teardown)
		}
	}
}

// WithInMemoryBus wires a contract.Bus using scg-service-bus/memory for unit tests.
func WithInMemoryBus() Option {
	return func(b *builder) {
		bus, closeFn := memory.New()
		b.bus = bus
		if closeFn != nil {
			b.busTds = append(b.busTds, closeFn)
		}
	}
}
