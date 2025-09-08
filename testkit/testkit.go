package testkit

import (
	"testing"

	db_contract "github.com/next-trace/scg-database/contract"
	bus_contract "github.com/next-trace/scg-service-bus/contract/bus"
	"github.com/next-trace/scg-test-kit/contract"
	"github.com/next-trace/scg-test-kit/db"
	"github.com/next-trace/scg-test-kit/grpc"
)

type testKit struct {
	t   *testing.T
	db  contract.DB
	rpc contract.GRPC
	bs  bus_contract.Bus
}

func (k *testKit) T() *testing.T         { return k.t }
func (k *testKit) DB() contract.DB       { return k.db }
func (k *testKit) GRPC() contract.GRPC   { return k.rpc }
func (k *testKit) Bus() bus_contract.Bus { return k.bs }

// NewTestCase builds a minimal kit; compose infra via options.
func NewTestCase(t *testing.T, opts ...Option) contract.TestKit {
	return build(t, opts...)
}

// NewIntegrationTest is identical to NewTestCase but conveys intent.
func NewIntegrationTest(t *testing.T, opts ...Option) contract.TestKit {
	return build(t, opts...)
}

func build(t *testing.T, opts ...Option) contract.TestKit {
	t.Helper()
	b := &builder{}
	for _, opt := range opts {
		opt(b)
	}

	// 1) DB: run factory if provided; otherwise use injected instance
	var dbc db_contract.Connection
	var dbTd func()
	var err error

	if b.dbFactory != nil {
		dbc, dbTd, err = b.dbFactory(t)
		if err != nil {
			t.Fatalf("ephemeral DB factory failed: %v", err)
		}
	} else {
		dbc, dbTd = b.dbConn, b.dbTd
	}

	var cdb contract.DB
	if dbc != nil {
		cdb = db.NewFromConn(dbc) // thin wrapper to contract.DB
	}
	if dbTd != nil {
		t.Cleanup(dbTd)
	}

	// 2) gRPC in-memory server
	var rpc contract.GRPC
	{
		r, td := grpc.New(t, b.register)
		rpc = r
		if td != nil {
			t.Cleanup(td)
		}
	}

	// 3) Bus teardowns (bus may be nil for pure unit tests)
	for _, f := range b.busTds {
		if f != nil {
			t.Cleanup(f)
		}
	}

	return &testKit{
		t:   t,
		db:  cdb,
		rpc: rpc,
		bs:  b.bus,
	}
}
