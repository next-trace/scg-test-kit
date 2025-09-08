# scg-test-kit

Contract-first test kit for NextTrace microservices. Provides simple, reusable building blocks for unit, integration, and end-to-end tests while depending only on contracts (scg-config/contract, scg-database/contract). No concrete tech details leak through public APIs.

## Components
- contract: TestKit, DB, GRPC (public API) — Bus via scg-service-bus
- grpc: in-memory gRPC server/client using bufconn + h2c
- testkit: root orchestrator with functional options (no Impl suffix)
- suite: DX facade with TestCase/IntegrationTest

## Usage

Unit-style (TestCase) with in-memory bus:

tk := suite.TestCase(
    t,
    testkit.WithGRPCRegister(func(mux *http.ServeMux) {
        // mux.Handle(myConnectHandler.Pattern(), myConnectHandler)
    }),
    testkit.WithInMemoryBus(),
)
_ = tk.GRPC().ClientConn()
_ = tk.Bus()

Integration-style (IntegrationTest) with ephemeral DB + real bus injected:

tk := suite.IntegrationTest(
    t,
    testkit.WithEphemeralDB(func(t *testing.T) (db_contract.Connection, func(), error) {
        // Use only scg-database/contract to create/migrate/connect
        // return (conn, teardown, nil)
        panic("implement in service")
    }),
    testkit.WithBusInstance(appBus /* bus_contract.Bus */, appBusClose),
    testkit.WithGRPCRegister(func(mux *http.ServeMux) {
        // mux.Handle(myConnectHandler.Pattern(), myConnectHandler)
    }),
)
_ = tk.DB().Conn()
_ = tk.Bus()
_ = tk.GRPC().ClientConn()

Notes:
- Only contracts are exposed publicly. DB/Bus concretions are injected by the app.
- No config/logging ownership in this library.

scg-test-kit

Objective
A small, contract-only test kit for SCG microservices. It exposes only DB and Bus contracts plus an in-memory gRPC client, without owning concrete adapters or config.

Usage examples

Unit-style (no DB, in-memory bus):

    tk := suite.TestCase(t, testkit.WithInMemoryBus())
    // use tk.Bus(), tk.GRPC().ClientConn()

Integration-style (ephemeral DB via factory + real bus):

    factory := func(t *testing.T) (db_contract.Connection, func(), error) {
        // construct ephemeral DB via your service's helpers (using scg-database contracts only)
        // return conn, teardown, nil
        return nil, func(){}, nil
    }
    tk := suite.IntegrationTest(t,
        testkit.WithEphemeralDB(factory),
        testkit.WithBusInstance(realBus, func(){ /* close */ }),
    )

Contracts

- contract.TestKit { T() *testing.T; DB() DB; GRPC() GRPC; Bus() bus_contract.Bus }
- DB { Conn() db_contract.Connection }
- GRPC { ClientConn() *grpc.ClientConn }

Options

- WithGRPCRegister(func(*http.ServeMux))
- WithDBInstance(conn db_contract.Connection, teardown func())
- WithEphemeralDB(factory func(t *testing.T) (db_contract.Connection, func(), error))
- WithBusInstance(bus bus_contract.Bus, teardown func())
- WithInMemoryBus()

Notes
- The kit never creates databases; your service should provide the factory and inject Connection.
- No concrete adapters or config ownership in this repo.
