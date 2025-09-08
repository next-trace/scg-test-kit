package testkit_test

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	db_contract "github.com/next-trace/scg-database/contract"
	"github.com/next-trace/scg-test-kit/suite"
	"github.com/next-trace/scg-test-kit/testkit"
)

// fakeConn implements db_contract.Connection minimally for tests.
// All methods are no-ops adequate for verifying wiring.
type fakeConn struct{}

func (f fakeConn) GetConnection() any                                              { return nil }
func (f fakeConn) Ping(context.Context) error                                      { return nil }
func (f fakeConn) Close() error                                                    { return nil }
func (f fakeConn) NewRepository(db_contract.Model) (db_contract.Repository, error) { return nil, nil }
func (f fakeConn) Transaction(_ context.Context, fn func(db_contract.Connection) error) error {
	if fn != nil {
		return fn(f)
	}
	return nil
}
func (f fakeConn) Select(context.Context, string, ...any) ([]map[string]any, error) { return nil, nil }
func (f fakeConn) Statement(context.Context, string, ...any) (sql.Result, error)    { return nil, nil }

func TestOptions_WithInMemoryBus_ProvidesBus(t *testing.T) {
	tk := suite.TestCase(t, testkit.WithInMemoryBus())
	if tk.Bus() == nil {
		t.Fatalf("expected non-nil bus from WithInMemoryBus")
	}
}

func TestOptions_WithDBInstance_WiresConnection(t *testing.T) {
	conn := fakeConn{}
	tk := suite.TestCase(t, testkit.WithDBInstance(conn, nil))
	if tk.DB() == nil {
		t.Fatalf("expected DB not nil")
	}
	if tk.DB().Conn() == nil { // returns interface value backed by fakeConn
		t.Fatalf("expected DB.Conn not nil")
	}
}

func TestOptions_WithEphemeralDB_Factory_WiresConnection(t *testing.T) {
	factory := func(_ *testing.T) (db_contract.Connection, func(), error) {
		called := true
		_ = called // silence linters; we only need to return a teardown
		return fakeConn{}, func() {}, nil
	}
	// Build kit using the ephemeral factory.
	tk := suite.TestCase(t, testkit.WithEphemeralDB(factory))
	if tk.DB() == nil {
		t.Fatalf("expected DB not nil from ephemeral factory")
	}
	if tk.DB().Conn() == nil {
		t.Fatalf("expected DB.Conn not nil from ephemeral factory")
	}
}

func TestOptions_CleanupTeardownsCalled(t *testing.T) {
	// WithDBInstance teardown is called at end of test scope
	var dbClosed bool
	t.Run("db instance teardown", func(t *testing.T) {
		_ = suite.TestCase(t, testkit.WithDBInstance(nil, func() { dbClosed = true }))
	})
	if !dbClosed {
		t.Fatalf("expected WithDBInstance teardown to be called after subtest")
	}

	// WithEphemeralDB factory and teardown are called
	var factoryCalled, ephClosed bool
	t.Run("ephemeral factory teardown", func(t *testing.T) {
		factory := func(_ *testing.T) (_ db_contract.Connection, td func(), _ error) {
			factoryCalled = true
			return db_contract.Connection(nil), func() { ephClosed = true }, nil
		}
		_ = suite.TestCase(t, testkit.WithEphemeralDB(factory))
	})
	if !factoryCalled || !ephClosed {
		t.Fatalf("expected ephemeral factory called=%v and teardown called=%v", factoryCalled, ephClosed)
	}

	// WithBusInstance teardown is called
	var busClosed bool
	t.Run("bus instance teardown", func(t *testing.T) {
		_ = suite.TestCase(t, testkit.WithBusInstance(nil, func() { busClosed = true }))
	})
	if !busClosed {
		t.Fatalf("expected WithBusInstance teardown to be called after subtest")
	}
}

func TestWithGRPCRegister_FunctionIsInvoked(t *testing.T) {
	var called bool
	reg := func(mux *http.ServeMux) {
		called = true
		// also attach a trivial handler
		mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	}
	k := suite.TestCase(t, testkit.WithGRPCRegister(reg))
	if k.GRPC().ClientConn() == nil {
		t.Fatalf("expected non-nil grpc client")
	}
	if !called {
		t.Fatalf("expected register function to be invoked during build")
	}
}
