# scg-test-kit

Contract-first test kit for NextTrace microservices. Provides simple, reusable building blocks for unit, integration, and end-to-end tests while depending only on contracts. No concrete tech details leak through public APIs.

## What this is
This is a test library for microservices. If imported only from _test.go, it will not be compiled into service binaries. The public API exposes only:
- DB via db_contract.Connection
- Bus via bus_contract.Bus
- gRPC via *grpc.ClientConn

## Components
- contract: TestKit, DB, GRPC (public API) — Bus via scg-service-bus
- grpc: in-memory gRPC server/client using bufconn + h2c
- testkit: root orchestrator with functional options (no Impl suffix)
- suite: DX facade with TestCase/IntegrationTest

## Quickstart (TestCase): in-memory bus + gRPC

```go
import (
  "net/http"
  "testing"

  "github.com/next-trace/scg-test-kit/suite"
  "github.com/next-trace/scg-test-kit/testkit"
)

func Test_Health(t *testing.T) {
  tk := suite.TestCase(
    t,
    testkit.WithGRPCRegister(func(mux *http.ServeMux) {
      mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
      })
    }),
    testkit.WithInMemoryBus(),
  )

  if tk.GRPC().ClientConn() == nil {
    t.Fatal("nil gRPC client conn")
  }
}
```

## Integration example (DB + real bus injected)

```go
import (
  "net/http"
  "testing"

  db_contract "github.com/next-trace/scg-database/contract"
  bus_contract "github.com/next-trace/scg-service-bus/contract/bus"

  "github.com/next-trace/scg-test-kit/suite"
  "github.com/next-trace/scg-test-kit/testkit"
)

func Test_App_Integration(t *testing.T) {
  // appBus: constructed by the service using your chosen adapter (Kafka/RabbitMQ/NATS).
  var appBus bus_contract.Bus
  var appBusClose func()

  tk := suite.IntegrationTest(
    t,
    testkit.WithEphemeralDB(func(t *testing.T) (db_contract.Connection, func(), error) {
      // Delegate to a service-side helper that uses ONLY scg-database/contract.
      // This helper must create a fresh DB, run migrations, connect, and return teardown.
      return BuildEphemeralDBFromEnv(t) // you provide this in the service repo
    }),
    testkit.WithBusInstance(appBus, appBusClose),
    testkit.WithGRPCRegister(func(mux *http.ServeMux) {
      // register your connect-go handlers here
    }),
  )

  _ = tk.DB().Conn()
  _ = tk.Bus()
  _ = tk.GRPC().ClientConn()
}
```

## Service-side DB helper (copy into your service repo)
Provide one of these contract-only patterns (pick the variant that matches your scg-database contract surface).

Variant A — If scg-database exposes admin + migrations via contract:
```go

package ephemeraldb

import (
  "os"
  "testing"

  "github.com/google/uuid"
  db_contract "github.com/next-trace/scg-database/contract"
)

func BuildFromEnv(t *testing.T) (db_contract.Connection, func(), error) {
  t.Helper()

  adminDSN := os.Getenv("DATABASE_ADMIN_DSN")
  migDir   := os.Getenv("MIGRATIONS_DIR")
  if adminDSN == "" || migDir == "" {
    t.Fatalf("missing DATABASE_ADMIN_DSN or MIGRATIONS_DIR")
  }

  admin, err := db_contract.ConnectAdmin(adminDSN)
  if err != nil { return nil, nil, err }

  name := "test_" + uuid.NewString()

  if err := admin.CreateDatabase(name); err != nil {
    _ = admin.Close()
    return nil, nil, err
  }

  runner := db_contract.NewMigrationRunner(migDir)
  if err := runner.Up(admin, name); err != nil {
    _ = admin.DropDatabase(name)
    _ = admin.Close()
    return nil, nil, err
  }

  conn, err := db_contract.ConnectDatabase(adminDSN, name)
  if err != nil {
    _ = admin.DropDatabase(name)
    _ = admin.Close()
    return nil, nil, err
  }

  teardown := func() {
    _ = conn.Close()
    _ = admin.DropDatabase(name)
    _ = admin.Close()
  }
  return conn, teardown, nil
}
```

Variant B — If scg-database provides a single contract helper for ephemeral DB:
```go

package ephemeraldb

import (
  "os"
  "testing"

  db_contract "github.com/next-trace/scg-database/contract"
)

func BuildFromEnv(t *testing.T) (db_contract.Connection, func(), error) {
  t.Helper()

  adminDSN := os.Getenv("DATABASE_ADMIN_DSN")
  migDir   := os.Getenv("MIGRATIONS_DIR")
  if adminDSN == "" || migDir == "" {
    t.Fatalf("missing DATABASE_ADMIN_DSN or MIGRATIONS_DIR")
  }

  // Example: a higher-level contract helper, if available in scg-database:
  // conn, teardown, err := db_contract.NewEphemeral(adminDSN, migDir)
  // return conn, teardown, err

  t.Fatalf("wire your scg-database contract helper here")
  return nil, nil, nil
}
```

In your tests, reference it as:
```go
return ephemeraldb.BuildFromEnv(t)
```

## Will this ship in prod?
No — as long as you import scg-test-kit only from _test.go, it’s not compiled into your service binary.
