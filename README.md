# scg-test-kit

`scg-test-kit` is the technology-agnostic testing toolkit for SCG Go services.

## 1. What This Library Is and Is NOT
- **Is NOT a provisioner**: This library does NOT start Docker containers (e.g., testcontainers), local processes, or cloud resources.
- **Is NOT technology-specific**: It contains no code for PostgreSQL, Kafka, Redis, etc.
- **IS an orchestrator**: It provides a unified `Harness` to manage the lifecycle of resources that your service provisions.
- **IS a cleanup manager**: It ensures all registered resources are cleaned up in LIFO (Last-In-First-Out) order using `testing.TB.Cleanup`.

## 2. Core Principles
- **Injection-Only**: Services provision their own dependencies and inject them into the `Harness`.
- **Lifecycle Management**: The `Harness` orchestrates cleanups, keeping test code clean.
- **Technology-Agnostic**: Core packages remain generic and stable.

## 3. Responsibility Boundary
- **Services**: Responsible for provisioning (e.g., starting a database container) and defining how to shut it down.
- **test-kit**: Responsible for holding the reference and ensuring the shutdown/cleanup is called at the end of the test.

## 4. Installation
```bash
go get github.com/next-trace/scg-test-kit
```

## 5. Quickstart

### Unit Test (No External Resources)
```go
func Test_Unit(t *testing.T) {
    // 1. Create a harness
    // You can also use testkit.NewUnitHarness(t) for semantic clarity
    h := testkit.New(t)

    // 2. Use it for shared helpers or cleanup
    h.RegisterCleanup(func() {
        // cleanup logic
    })
}
```

### Integration Test (With External Resources)
```go
func Test_Integration(t *testing.T) {
    // 1. Service provisions its own dependency
    client := MyClient{Addr: "localhost:1234"}
    cleanup := func() error { return client.Close() }

    // 2. Inject into Harness
    // You can also use testkit.NewIntegrationHarness(t) for semantic clarity
    h := testkit.New(t,
        testkit.WithResource("myClient", client, cleanup),
    )
    
    // 3. Retrieve when needed
    val, ok := testkit.Resource[MyClient](h, "myClient")
    // ...
}
```

## 6. MIGRATION GUIDE (v0.2.2+)

If you are upgrading from an older version, please follow these steps:

1. **Automatic Cleanup**: You no longer need to call `h.Cleanup()` manually. It is automatically registered with `t.Cleanup()`.
2. **Unified Entrypoint**: `testkit.New(t)` is the primary constructor, but `NewUnitHarness` and `NewIntegrationHarness` are preserved as semantic aliases.
3. **Browser Harness**: `testkit.NewBrowserHarness` remains available but now uses the unified `New` internally.

## 7. Running Tests
To run all tests in the repository:
```bash
go test ./...
```
Or use the SCG helper:
```bash
./scg doctor:all
```

## 8. Standards & Decisions
This library follows the [SCG Library Standards](Docs/LIBRARY_REPO_STRUCTURE.md).

- **Documentation**: [Docs/](Docs/)
- **Public API**: [Docs/public_api.md](Docs/public_api.md)
- **ADR (Architectural Decisions)**: [Docs/adr/](Docs/adr/)
- **Changelog**: [CHANGELOG.md](CHANGELOG.md)
- **License**: [LICENSE](LICENSE)
