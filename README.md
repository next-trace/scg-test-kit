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

### Generic Resource Injection with WithResource
```go
func Test_CustomResource(t *testing.T) {
    // 1. Service provisions its own dependency
    client := MyClient{Addr: "localhost:1234"}
    cleanup := func() error { return client.Close() }

    // 2. Inject into Harness
    h := testkit.NewHarness(t,
        testkit.WithResource("myClient", client, cleanup),
    )
    
    // 3. Retrieve when needed
    val, ok := testkit.Resource[MyClient](h, "myClient")
    // ...
}
```

## 6. Running Tests
To run all tests in the repository:
```bash
go test ./...
```
Or use the SCG helper:
```bash
./scg doctor:all
```

## 7. Standards & Decisions
This library follows the [SCG Library Standards](Docs/LIBRARY_REPO_STRUCTURE.md).

- **Documentation**: [Docs/](Docs/)
- **Public API**: [Docs/public_api.md](Docs/public_api.md)
- **ADR (Architectural Decisions)**: [Docs/adr/](Docs/adr/)
- **Changelog**: [CHANGELOG.md](CHANGELOG.md)
- **License**: [LICENSE](LICENSE)
